package authx

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventType(t *testing.T) {
	tests := []struct {
		name string
		t    EventType
	}{
		{"AuthSuccess", EventAuthSuccess},
		{"AuthFailure", EventAuthFailure},
		{"AuthzAllowed", EventAuthzAllowed},
		{"AuthzDenied", EventAuthzDenied},
		{"PolicyLoaded", EventPolicyLoaded},
		{"PolicyReplaced", EventPolicyReplaced},
		{"PolicyReloadFailed", EventPolicyReloadFailed},
		{"ProviderFallback", EventProviderFallback},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, string(tt.t))
		})
	}
}

func TestAuthXEvent(t *testing.T) {
	t.Run("new event with timestamp", func(t *testing.T) {
		event := NewAuthXEvent(EventAuthSuccess)
		assert.Equal(t, EventAuthSuccess, event.Type)
		assert.NotEmpty(t, event.Timestamp)
		assert.NotNil(t, event.Metadata)
	})

	t.Run("with identity", func(t *testing.T) {
		identity := NewIdentity("u1", "user", "Test")
		event := NewAuthXEvent(EventAuthSuccess).WithIdentity(identity)
		assert.Equal(t, identity, event.Identity)
	})

	t.Run("with credential", func(t *testing.T) {
		cred := &PasswordCredential{Password: "secret"}
		event := NewAuthXEvent(EventAuthFailure).WithCredential(cred)
		assert.Equal(t, cred, event.Credential)
	})

	t.Run("with request", func(t *testing.T) {
		request := NewRequest("read", "/api/users", nil)
		event := NewAuthXEvent(EventAuthzAllowed).WithRequest(request)
		assert.Equal(t, request, event.Request)
	})

	t.Run("with decision", func(t *testing.T) {
		decision := Decision{Allowed: true}
		event := NewAuthXEvent(EventAuthzAllowed).WithDecision(decision)
		assert.Equal(t, decision, event.Decision)
	})

	t.Run("with error", func(t *testing.T) {
		err := errors.New("test error")
		event := NewAuthXEvent(EventAuthFailure).WithError(err)
		assert.Equal(t, err, event.Error)
	})

	t.Run("with metadata", func(t *testing.T) {
		event := NewAuthXEvent(EventPolicyLoaded).WithMetadata("key", "value")
		assert.Equal(t, "value", event.Metadata["key"])
	})

	t.Run("with multiple metadata", func(t *testing.T) {
		event := NewAuthXEvent(EventPolicyLoaded).
			WithMetadata("version", "1").
			WithMetadata("rules", "10")
		assert.Equal(t, "1", event.Metadata["version"])
		assert.Equal(t, "10", event.Metadata["rules"])
	})

	t.Run("Name returns type string", func(t *testing.T) {
		event := NewAuthXEvent(EventAuthSuccess)
		assert.Equal(t, "auth.success", event.Name())
	})
}

func TestEventHandlerAdapter(t *testing.T) {
	t.Run("executes function", func(t *testing.T) {
		called := int32(0)
		handler := func(ctx context.Context, event *AuthXEvent) error {
			atomic.AddInt32(&called, 1)
			return nil
		}

		adapter := EventHandlerAdapter(handler)
		event := NewAuthXEvent(EventAuthSuccess)
		err := adapter(context.Background(), event)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&called))
	})

	t.Run("nil function is no-op", func(t *testing.T) {
		adapter := EventHandlerAdapter(nil)
		event := NewAuthXEvent(EventAuthSuccess)
		err := adapter(context.Background(), event)
		assert.NoError(t, err)
	})

	t.Run("wrong event type is no-op", func(t *testing.T) {
		called := false
		handler := func(ctx context.Context, event *AuthXEvent) error {
			called = true
			return nil
		}

		adapter := EventHandlerAdapter(handler)
		wrongEvent := &testEvent{name: "wrong"}
		err := adapter(context.Background(), wrongEvent)
		assert.NoError(t, err)
		assert.False(t, called)
	})
}

func TestLoggingEventHandler(t *testing.T) {
	t.Run("logs auth success", func(t *testing.T) {
		logger := slog.Default()
		handler := LoggingEventHandler(logger)
		identity := NewIdentity("u1", "user", "Test")
		event := NewAuthXEvent(EventAuthSuccess).WithIdentity(identity)
		err := handler(context.Background(), event)
		assert.NoError(t, err)
	})

	t.Run("nil logger is no-op", func(t *testing.T) {
		handler := LoggingEventHandler(nil)
		event := NewAuthXEvent(EventAuthSuccess)
		err := handler(context.Background(), event)
		assert.NoError(t, err)
	})

	t.Run("nil event is no-op", func(t *testing.T) {
		logger := slog.Default()
		handler := LoggingEventHandler(logger)
		err := handler(context.Background(), nil)
		assert.NoError(t, err)
	})
}

