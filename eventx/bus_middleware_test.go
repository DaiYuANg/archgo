package eventx

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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

func TestCloseWaitsInFlightSyncDispatch(t *testing.T) {
	t.Parallel()

	bus := New()

	started := make(chan struct{})
	release := make(chan struct{})

	_, err := Subscribe(bus, func(ctx context.Context, evt userCreated) error {
		close(started)
		<-release
		return nil
	})
	require.NoError(t, err)

	publishDone := make(chan error, 1)
	go func() {
		publishDone <- bus.Publish(context.Background(), userCreated{ID: 1})
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("sync handler did not start in time")
	}

	closeDone := make(chan error, 1)
	go func() {
		closeDone <- bus.Close()
	}()

	select {
	case err = <-closeDone:
		require.NoError(t, err)
		t.Fatal("close returned before in-flight sync dispatch finished")
	case <-time.After(50 * time.Millisecond):
	}

	close(release)

	select {
	case err = <-publishDone:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("publish did not finish in time")
	}

	select {
	case err = <-closeDone:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("close did not finish in time")
	}
}
