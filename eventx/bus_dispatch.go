package eventx

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/DaiYuANg/arcgo/observability"
	"github.com/samber/lo"
)

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