func TestEventPublisher(t *testing.T) {
	t.Run("publish auth success", func(t *testing.T) {
		publisher := NewEventPublisher()
		defer func() { _ = publisher.Close() }()

		called := int32(0)
		handler := func(ctx context.Context, event *AuthXEvent) error {
			atomic.AddInt32(&called, 1)
			assert.Equal(t, EventAuthSuccess, event.Type)
			return nil
		}

		_, _ = publisher.Subscribe(handler)
		identity := NewIdentity("u1", "user", "Test")
		err := publisher.PublishAuthSuccess(context.Background(), identity)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&called))
	})

	t.Run("publish auth failure", func(t *testing.T) {
		publisher := NewEventPublisher()
		defer func() { _ = publisher.Close() }()

		called := int32(0)
		handler := func(ctx context.Context, event *AuthXEvent) error {
			atomic.AddInt32(&called, 1)
			assert.Equal(t, EventAuthFailure, event.Type)
			return nil
		}

		_, _ = publisher.Subscribe(handler)
		cred := &PasswordCredential{Password: "secret"}
		err := errors.New("auth failed")
		publishErr := publisher.PublishAuthFailure(context.Background(), cred, err)
		assert.NoError(t, publishErr)
		assert.Equal(t, int32(1), atomic.LoadInt32(&called))
	})

	t.Run("publish authz allowed", func(t *testing.T) {
		publisher := NewEventPublisher()
		defer func() { _ = publisher.Close() }()

		called := int32(0)
		handler := func(ctx context.Context, event *AuthXEvent) error {
			atomic.AddInt32(&called, 1)
			return nil
		}

		_, _ = publisher.Subscribe(handler)
		identity := NewIdentity("u1", "user", "Test")
		request := NewRequest("read", "/api/users", nil)
		err := publisher.PublishAuthzAllowed(context.Background(), identity, request)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&called))
	})

	t.Run("publish authz denied", func(t *testing.T) {
		publisher := NewEventPublisher()
		defer func() { _ = publisher.Close() }()

		called := int32(0)
		handler := func(ctx context.Context, event *AuthXEvent) error {
			atomic.AddInt32(&called, 1)
			return nil
		}

		_, _ = publisher.Subscribe(handler)
		identity := NewIdentity("u1", "user", "Test")
		request := NewRequest("write", "/api/admin", nil)
		err := publisher.PublishAuthzDenied(context.Background(), identity, request)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&called))
	})

	t.Run("publish policy loaded", func(t *testing.T) {
		publisher := NewEventPublisher()
		defer func() { _ = publisher.Close() }()

		called := int32(0)
		handler := func(ctx context.Context, event *AuthXEvent) error {
			atomic.AddInt32(&called, 1)
			assert.Equal(t, "5", event.Metadata["version"])
			return nil
		}

		_, _ = publisher.Subscribe(handler)
		err := publisher.PublishPolicyLoaded(context.Background(), 5, 10, 3)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&called))
	})

	t.Run("publish policy replaced", func(t *testing.T) {
		publisher := NewEventPublisher()
		defer func() { _ = publisher.Close() }()

		called := int32(0)
		handler := func(ctx context.Context, event *AuthXEvent) error {
			atomic.AddInt32(&called, 1)
			assert.Equal(t, "10", event.Metadata["version"])
			return nil
		}

		_, _ = publisher.Subscribe(handler)
		err := publisher.PublishPolicyReplaced(context.Background(), 10)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&called))
	})

	t.Run("publish policy reload failed", func(t *testing.T) {
		publisher := NewEventPublisher()
		defer func() { _ = publisher.Close() }()

		called := int32(0)
		handler := func(ctx context.Context, event *AuthXEvent) error {
			atomic.AddInt32(&called, 1)
			return nil
		}

		_, _ = publisher.Subscribe(handler)
		err := errors.New("reload failed")
		publishErr := publisher.PublishPolicyReloadFailed(context.Background(), err)
		assert.NoError(t, publishErr)
		assert.Equal(t, int32(1), atomic.LoadInt32(&called))
	})

	t.Run("publish provider fallback", func(t *testing.T) {
		publisher := NewEventPublisher()
		defer func() { _ = publisher.Close() }()

		called := int32(0)
		handler := func(ctx context.Context, event *AuthXEvent) error {
			atomic.AddInt32(&called, 1)
			return nil
		}

		_, _ = publisher.Subscribe(handler)
		err := errors.New("provider failed")
		publishErr := publisher.PublishProviderFallback(context.Background(), err)
		assert.NoError(t, publishErr)
		assert.Equal(t, int32(1), atomic.LoadInt32(&called))
	})

	t.Run("nil publisher is no-op", func(t *testing.T) {
		var publisher *EventPublisher
		identity := NewIdentity("u1", "user", "Test")
		err := publisher.PublishAuthSuccess(context.Background(), identity)
		assert.NoError(t, err)
	})

	t.Run("subscribe and unsubscribe", func(t *testing.T) {
		publisher := NewEventPublisher()
		defer func() { _ = publisher.Close() }()

		called := int32(0)
		handler := func(ctx context.Context, event *AuthXEvent) error {
			atomic.AddInt32(&called, 1)
			return nil
		}

		unsub, err := publisher.Subscribe(handler)
		assert.NoError(t, err)
		assert.NotNil(t, unsub)

		// Publish before unsubscribe
		err = publisher.Publish(context.Background(), NewAuthXEvent(EventAuthSuccess))
		assert.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&called))

		// Unsubscribe
		unsub()

		// Publish after unsubscribe - should not be called
		err = publisher.Publish(context.Background(), NewAuthXEvent(EventAuthSuccess))
		assert.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&called))
	})

	t.Run("subscriber count", func(t *testing.T) {
		publisher := NewEventPublisher()
		defer func() { _ = publisher.Close() }()

		assert.Equal(t, 0, publisher.SubscriberCount())

		unsub1, _ := publisher.Subscribe(func(ctx context.Context, event *AuthXEvent) error {
			return nil
		})
		assert.Equal(t, 1, publisher.SubscriberCount())

		unsub2, _ := publisher.Subscribe(func(ctx context.Context, event *AuthXEvent) error {
			return nil
		})
		assert.Equal(t, 2, publisher.SubscriberCount())

		unsub1()
		assert.Equal(t, 1, publisher.SubscriberCount())

		unsub2()
		assert.Equal(t, 0, publisher.SubscriberCount())
	})
}

