package eventx

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/DaiYuANg/arcgo/observability"
)

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

	if b.closed {
		return ErrBusClosed
	}
	handlers := b.snapshotHandlersByEventType(reflect.TypeOf(event))

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

	obs := b.observabilitySafe()
	start := time.Now()
	ctx, span := obs.StartSpan(ctx, "eventx.publish.async.enqueue",
		observability.String("event_name", eventName(event)),
	)
	defer span.End()

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
	handlers := b.snapshotHandlersByEventType(reflect.TypeOf(event))

	// Use ants pool if enabled
	if b.antsPool != nil {
		task := publishTask{
			ctx:      ctx,
			event:    event,
			handlers: handlers,
		}
		err := b.antsPool.Submit(func() {
			b.executeTask(task)
		})
		if err != nil {
			span.RecordError(err)
			obs.AddCounter(ctx, metricAsyncEnqueueTotal, 1,
				observability.String("result", "pool_error"),
				observability.String("event_name", eventName(event)),
			)
			obs.RecordHistogram(ctx, metricAsyncEnqueueDurationMS, float64(time.Since(start).Milliseconds()),
				observability.String("result", "pool_error"),
				observability.String("event_name", eventName(event)),
			)
			return fmt.Errorf("failed to submit task to ants pool: %w", err)
		}

		obs.AddCounter(ctx, metricAsyncEnqueueTotal, 1,
			observability.String("result", "submitted"),
			observability.String("event_name", eventName(event)),
		)
		obs.RecordHistogram(ctx, metricAsyncEnqueueDurationMS, float64(time.Since(start).Milliseconds()),
			observability.String("result", "submitted"),
			observability.String("event_name", eventName(event)),
		)
		return nil
	}

	// Fallback to legacy channel-based queue
	if b.asyncQueue == nil {
		return b.Publish(ctx, event)
	}

	select {
	case b.asyncQueue <- publishTask{ctx: ctx, event: event, handlers: handlers}:
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

func (b *Bus) executeTask(task publishTask) {
	err := b.dispatch(task.ctx, task.event, task.handlers, "async")
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
}

func (b *Bus) workerLoop() {
	defer b.workerWG.Done()
	for task := range b.asyncQueue {
		b.executeTask(task)
	}
}
