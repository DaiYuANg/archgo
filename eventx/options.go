package eventx

import (
	"context"

	"github.com/DaiYuANg/arcgo/observability"
)

const (
	defaultAsyncWorkers   = 4
	defaultAsyncQueueSize = 256
)

type asyncErrorHandler func(ctx context.Context, event Event, err error)

type options struct {
	// ants pool options
	useAntsPool          bool
	antsPoolSize         int
	antsMaxBlockingCalls int

	// legacy options (for backward compatibility)
	asyncWorkers   int
	asyncQueueSize int

	parallel      bool
	middleware    []Middleware
	onAsyncError  asyncErrorHandler
	observability observability.Observability
}

func defaultOptions() options {
	return options{
		useAntsPool:          true,
		antsPoolSize:         defaultAsyncWorkers,
		antsMaxBlockingCalls: -1, // -1 means infinite blocking calls

		// legacy options
		asyncWorkers:   defaultAsyncWorkers,
		asyncQueueSize: defaultAsyncQueueSize,

		parallel:      false,
		middleware:    nil,
		onAsyncError:  nil,
		observability: observability.Nop(),
	}
}

// Option configures Bus.
type Option func(*options)

// WithAntsPool enables ants goroutine pool with the given size.
// This is the recommended way for async event dispatch.
func WithAntsPool(size int) Option {
	return func(o *options) {
		o.useAntsPool = true
		o.antsPoolSize = size
	}
}

// WithAntsPoolWithMaxBlockingCalls configures ants pool with max blocking calls limit.
// maxBlockingCalls <= 0 means infinite.
func WithAntsPoolWithMaxBlockingCalls(size int, maxBlockingCalls int) Option {
	return func(o *options) {
		o.useAntsPool = true
		o.antsPoolSize = size
		o.antsMaxBlockingCalls = maxBlockingCalls
	}
}

// WithAsyncWorkers sets worker count for async publish (legacy mode).
// Values <= 0 disable async workers.
// Deprecated: Use WithAntsPool instead for better performance.
func WithAsyncWorkers(workers int) Option {
	return func(o *options) {
		o.useAntsPool = false
		o.asyncWorkers = workers
	}
}

// WithAsyncQueueSize sets async queue size (legacy mode).
// Values <= 0 disable async queueing.
// Deprecated: Ants pool handles queueing internally.
func WithAsyncQueueSize(size int) Option {
	return func(o *options) {
		o.useAntsPool = false
		o.asyncQueueSize = size
	}
}

// WithParallelDispatch controls whether handlers of the same event are dispatched in parallel.
// Default is false (serial dispatch).
func WithParallelDispatch(enabled bool) Option {
	return func(o *options) {
		o.parallel = enabled
	}
}

// WithMiddleware appends global middleware.
func WithMiddleware(mw ...Middleware) Option {
	return func(o *options) {
		o.middleware = append(o.middleware, mw...)
	}
}

// WithAsyncErrorHandler sets callback for async dispatch errors.
func WithAsyncErrorHandler(handler func(ctx context.Context, event Event, err error)) Option {
	return func(o *options) {
		o.onAsyncError = handler
	}
}

// WithObservability sets optional observability integration for bus runtime.
func WithObservability(obs observability.Observability) Option {
	return func(o *options) {
		o.observability = obs
	}
}

type subscribeOptions struct {
	middleware []Middleware
}

func defaultSubscribeOptions() subscribeOptions {
	return subscribeOptions{
		middleware: nil,
	}
}

// SubscribeOption configures per-subscription behavior.
type SubscribeOption func(*subscribeOptions)

// WithSubscriberMiddleware appends subscription-level middleware.
func WithSubscriberMiddleware(mw ...Middleware) SubscribeOption {
	return func(o *subscribeOptions) {
		o.middleware = append(o.middleware, mw...)
	}
}
