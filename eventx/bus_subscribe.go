package eventx

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/samber/lo"
)

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

	if b.closed {
		return nil, ErrBusClosed
	}

	b.nextID++
	id := b.nextID

	b.subsByType.Put(eventType, id, &subscription{
		id:      id,
		handler: finalHandler,
	})

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

	handlers := make([]HandlerFunc, 0, len(row))
	for _, sub := range row {
		if sub != nil && sub.handler != nil {
			handlers = append(handlers, sub.handler)
		}
	}
	return handlers
}
