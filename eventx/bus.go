package eventx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/DaiYuANg/arcgo/observability"
	"github.com/samber/lo"
)

// Event is the common event contract for strongly typed publish/subscribe.
type Event interface {
	Name() string
}

type subscription struct {
	id      uint64
	handler HandlerFunc
}

type publishTask struct {
	ctx   context.Context
	event Event
}

// Bus is an in-memory strongly typed event bus.
type Bus struct {
	mu            sync.RWMutex
	closed        bool
	nextID        uint64
	subsByType    map[reflect.Type]map[uint64]*subscription
	parallel      bool
	middleware    []Middleware
	onAsyncErr    asyncErrorHandler
	asyncQueue    chan publishTask
	workerWG      sync.WaitGroup
	queueTaskWG   sync.WaitGroup
	dispatchWG    sync.WaitGroup
	observability observability.Observability
	logger        *slog.Logger
}

const (
	metricDispatchTotal           = "eventx_dispatch_total"
	metricDispatchDurationMS      = "eventx_dispatch_duration_ms"
	metricAsyncEnqueueTotal       = "eventx_async_enqueue_total"
	metricAsyncEnqueueDurationMS  = "eventx_async_enqueue_duration_ms"
	metricAsyncDispatchErrorTotal = "eventx_async_dispatch_error_total"
)

// New creates a new Bus.
func New(opts ...Option) *Bus {
	cfg := defaultOptions()
	lo.ForEach(opts, func(opt Option, _ int) {
		if opt != nil {
			opt(&cfg)
		}
	})

	b := &Bus{
		subsByType:    make(map[reflect.Type]map[uint64]*subscription),
		parallel:      cfg.parallel,
		middleware:    cfg.middleware,
		onAsyncErr:    cfg.onAsyncError,
		observability: observability.Normalize(cfg.observability, nil),
	}
	b.logger = b.observability.Logger().With("component", "eventx.bus")

	if cfg.asyncWorkers > 0 && cfg.asyncQueueSize > 0 {
		b.asyncQueue = make(chan publishTask, cfg.asyncQueueSize)
		for i := 0; i < cfg.asyncWorkers; i++ {
			b.workerWG.Add(1)
			go b.workerLoop()
		}
	}

	return b
}

// Subscribe registers a strongly typed handler and returns an unsubscribe function.
func Subscribe[T Event](b *Bus, handler func(context.Context, T) error, opts ...SubscribeOption) (func(), error) {
	if b == nil {
		return nil, ErrNilBus
	}
	if handler == nil {
		return nil, ErrNilHandler
	}

	cfg := defaultSubscribeOptions()
	lo.ForEach(opts, func(opt SubscribeOption, _ int) {
		if opt != nil {
			opt(&cfg)
		}
	})

	eventType := reflect.TypeFor[T]()
	base := func(ctx context.Context, event Event) error {
		typed, ok := any(event).(T)
		if !ok {
			return fmt.Errorf("eventx: event type mismatch, expect %v, got %T", eventType, event)
		}
		return handler(ctx, typed)
	}

	// Global middleware wraps subscription middleware.
	finalHandler := chain(chain(base, cfg.middleware), b.middleware)

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil, ErrBusClosed
	}

	b.nextID++
	id := b.nextID
	byID := b.subsByType[eventType]
	if byID == nil {
		byID = make(map[uint64]*subscription)
		b.subsByType[eventType] = byID
	}
	byID[id] = &subscription{
		id:      id,
		handler: finalHandler,
	}

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			inner := b.subsByType[eventType]
			if inner == nil {
				return
			}
			delete(inner, id)
			if len(inner) == 0 {
				delete(b.subsByType, eventType)
			}
		})
	}
	return unsubscribe, nil
}

// Publish dispatches one event synchronously to all matching subscribers.
func (b *Bus) Publish(ctx context.Context, event Event) error {
	if b == nil {
		return ErrNilBus
	}
	if event == nil {
		return ErrNilEvent
	}
	if ctx == nil {
		ctx = context.Background()
	}

	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return ErrBusClosed
	}
	handlers := b.snapshotHandlersByEventTypeLocked(reflect.TypeOf(event))
	b.mu.RUnlock()

	return b.dispatch(ctx, event, handlers, "sync")
}

// PublishAsync enqueues one event for asynchronous dispatch.
func (b *Bus) PublishAsync(ctx context.Context, event Event) error {
	if b == nil {
		return ErrNilBus
	}
	if event == nil {
		return ErrNilEvent
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if b.asyncQueue == nil {
		// Keep behavior predictable: fallback to sync when async is disabled.
		return b.Publish(ctx, event)
	}

	obs := b.observabilitySafe()
	start := time.Now()
	ctx, span := obs.StartSpan(ctx, "eventx.publish.async.enqueue",
		observability.String("event_name", eventName(event)),
	)
	defer span.End()

	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		err := ErrBusClosed
		span.RecordError(err)
		obs.AddCounter(ctx, metricAsyncEnqueueTotal, 1,
			observability.String("result", "closed"),
			observability.String("event_name", eventName(event)),
		)
		obs.RecordHistogram(ctx, metricAsyncEnqueueDurationMS, float64(time.Since(start).Milliseconds()),
			observability.String("result", "closed"),
			observability.String("event_name", eventName(event)),
		)
		return ErrBusClosed
	}

	b.queueTaskWG.Add(1)
	select {
	case b.asyncQueue <- publishTask{ctx: ctx, event: event}:
		obs.AddCounter(ctx, metricAsyncEnqueueTotal, 1,
			observability.String("result", "enqueued"),
			observability.String("event_name", eventName(event)),
		)
		obs.RecordHistogram(ctx, metricAsyncEnqueueDurationMS, float64(time.Since(start).Milliseconds()),
			observability.String("result", "enqueued"),
			observability.String("event_name", eventName(event)),
		)
		return nil
	default:
		b.queueTaskWG.Done()
		span.RecordError(ErrAsyncQueueFull)
		obs.AddCounter(ctx, metricAsyncEnqueueTotal, 1,
			observability.String("result", "queue_full"),
			observability.String("event_name", eventName(event)),
		)
		obs.RecordHistogram(ctx, metricAsyncEnqueueDurationMS, float64(time.Since(start).Milliseconds()),
			observability.String("result", "queue_full"),
			observability.String("event_name", eventName(event)),
		)
		return ErrAsyncQueueFull
	}
}