func TestEventPublisherWithOptions(t *testing.T) {
	t.Run("with logger", func(t *testing.T) {
		logger := slog.Default()
		publisher := NewEventPublisher(WithEventPublisherLogger(logger))
		defer func() { _ = publisher.Close() }()

		_, _ = publisher.Subscribe(LoggingEventHandler(logger))
		identity := NewIdentity("u1", "user", "Test")
		err := publisher.PublishAuthSuccess(context.Background(), identity)
		assert.NoError(t, err)
	})

	t.Run("with parallel dispatch", func(t *testing.T) {
		publisher := NewEventPublisher(WithEventPublisherParallel(true))
		defer func() { _ = publisher.Close() }()

		called := int32(0)
		for i := 0; i < 3; i++ {
			_, _ = publisher.Subscribe(func(ctx context.Context, event *AuthXEvent) error {
				atomic.AddInt32(&called, 1)
				return nil
			})
		}

		err := publisher.Publish(context.Background(), NewAuthXEvent(EventAuthSuccess))
		assert.NoError(t, err)
		assert.Equal(t, int32(3), atomic.LoadInt32(&called))
	})

	t.Run("close is idempotent", func(t *testing.T) {
		publisher := NewEventPublisher()
		err1 := publisher.Close()
		err2 := publisher.Close()
		assert.NoError(t, err1)
		assert.NoError(t, err2)
	})
}

func TestStrconvHelpers(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"zero", 0, "0"},
		{"positive", 123, "123"},
		{"negative", -456, "-456"},
		{"one", 1, "1"},
		{"large", 999999, "999999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strconv.Itoa(tt.input)
			if got != tt.expected {
				t.Errorf("strconv.Itoa(%d) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestItoa64(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"zero", 0, "0"},
		{"positive", 123, "123"},
		{"negative", -456, "-456"},
		{"one", 1, "1"},
		{"large", 9999999999, "9999999999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strconv.FormatInt(tt.input, 10)
			if got != tt.expected {
				t.Errorf("strconv.FormatInt(%d) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// testEvent is a test helper implementing eventx.Event.
type testEvent struct {
	name string
}

func (e *testEvent) Name() string {
	return e.name
}

func TestManagerWithCustomEventPublisher(t *testing.T) {
	t.Run("custom event publisher is used", func(t *testing.T) {
		called := int32(0)

		// Create custom event publisher
		customPublisher := NewEventPublisher()
		defer func() { _ = customPublisher.Close() }()

		_, _ = customPublisher.Subscribe(func(ctx context.Context, event *AuthXEvent) error {
			atomic.AddInt32(&called, 1)
			return nil
		})

		// Create manager with custom publisher
		manager, err := NewManager(
			WithEventPublisher(customPublisher),
		)
		assert.NoError(t, err)
		assert.NotNil(t, manager)

		// Verify custom publisher is returned
		assert.Equal(t, customPublisher, manager.EventPublisher())

		// Publish event through manager's publisher
		identity := NewIdentity("u1", "user", "Test")
		err = manager.EventPublisher().PublishAuthSuccess(context.Background(), identity)
		assert.NoError(t, err)

		// Verify custom handler was called
		assert.Equal(t, int32(1), atomic.LoadInt32(&called))
	})

	t.Run("default event publisher is created when not provided", func(t *testing.T) {
		logger := slog.Default()

		manager, err := NewManager(
			WithLogger(logger),
		)
		assert.NoError(t, err)
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.EventPublisher())

		// Should not panic
		assert.NotNil(t, manager.EventPublisher())
	})
}
