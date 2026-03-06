package eventx

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type userCreated struct {
	ID int
}

func nilContext() context.Context {
	return nil
}

func (e userCreated) Name() string {
	return "user.created"
}

func TestPublishSync(t *testing.T) {
	t.Parallel()

	bus := New()
	defer func() { _ = bus.Close() }()

	var got int64
	unsubscribe, err := Subscribe(bus, func(ctx context.Context, evt userCreated) error {
		atomic.AddInt64(&got, int64(evt.ID))
		return nil
	})
	require.NoError(t, err)
	defer unsubscribe()

	err = bus.Publish(context.Background(), userCreated{ID: 7})
	require.NoError(t, err)
	require.EqualValues(t, 7, atomic.LoadInt64(&got))
}

func TestPublishNilEvent(t *testing.T) {
	t.Parallel()

	bus := New()
	defer func() { _ = bus.Close() }()

	var evt Event
	err := bus.Publish(context.Background(), evt)
	require.ErrorIs(t, err, ErrNilEvent)
}

func TestSubscribeNilHandler(t *testing.T) {
	t.Parallel()

	bus := New()
	defer func() { _ = bus.Close() }()

	_, err := Subscribe[userCreated](bus, nil)
	require.ErrorIs(t, err, ErrNilHandler)
}

func TestNilBus(t *testing.T) {
	t.Parallel()

	var nilBus *Bus

	_, err := Subscribe(nilBus, func(ctx context.Context, evt userCreated) error { return nil })
	require.ErrorIs(t, err, ErrNilBus)

	err = nilBus.Publish(context.Background(), userCreated{ID: 1})
	require.ErrorIs(t, err, ErrNilBus)

	err = nilBus.PublishAsync(context.Background(), userCreated{ID: 1})
	require.ErrorIs(t, err, ErrNilBus)
}

func TestPublishNilContext(t *testing.T) {
	t.Parallel()

	bus := New()
	defer func() { _ = bus.Close() }()

	_, err := Subscribe(bus, func(ctx context.Context, evt userCreated) error {
		if ctx == nil {
			return errors.New("nil context")
		}
		return nil
	})
	require.NoError(t, err)

	err = bus.Publish(nilContext(), userCreated{ID: 1})
	require.NoError(t, err)
}