// Close stops accepting new events, drains async queue, and waits in-flight handlers.
func (b *Bus) Close() error {
	if b == nil {
		return nil
	}

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	queue := b.asyncQueue
	if queue != nil {
		close(queue)
	}
	b.mu.Unlock()

	if queue != nil {
		b.workerWG.Wait()
		b.queueTaskWG.Wait()
	}
	b.dispatchWG.Wait()
	return nil
}

// SubscriberCount returns active subscriber count.
func (b *Bus) SubscriberCount() int {
	if b == nil {
		return 0
	}
	b.mu.RLock()
	defer b.mu.RUnlock()

	return lo.SumBy(lo.Values(b.subsByType), func(byID map[uint64]*subscription) int {
		return len(byID)
	})
}

func (b *Bus) workerLoop() {
	defer b.workerWG.Done()
	for task := range b.asyncQueue {
		b.mu.RLock()
		handlers := b.snapshotHandlersByEventTypeLocked(reflect.TypeOf(task.event))
		b.mu.RUnlock()

		err := b.dispatch(task.ctx, task.event, handlers, "async")
		if err != nil && b.onAsyncErr != nil {
			b.onAsyncErr(task.ctx, task.event, err)
		} else if err != nil {
			b.logger.Warn("async dispatch failed",
				"event_name", eventName(task.event),
				"error", err.Error(),
			)
		}
		if err != nil {
			b.observabilitySafe().AddCounter(task.ctx, metricAsyncDispatchErrorTotal, 1,
				observability.String("event_name", eventName(task.event)),
			)
		}
		b.queueTaskWG.Done()
	}
}

func (b *Bus) snapshotHandlersByEventTypeLocked(eventType reflect.Type) []HandlerFunc {
	byID := b.subsByType[eventType]
	if len(byID) == 0 {
		return nil
	}

	return lo.FilterMap(lo.Values(byID), func(sub *subscription, _ int) (HandlerFunc, bool) {
		if sub == nil || sub.handler == nil {
			return nil, false
		}
		return sub.handler, true
	})
}

func (b *Bus) dispatch(ctx context.Context, event Event, handlers []HandlerFunc, mode string) error {
	if len(handlers) == 0 {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	obs := b.observabilitySafe()
	start := time.Now()
	ctx, span := obs.StartSpan(ctx, "eventx.dispatch",
		observability.String("mode", mode),
		observability.String("event_name", eventName(event)),
		observability.Int64("handlers", int64(len(handlers))),
	)
	defer span.End()

	result := "success"
	defer func() {
		obs.AddCounter(ctx, metricDispatchTotal, 1,
			observability.String("mode", mode),
			observability.String("result", result),
			observability.String("event_name", eventName(event)),
		)
		obs.RecordHistogram(ctx, metricDispatchDurationMS, float64(time.Since(start).Milliseconds()),
			observability.String("mode", mode),
			observability.String("result", result),
			observability.String("event_name", eventName(event)),
		)
	}()

	b.dispatchWG.Add(1)
	defer b.dispatchWG.Done()

	var err error
	if b.parallel {
		err = b.dispatchParallel(ctx, event, handlers)
	} else {
		err = b.dispatchSerial(ctx, event, handlers)
	}

	if err != nil {
		result = "error"
		span.RecordError(err)
	}
	return err
}

func (b *Bus) dispatchSerial(ctx context.Context, event Event, handlers []HandlerFunc) error {
	errs := lo.FilterMap(handlers, func(handler HandlerFunc, _ int) (error, bool) {
		if handler == nil {
			return nil, false
		}
		err := handler(ctx, event)
		return err, err != nil
	})
	return errors.Join(errs...)
}

func (b *Bus) dispatchParallel(ctx context.Context, event Event, handlers []HandlerFunc) error {
	errCh := make(chan error, len(handlers))
	var wg sync.WaitGroup

	lo.ForEach(handlers, func(handler HandlerFunc, _ int) {
		if handler == nil {
			return
		}
		wg.Add(1)
		go func(h HandlerFunc) {
			defer wg.Done()
			if err := h(ctx, event); err != nil {
				errCh <- err
			}
		}(handler)
	})

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (b *Bus) observabilitySafe() observability.Observability {
	if b == nil {
		return observability.Nop()
	}
	return observability.Normalize(b.observability, b.logger)
}

func eventName(event Event) string {
	if event == nil {
		return ""
	}

	name := strings.TrimSpace(event.Name())
	if name != "" {
		return name
	}
	return reflect.TypeOf(event).String()
}
