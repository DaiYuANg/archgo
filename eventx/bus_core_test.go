package eventx

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

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

func TestSubscribeOnce(t *testing.T) {
	t.Parallel()

	bus := New()
	defer func() { _ = bus.Close() }()

	var count int64
	_, err := SubscribeOnce(bus, func(ctx context.Context, evt userCreated) error {
		atomic.AddInt64(&count, 1)
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, bus.Publish(context.Background(), userCreated{ID: 1}))
	require.NoError(t, bus.Publish(context.Background(), userCreated{ID: 2}))
	require.EqualValues(t, 1, atomic.LoadInt64(&count))
}

func TestSubscribeN(t *testing.T) {
	t.Parallel()

	bus := New()
	defer func() { _ = bus.Close() }()

	var count int64
	_, err := SubscribeN(bus, 2, func(ctx context.Context, evt userCreated) error {
		atomic.AddInt64(&count, 1)
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, bus.Publish(context.Background(), userCreated{ID: 1}))
	require.NoError(t, bus.Publish(context.Background(), userCreated{ID: 2}))
	require.NoError(t, bus.Publish(context.Background(), userCreated{ID: 3}))
	require.EqualValues(t, 2, atomic.LoadInt64(&count))
}

func TestSubscribeNInvalidCount(t *testing.T) {
	t.Parallel()

	bus := New()
	defer func() { _ = bus.Close() }()

	_, err := SubscribeN(bus, 0, func(ctx context.Context, evt userCreated) error {
		return nil
	})
	require.ErrorIs(t, err, ErrInvalidSubscribeCount)
}

