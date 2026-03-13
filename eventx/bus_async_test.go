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

func TestPublishAsyncNilContext(t *testing.T) {
	t.Parallel()

	bus := New(
		WithAsyncWorkers(1),
		WithAsyncQueueSize(8),
	)

	nilCtx := make(chan bool, 1)
	_, err := Subscribe(bus, func(ctx context.Context, evt userCreated) error {
		nilCtx <- ctx == nil
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

func TestLegacyAsyncCloseWhilePublishing(t *testing.T) {
	t.Parallel()

	bus := New(
		WithAsyncWorkers(1),
		WithAsyncQueueSize(4),
	)

	_, err := Subscribe(bus, func(ctx context.Context, evt userCreated) error {
		time.Sleep(2 * time.Millisecond)
		return nil
	})
	require.NoError(t, err)

	var wg sync.WaitGroup
	errCh := make(chan error, 200)
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			errCh <- bus.PublishAsync(context.Background(), userCreated{ID: index})
		}(i)
	}

	time.Sleep(5 * time.Millisecond)
	require.NoError(t, bus.Close())

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err == nil {
			continue
		}
		if errors.Is(err, ErrBusClosed) || errors.Is(err, ErrAsyncQueueFull) {
			continue
		}
		require.NoError(t, err)
	}
}

