package authx

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/DaiYuANg/arcgo/eventx"
	"github.com/DaiYuANg/arcgo/observability"
	"github.com/samber/lo"
)

// EventType represents the type of authx event.
type EventType string

// Authentication events.
const (
	// EventAuthSuccess is emitted when authentication succeeds.
	EventAuthSuccess EventType = "auth.success"
	// EventAuthFailure is emitted when authentication fails.
	EventAuthFailure EventType = "auth.failure"
)

// Authorization events.
const (
	// EventAuthzAllowed is emitted when authorization decision is allow.
	EventAuthzAllowed EventType = "authz.allowed"
	// EventAuthzDenied is emitted when authorization decision is deny.
	EventAuthzDenied EventType = "authz.denied"
)

// Policy events.
const (
	// EventPolicyLoaded is emitted when policies are loaded from source.
	EventPolicyLoaded EventType = "policy.loaded"
	// EventPolicyReplaced is emitted when policies are replaced.
	EventPolicyReplaced EventType = "policy.replaced"
	// EventPolicyReloadFailed is emitted when policy reload fails.
	EventPolicyReloadFailed EventType = "policy.reload_failed"
)

// Provider events.
const (
	// EventProviderFallback is emitted when provider chain falls back to next provider.
	EventProviderFallback EventType = "provider.fallback"
)

// AuthXEvent represents an authx-specific event that wraps eventx.Event.
// It implements the eventx.Event interface via the Name() method.
type AuthXEvent struct {
	// Type is the event type.
	Type EventType
	// Timestamp is when the event occurred.
	Timestamp time.Time
	// Identity is the associated identity (if any).
	Identity Identity
	// Credential is the credential used (if any).
	Credential Credential
	// Request is the authorization request (if any).
	Request Request
	// Decision is the authorization decision (if any).
	Decision Decision
	// Error is the associated error (if any).
	Error error
	// Metadata contains additional event-specific data.
	Metadata map[string]string
}

// Name returns the event name for eventx compatibility.
func (e *AuthXEvent) Name() string {
	return string(e.Type)
}

// NewAuthXEvent creates a new authx event with current timestamp.
func NewAuthXEvent(eventType EventType) *AuthXEvent {
	return &AuthXEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}
}

// WithIdentity attaches identity to the event.
func (e *AuthXEvent) WithIdentity(identity Identity) *AuthXEvent {
	e.Identity = identity
	return e
}

// WithCredential attaches credential to the event.
func (e *AuthXEvent) WithCredential(credential Credential) *AuthXEvent {
	e.Credential = credential
	return e
}

// WithRequest attaches authorization request to the event.
func (e *AuthXEvent) WithRequest(request Request) *AuthXEvent {
	e.Request = request
	return e
}

// WithDecision attaches authorization decision to the event.
func (e *AuthXEvent) WithDecision(decision Decision) *AuthXEvent {
	e.Decision = decision
	return e
}

// WithError attaches error to the event.
func (e *AuthXEvent) WithError(err error) *AuthXEvent {
	e.Error = err
	return e
}

// WithMetadata attaches metadata to the event.
func (e *AuthXEvent) WithMetadata(key, value string) *AuthXEvent {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return e
}

// EventHandler handles authx events.
// This is a wrapper around eventx.HandlerFunc for authx-specific events.
type EventHandler func(ctx context.Context, event *AuthXEvent) error

// EventHandlerAdapter adapts an authx EventHandler to eventx.HandlerFunc.
func EventHandlerAdapter(handler EventHandler) eventx.HandlerFunc {
	return func(ctx context.Context, event eventx.Event) error {
		if authxEvent, ok := event.(*AuthXEvent); ok && handler != nil {
			return handler(ctx, authxEvent)
		}
		return nil
	}
}