func TestPublishAsyncNilContext(t *testing.T) {
	t.Parallel()

	bus := New(
		WithAsyncWorkers(1),
		WithAsyncQueueSize(8),
	)

	nilCtx := make(chan bool, 1)
	_, err := Subscribe(bus, func(ctx context.Context, evt userCreated) error {
		nilCtx <- (ctx == nil)
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, bus.PublishAsync(nilContext(), userCreated{ID: 1}))
	require.NoError(t, bus.Close())

	select {
	case gotNil := <-nilCtx:
		require.False(t, gotNil)
	case <-time.After(time.Second):
		t.Fatal("async handler did not run in time")
	}
}

func TestUnsubscribe(t *testing.T) {
	t.Parallel()

	bus := New()
	defer func() { _ = bus.Close() }()

	var count int64
	unsubscribe, err := Subscribe(bus, func(ctx context.Context, evt userCreated) error {
		atomic.AddInt64(&count, 1)
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, bus.Publish(context.Background(), userCreated{ID: 1}))
	unsubscribe()
	require.NoError(t, bus.Publish(context.Background(), userCreated{ID: 1}))

	require.EqualValues(t, 1, atomic.LoadInt64(&count))
}

func TestPublishAsyncAndCloseDrain(t *testing.T) {
	t.Parallel()

	bus := New(
		WithAsyncWorkers(2),
		WithAsyncQueueSize(16),
	)

	var count int64
	_, err := Subscribe(bus, func(ctx context.Context, evt userCreated) error {
		atomic.AddInt64(&count, 1)
		return nil
	})
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		require.NoError(t, bus.PublishAsync(context.Background(), userCreated{ID: i}))
	}

	require.NoError(t, bus.Close())
	require.EqualValues(t, 10, atomic.LoadInt64(&count))
}

func TestPublishAsyncFallbackToSync(t *testing.T) {
	t.Parallel()

	bus := New(
		WithAsyncWorkers(0),
		WithAsyncQueueSize(0),
	)
	defer func() { _ = bus.Close() }()

	var count int64
	_, err := Subscribe(bus, func(ctx context.Context, evt userCreated) error {
		atomic.AddInt64(&count, 1)
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, bus.PublishAsync(context.Background(), userCreated{ID: 1}))
	require.EqualValues(t, 1, atomic.LoadInt64(&count))
}

func TestPublishAsyncQueueFull(t *testing.T) {
	t.Parallel()

	bus := New(
		WithAsyncWorkers(1),
		WithAsyncQueueSize(1),
	)
	defer func() { _ = bus.Close() }()

	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once

	_, err := Subscribe(bus, func(ctx context.Context, evt userCreated) error {
		once.Do(func() {
			close(started)
		})
		<-release
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, bus.PublishAsync(context.Background(), userCreated{ID: 1}))
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler did not start in time")
	}

	require.NoError(t, bus.PublishAsync(context.Background(), userCreated{ID: 2}))
	err = bus.PublishAsync(context.Background(), userCreated{ID: 3})
	require.ErrorIs(t, err, ErrAsyncQueueFull)

	close(release)
}

func TestMiddlewareOrder(t *testing.T) {
	t.Parallel()

	order := make([]string, 0, 5)

	bus := New(
		WithMiddleware(func(next HandlerFunc) HandlerFunc {
			return func(ctx context.Context, event Event) error {
				order = append(order, "global-before")
				err := next(ctx, event)
				order = append(order, "global-after")
				return err
			}
		}),
	)
	defer func() { _ = bus.Close() }()

	_, err := Subscribe(bus,
		func(ctx context.Context, evt userCreated) error {
			order = append(order, "handler")
			return nil
		},
		WithSubscriberMiddleware(func(next HandlerFunc) HandlerFunc {
			return func(ctx context.Context, event Event) error {
				order = append(order, "subscriber-before")
				err := next(ctx, event)
				order = append(order, "subscriber-after")
				return err
			}
		}),
	)
	require.NoError(t, err)

	require.NoError(t, bus.Publish(context.Background(), userCreated{ID: 1}))
	require.Equal(t, []string{
		"global-before",
		"subscriber-before",
		"handler",
		"subscriber-after",
		"global-after",
	}, order)
}

func TestRecoverMiddleware(t *testing.T) {
	t.Parallel()

	bus := New(
		WithMiddleware(RecoverMiddleware()),
	)
	defer func() { _ = bus.Close() }()

	_, err := Subscribe(bus, func(ctx context.Context, evt userCreated) error {
		panic("boom")
	})
	require.NoError(t, err)

	err = bus.Publish(context.Background(), userCreated{ID: 1})
	require.Error(t, err)
	require.Contains(t, err.Error(), "recovered panic")
}

func TestAsyncErrorHandler(t *testing.T) {
	t.Parallel()

	var got int64
	bus := New(
		WithAsyncWorkers(1),
		WithAsyncQueueSize(8),
		WithAsyncErrorHandler(func(ctx context.Context, event Event, err error) {
			if err != nil {
				atomic.AddInt64(&got, 1)
			}
		}),
	)

	_, err := Subscribe(bus, func(ctx context.Context, evt userCreated) error {
		return errors.New("boom")
	})
	require.NoError(t, err)

	require.NoError(t, bus.PublishAsync(context.Background(), userCreated{ID: 1}))
	require.NoError(t, bus.Close())
	require.EqualValues(t, 1, atomic.LoadInt64(&got))
}

func TestCloseRejectsNewRequests(t *testing.T) {
	t.Parallel()

	bus := New(
		WithAsyncWorkers(1),
		WithAsyncQueueSize(8),
	)
	require.NoError(t, bus.Close())

	err := bus.Publish(context.Background(), userCreated{ID: 1})
	require.ErrorIs(t, err, ErrBusClosed)

	err = bus.PublishAsync(context.Background(), userCreated{ID: 1})
	require.ErrorIs(t, err, ErrBusClosed)

	_, err = Subscribe(bus, func(ctx context.Context, evt userCreated) error {
		return nil
	})
	require.ErrorIs(t, err, ErrBusClosed)
}

func TestSubscriberCount(t *testing.T) {
	t.Parallel()

	bus := New()
	defer func() { _ = bus.Close() }()

	unsub1, err := Subscribe(bus, func(ctx context.Context, evt userCreated) error { return nil })
	require.NoError(t, err)
	unsub2, err := Subscribe(bus, func(ctx context.Context, evt userCreated) error { return nil })
	require.NoError(t, err)

	require.Equal(t, 2, bus.SubscriberCount())
	unsub1()
	require.Equal(t, 1, bus.SubscriberCount())
	unsub2()
	require.Equal(t, 0, bus.SubscriberCount())
}

func TestParallelDispatchHandlersRunConcurrently(t *testing.T) {
	t.Parallel()

	bus := New(WithParallelDispatch(true))
	defer func() { _ = bus.Close() }()

	started := make(chan struct{}, 2)
	release := make(chan struct{})

	_, err := Subscribe(bus, func(ctx context.Context, evt userCreated) error {
		started <- struct{}{}
		<-release
		return nil
	})
	require.NoError(t, err)

	_, err = Subscribe(bus, func(ctx context.Context, evt userCreated) error {
		started <- struct{}{}
		<-release
		return nil
	})
	require.NoError(t, err)

	done := make(chan error, 1)
	go func() {
		done <- bus.Publish(context.Background(), userCreated{ID: 1})
	}()

	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("handlers did not start in parallel in time")
		}
	}

	close(release)

	select {
	case err = <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("publish did not finish in time")
	}
}

func TestParallelDispatchJoinErrors(t *testing.T) {
	t.Parallel()

	bus := New(WithParallelDispatch(true))
	defer func() { _ = bus.Close() }()

	_, err := Subscribe(bus, func(ctx context.Context, evt userCreated) error {
		return errors.New("err-a")
	})
	require.NoError(t, err)

	_, err = Subscribe(bus, func(ctx context.Context, evt userCreated) error {
		return errors.New("err-b")
	})
	require.NoError(t, err)

	err = bus.Publish(context.Background(), userCreated{ID: 1})
	require.Error(t, err)
	require.ErrorContains(t, err, "err-a")
	require.ErrorContains(t, err, "err-b")
}
