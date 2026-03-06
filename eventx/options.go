package eventx

import "context"

const (
	defaultAsyncWorkers   = 4
	defaultAsyncQueueSize = 256
)

type asyncErrorHandler func(ctx context.Context, event Event, err error)

type options struct {
	asyncWorkers   int
	asyncQueueSize int
	parallel       bool
	middleware     []Middleware
	onAsyncError   asyncErrorHandler
}

func defaultOptions() options {
	return options{
		asyncWorkers:   defaultAsyncWorkers,
		asyncQueueSize: defaultAsyncQueueSize,
		parallel:       false,
		middleware:     nil,
		onAsyncError:   nil,
	}
}

// Option configures Bus.
type Option func(*options)

// WithAsyncWorkers sets worker count for async publish. Values <= 0 disable async workers.
func WithAsyncWorkers(workers int) Option {
	return func(o *options) {
		o.asyncWorkers = workers
	}
}

// WithAsyncQueueSize sets async queue size. Values <= 0 disable async queueing.
func WithAsyncQueueSize(size int) Option {
	return func(o *options) {
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
