package eventx

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/samber/lo"
)

// Subscribe registers a strongly typed handler and returns an unsubscribe function.
func Subscribe[T Event](b BusRuntime, handler func(context.Context, T) error, opts ...SubscribeOption) (func(), error) {
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

	return b.subscribe(eventType, base, cfg.middleware, 0)
}

// SubscribeOnce registers a strongly typed handler that will auto-unsubscribe
// after handling one event.
func SubscribeOnce[T Event](b BusRuntime, handler func(context.Context, T) error, opts ...SubscribeOption) (func(), error) {
	return SubscribeN(b, 1, handler, opts...)
}

// SubscribeN registers a strongly typed handler that will auto-unsubscribe
// after handling n events.
func SubscribeN[T Event](b BusRuntime, n int, handler func(context.Context, T) error, opts ...SubscribeOption) (func(), error) {
	if n <= 0 {
		return nil, ErrInvalidSubscribeCount
	}
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

	return b.subscribe(eventType, base, cfg.middleware, n)
}

func (b *Bus) subscribe(eventType reflect.Type, base HandlerFunc, subscriberMiddleware []Middleware, maxCalls int) (func(), error) {
	if b == nil {
		return nil, ErrNilBus
	}

	// Global middleware wraps subscription middleware.
	finalHandler := chain(chain(base, subscriberMiddleware), b.middleware)

	id, err := b.registerSubscription(eventType, func(id uint64) HandlerFunc {
		if maxCalls <= 0 {
			return finalHandler
		}

		wrapped := finalHandler
		remaining := int64(maxCalls)
		return func(ctx context.Context, event Event) error {
			err := wrapped(ctx, event)
			if atomic.AddInt64(&remaining, -1) <= 0 {
				b.subsByType.Delete(eventType, id)
			}
			return err
		}
	})
	if err != nil {
		return nil, err
	}

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			b.subsByType.Delete(eventType, id)
		})
	}
	return unsubscribe, nil
}

func (b *Bus) snapshotHandlersByEventType(eventType reflect.Type) []HandlerFunc {
	row := b.subsByType.Row(eventType)
	if len(row) == 0 {
		return nil
	}

	return lo.FilterMap(lo.Values(row), func(sub *subscription, _ int) (HandlerFunc, bool) {
		if sub == nil || sub.handler == nil {
			return nil, false
		}
		return sub.handler, true
	})
}