// LoggingEventHandler creates a logging event handler for authx events.
func LoggingEventHandler(logger *slog.Logger) EventHandler {
	log := normalizeLogger(logger)
	return func(ctx context.Context, event *AuthXEvent) error {
		if log == nil || event == nil {
			return nil
		}

		l := log.With("event", string(event.Type))

		switch event.Type {
		case EventAuthSuccess:
			if event.Identity != nil {
				l.Info("authentication success",
					"identity_id", event.Identity.ID(),
					"identity_type", event.Identity.Type())
			} else {
				l.Info("authentication success")
			}
		case EventAuthFailure:
			var errStr, credKind string
			if event.Error != nil {
				errStr = event.Error.Error()
			}
			if event.Credential != nil {
				credKind = event.Credential.Kind()
			}
			l.Warn("authentication failure",
				"error", errStr,
				"credential_kind", credKind)
		case EventAuthzAllowed:
			var identityID, action, resource string
			if event.Identity != nil {
				identityID = event.Identity.ID()
			}
			if event.Request.Action != "" {
				action = event.Request.Action
			}
			if event.Request.Resource != "" {
				resource = event.Request.Resource
			}
			l.Debug("authorization allowed",
				"identity_id", identityID,
				"action", action,
				"resource", resource)
		case EventAuthzDenied:
			var identityID, action, resource string
			if event.Identity != nil {
				identityID = event.Identity.ID()
			}
			if event.Request.Action != "" {
				action = event.Request.Action
			}
			if event.Request.Resource != "" {
				resource = event.Request.Resource
			}
			l.Warn("authorization denied",
				"identity_id", identityID,
				"action", action,
				"resource", resource)
		case EventPolicyLoaded:
			l.Info("policies loaded",
				"version", event.Metadata["version"],
				"permission_rules", event.Metadata["permission_rules"],
				"role_bindings", event.Metadata["role_bindings"])
		case EventPolicyReplaced:
			l.Info("policies replaced",
				"version", event.Metadata["version"])
		case EventPolicyReloadFailed:
			var errStr string
			if event.Error != nil {
				errStr = event.Error.Error()
			}
			l.Error("policy reload failed",
				"error", errStr)
		case EventProviderFallback:
			var errStr string
			if event.Error != nil {
				errStr = event.Error.Error()
			}
			l.Warn("provider fallback triggered",
				"error", errStr)
		}
		return nil
	}
}

// EventPublisher is a facade over eventx.Bus for authx-specific events.
// It provides convenient methods for publishing common authx events.
type EventPublisher struct {
	bus *eventx.Bus
}

// NewEventPublisher creates a new event publisher using eventx.Bus.
func NewEventPublisher(opts ...EventPublisherOption) *EventPublisher {
	cfg := eventPublisherConfig{}
	lo.ForEach(opts, func(opt EventPublisherOption, _ int) {
		if opt != nil {
			opt(&cfg)
		}
	})

	// Build eventx options
	eventxOpts := make([]eventx.Option, 0)

	if cfg.observability != nil {
		eventxOpts = append(eventxOpts, eventx.WithObservability(cfg.observability))
	}

	if cfg.parallel {
		eventxOpts = append(eventxOpts, eventx.WithParallelDispatch(true))
	}

	if cfg.asyncWorkers > 0 {
		eventxOpts = append(eventxOpts, eventx.WithAsyncWorkers(cfg.asyncWorkers))
		eventxOpts = append(eventxOpts, eventx.WithAsyncQueueSize(cfg.asyncQueueSize))
		if cfg.onAsyncError != nil {
			eventxOpts = append(eventxOpts, eventx.WithAsyncErrorHandler(cfg.onAsyncError))
		}
	}

	bus := eventx.New(eventxOpts...)

	// Add global logging middleware if logger is provided
	if cfg.logger != nil {
		bus = eventx.New(append(eventxOpts, eventx.WithMiddleware(
			eventx.RecoverMiddleware(),
		))...)
	}

	return &EventPublisher{
		bus: bus,
	}
}

// EventPublisherOption configures an EventPublisher.
type EventPublisherOption func(*eventPublisherConfig)

type eventPublisherConfig struct {
	observability  observability.Observability
	logger         *slog.Logger
	parallel       bool
	asyncWorkers   int
	asyncQueueSize int
	onAsyncError   func(ctx context.Context, event eventx.Event, err error)
}

// WithEventPublisherLogger sets the logger for the event publisher.
func WithEventPublisherLogger(logger *slog.Logger) EventPublisherOption {
	return func(cfg *eventPublisherConfig) {
		cfg.logger = logger
	}
}

// WithEventPublisherParallel enables parallel event dispatch.
func WithEventPublisherParallel(enabled bool) EventPublisherOption {
	return func(cfg *eventPublisherConfig) {
		cfg.parallel = enabled
	}
}

// WithEventPublisherAsync configures async event publishing.
func WithEventPublisherAsync(workers, queueSize int, onError func(ctx context.Context, event eventx.Event, err error)) EventPublisherOption {
	return func(cfg *eventPublisherConfig) {
		cfg.asyncWorkers = workers
		cfg.asyncQueueSize = queueSize
		cfg.onAsyncError = onError
	}
}

// WithEventPublisherObservability sets observability for the event publisher.
func WithEventPublisherObservability(obs observability.Observability) EventPublisherOption {
	return func(cfg *eventPublisherConfig) {
		cfg.observability = obs
	}
}

