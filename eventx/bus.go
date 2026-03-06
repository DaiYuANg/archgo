package eventx

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
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
	mu          sync.RWMutex
	closed      bool
	nextID      uint64
	subsByType  map[reflect.Type]map[uint64]*subscription
	parallel    bool
	middleware  []Middleware
	onAsyncErr  asyncErrorHandler
	asyncQueue  chan publishTask
	workerWG    sync.WaitGroup
	queueTaskWG sync.WaitGroup
	dispatchWG  sync.WaitGroup
}

// New creates a new Bus.
func New(opts ...Option) *Bus {
	cfg := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	b := &Bus{
		subsByType: make(map[reflect.Type]map[uint64]*subscription),
		parallel:   cfg.parallel,
		middleware: cfg.middleware,
		onAsyncErr: cfg.onAsyncError,
	}

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
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

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

	return b.dispatch(ctx, event, handlers)
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

	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		return ErrBusClosed
	}

	b.queueTaskWG.Add(1)
	select {
	case b.asyncQueue <- publishTask{ctx: ctx, event: event}:
		return nil
	default:
		b.queueTaskWG.Done()
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

	total := 0
	for _, byID := range b.subsByType {
		total += len(byID)
	}
	return total
}

func (b *Bus) workerLoop() {
	defer b.workerWG.Done()
	for task := range b.asyncQueue {
		b.mu.RLock()
		handlers := b.snapshotHandlersByEventTypeLocked(reflect.TypeOf(task.event))
		b.mu.RUnlock()

		err := b.dispatch(task.ctx, task.event, handlers)
		if err != nil && b.onAsyncErr != nil {
			b.onAsyncErr(task.ctx, task.event, err)
		}
		b.queueTaskWG.Done()
	}
}

func (b *Bus) snapshotHandlersByEventTypeLocked(eventType reflect.Type) []HandlerFunc {
	byID := b.subsByType[eventType]
	if len(byID) == 0 {
		return nil
	}

	out := make([]HandlerFunc, 0, len(byID))
	for _, sub := range byID {
		out = append(out, sub.handler)
	}
	return out
}

func (b *Bus) dispatch(ctx context.Context, event Event, handlers []HandlerFunc) error {
	if len(handlers) == 0 {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	b.dispatchWG.Add(1)
	defer b.dispatchWG.Done()

	if b.parallel {
		return b.dispatchParallel(ctx, event, handlers)
	}

	return b.dispatchSerial(ctx, event, handlers)
}

func (b *Bus) dispatchSerial(ctx context.Context, event Event, handlers []HandlerFunc) error {
	var errs []error
	for _, handler := range handlers {
		if handler == nil {
			continue
		}
		if err := handler(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (b *Bus) dispatchParallel(ctx context.Context, event Event, handlers []HandlerFunc) error {
	errCh := make(chan error, len(handlers))
	var wg sync.WaitGroup

	for _, handler := range handlers {
		if handler == nil {
			continue
		}
		wg.Add(1)
		go func(h HandlerFunc) {
			defer wg.Done()
			if err := h(ctx, event); err != nil {
				errCh <- err
			}
		}(handler)
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