// Close stops the event publisher and waits for in-flight events.
func (p *EventPublisher) Close() error {
	if p == nil || p.bus == nil {
		return nil
	}
	return p.bus.Close()
}

// Subscribe registers an authx event handler.
func (p *EventPublisher) Subscribe(handler EventHandler) (func(), error) {
	if p == nil || p.bus == nil || handler == nil {
		return func() {}, nil
	}

	// Wrap authx handler to eventx handler
	wrappedHandler := func(ctx context.Context, event *AuthXEvent) error {
		return handler(ctx, event)
	}

	return eventx.Subscribe(p.bus, wrappedHandler)
}

// Publish sends an event to the bus.
func (p *EventPublisher) Publish(ctx context.Context, event *AuthXEvent) error {
	if p == nil || p.bus == nil || event == nil {
		return nil
	}
	return p.bus.Publish(ctx, event)
}

// PublishAsync sends an event for async dispatch.
func (p *EventPublisher) PublishAsync(ctx context.Context, event *AuthXEvent) error {
	if p == nil || p.bus == nil || event == nil {
		return nil
	}
	return p.bus.PublishAsync(ctx, event)
}

// PublishAuthSuccess publishes an authentication success event.
func (p *EventPublisher) PublishAuthSuccess(ctx context.Context, identity Identity) error {
	if p == nil || p.bus == nil || identity == nil {
		return nil
	}
	event := NewAuthXEvent(EventAuthSuccess).WithIdentity(identity)
	return p.bus.Publish(ctx, event)
}

// PublishAuthFailure publishes an authentication failure event.
func (p *EventPublisher) PublishAuthFailure(ctx context.Context, credential Credential, err error) error {
	if p == nil || p.bus == nil {
		return nil
	}
	event := NewAuthXEvent(EventAuthFailure).
		WithCredential(credential).
		WithError(err)
	return p.bus.Publish(ctx, event)
}

// PublishAuthzAllowed publishes an authorization allowed event.
func (p *EventPublisher) PublishAuthzAllowed(ctx context.Context, identity Identity, request Request) error {
	if p == nil || p.bus == nil || identity == nil {
		return nil
	}
	event := NewAuthXEvent(EventAuthzAllowed).
		WithIdentity(identity).
		WithRequest(request)
	return p.bus.Publish(ctx, event)
}

// PublishAuthzDenied publishes an authorization denied event.
func (p *EventPublisher) PublishAuthzDenied(ctx context.Context, identity Identity, request Request) error {
	if p == nil || p.bus == nil || identity == nil {
		return nil
	}
	event := NewAuthXEvent(EventAuthzDenied).
		WithIdentity(identity).
		WithRequest(request)
	return p.bus.Publish(ctx, event)
}

// PublishPolicyLoaded publishes a policy loaded event.
func (p *EventPublisher) PublishPolicyLoaded(ctx context.Context, version int64, permissions, roleBindings int) error {
	if p == nil || p.bus == nil {
		return nil
	}
	event := NewAuthXEvent(EventPolicyLoaded).
		WithMetadata("version", strconv.FormatInt(version, 10)).
		WithMetadata("permission_rules", strconv.Itoa(permissions)).
		WithMetadata("role_bindings", strconv.Itoa(roleBindings))
	return p.bus.Publish(ctx, event)
}

// PublishPolicyReplaced publishes a policy replaced event.
func (p *EventPublisher) PublishPolicyReplaced(ctx context.Context, version int64) error {
	if p == nil || p.bus == nil {
		return nil
	}
	event := NewAuthXEvent(EventPolicyReplaced).
		WithMetadata("version", strconv.FormatInt(version, 10))
	return p.bus.Publish(ctx, event)
}

// PublishPolicyReloadFailed publishes a policy reload failed event.
func (p *EventPublisher) PublishPolicyReloadFailed(ctx context.Context, err error) error {
	if p == nil || p.bus == nil {
		return nil
	}
	event := NewAuthXEvent(EventPolicyReloadFailed).
		WithError(err)
	return p.bus.Publish(ctx, event)
}

// PublishProviderFallback publishes a provider fallback event.
func (p *EventPublisher) PublishProviderFallback(ctx context.Context, err error) error {
	if p == nil || p.bus == nil {
		return nil
	}
	event := NewAuthXEvent(EventProviderFallback).
		WithError(err)
	return p.bus.Publish(ctx, event)
}

// SubscriberCount returns the number of active subscribers.
func (p *EventPublisher) SubscriberCount() int {
	if p == nil || p.bus == nil {
		return 0
	}
	return p.bus.SubscriberCount()
}
