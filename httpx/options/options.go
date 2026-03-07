// Package options provides package-level APIs.
package options

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/httpx/adapter"
	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
)

// ServerOptions collects higher-level server construction settings.
type ServerOptions struct {
	Adapter            adapter.Adapter
	Logger             *slog.Logger
	BasePath           string
	PrintRoutes        bool
	EnableValidation   bool
	Validator          *validator.Validate
	OpenAPIDocsEnabled bool
	HumaTitle          string
	HumaVersion        string
	HumaDescription    string
	DocsPath           string
	OpenAPIPath        string
	SchemasPath        string
	DocsRenderer       string
	EnablePanicRecover bool
	EnableAccessLog    bool
}

// DefaultServerOptions provides default behavior.
func DefaultServerOptions() *ServerOptions {
	return &ServerOptions{
		Logger:             slog.Default(),
		PrintRoutes:        false,
		OpenAPIDocsEnabled: true,
		HumaTitle:          "My API",
		HumaVersion:        "1.0.0",
		HumaDescription:    "API Documentation",
		DocsPath:           "/docs",
		OpenAPIPath:        "/openapi",
		SchemasPath:        "/schemas",
		DocsRenderer:       httpx.DocsRendererStoplightElements,
		EnablePanicRecover: true,
		EnableAccessLog:    false,
	}
}

// ServerOption mutates `ServerOptions`.
type ServerOption func(*ServerOptions)

// Compose combines multiple option functions into one.
func Compose(opts ...ServerOption) ServerOption {
	return func(o *ServerOptions) {
		lo.ForEach(opts, func(opt ServerOption, _ int) {
			if opt != nil {
				opt(o)
			}
		})
	}
}

// WithAdapter configures related behavior.
func WithAdapter(adapter adapter.Adapter) ServerOption {
	return func(o *ServerOptions) {
		o.Adapter = adapter
	}
}

// WithLogger configures related behavior.
func WithLogger(logger *slog.Logger) ServerOption {
	return func(o *ServerOptions) {
		o.Logger = logger
	}
}

// WithBasePath configures related behavior.
func WithBasePath(path string) ServerOption {
	return func(o *ServerOptions) {
		o.BasePath = path
	}
}

// WithPrintRoutes configures related behavior.
func WithPrintRoutes(enabled bool) ServerOption {
	return func(o *ServerOptions) {
		o.PrintRoutes = enabled
	}
}

// WithValidation configures related behavior.
func WithValidation(enabled bool) ServerOption {
	return func(o *ServerOptions) {
		o.EnableValidation = enabled
	}
}

// WithValidator configures related behavior.
func WithValidator(v *validator.Validate) ServerOption {
	return func(o *ServerOptions) {
		o.Validator = v
	}
}

// WithOpenAPIDocs configures related behavior.
func WithOpenAPIDocs(enabled bool) ServerOption {
	return func(o *ServerOptions) {
		o.OpenAPIDocsEnabled = enabled
	}
}

// WithOpenAPIInfo sets OpenAPI title, version, and description fields.
func WithOpenAPIInfo(title, version, description string) ServerOption {
	return func(o *ServerOptions) {
		o.HumaTitle = title
		o.HumaVersion = version
		o.HumaDescription = description
	}
}

// WithPanicRecover enables or disables panic recovery for typed httpx handlers.
func WithPanicRecover(enabled bool) ServerOption {
	return func(o *ServerOptions) {
		o.EnablePanicRecover = enabled
	}
}

// WithAccessLog enables or disables request logging in the httpx layer.
func WithAccessLog(enabled bool) ServerOption {
	return func(o *ServerOptions) {
		o.EnableAccessLog = enabled
	}
}

// Build converts `ServerOptions` into `httpx.ServerOption` values.
func (o *ServerOptions) Build() []httpx.ServerOption {
	opts := []httpx.ServerOption{
		httpx.WithLogger(o.Logger),
		httpx.WithPrintRoutes(o.PrintRoutes),
	}

	if o.Adapter != nil {
		opts = append(opts, httpx.WithAdapter(o.Adapter))
	}

	if o.BasePath != "" {
		opts = append(opts, httpx.WithBasePath(o.BasePath))
	}

	if o.Validator != nil {
		opts = append(opts, httpx.WithValidator(o.Validator))
	} else if o.EnableValidation {
		opts = append(opts, httpx.WithValidation())
	}

	opts = append(opts, httpx.WithOpenAPIInfo(o.HumaTitle, o.HumaVersion, o.HumaDescription))
	opts = append(opts, httpx.WithOpenAPIDocs(o.OpenAPIDocsEnabled))
	opts = append(opts, httpx.WithDocs(httpx.DocsOptions{
		Enabled:     o.OpenAPIDocsEnabled,
		DocsPath:    o.DocsPath,
		OpenAPIPath: o.OpenAPIPath,
		SchemasPath: o.SchemasPath,
		Renderer:    o.DocsRenderer,
	}))
	opts = append(opts, httpx.WithPanicRecover(o.EnablePanicRecover))
	opts = append(opts, httpx.WithAccessLog(o.EnableAccessLog))

	return opts
}

// HTTPClientOptions collects standard `http.Client` construction settings.
type HTTPClientOptions struct {
	Timeout   time.Duration
	Transport http.RoundTripper
	Jar       http.CookieJar
}

// DefaultHTTPClientOptions provides default behavior.
func DefaultHTTPClientOptions() *HTTPClientOptions {
	return &HTTPClientOptions{
		Timeout: 30 * time.Second,
	}
}

// HTTPClientOption mutates `HTTPClientOptions`.
type HTTPClientOption func(*HTTPClientOptions)

// WithHTTPTimeout configures related behavior.
func WithHTTPTimeout(timeout time.Duration) HTTPClientOption {
	return func(o *HTTPClientOptions) {
		o.Timeout = timeout
	}
}

// WithHTTPTransport configures related behavior.
func WithHTTPTransport(transport http.RoundTripper) HTTPClientOption {
	return func(o *HTTPClientOptions) {
		o.Transport = transport
	}
}

// WithHTTPCookieJar configures related behavior.
func WithHTTPCookieJar(jar http.CookieJar) HTTPClientOption {
	return func(o *HTTPClientOptions) {
		o.Jar = jar
	}
}

// Build constructs an `http.Client` from the configured options.
func (o *HTTPClientOptions) Build() *http.Client {
	return &http.Client{
		Timeout:   o.Timeout,
		Transport: o.Transport,
		Jar:       o.Jar,
	}
}

// ContextOptions collects helper settings for building a context.Context.
type ContextOptions struct {
	Timeout   time.Duration
	Deadline  time.Time
	ValueKeys map[contextValueKey]any
}

type contextValueKey string

// ContextOption mutates `ContextOptions`.
type ContextOption func(*ContextOptions)

// WithContextTimeout configures related behavior.
func WithContextTimeout(timeout time.Duration) ContextOption {
	return func(o *ContextOptions) {
		o.Timeout = timeout
	}
}

// WithContextDeadline configures related behavior.
func WithContextDeadline(deadline time.Time) ContextOption {
	return func(o *ContextOptions) {
		o.Deadline = deadline
	}
}

// WithContextValue configures related behavior.
func WithContextValue(key string, value any) ContextOption {
	return func(o *ContextOptions) {
		if o.ValueKeys == nil {
			o.ValueKeys = make(map[contextValueKey]any)
		}
		o.ValueKeys[contextValueKey(key)] = value
	}
}

// Build creates a context and optional cancel function from the configured values.
func (o *ContextOptions) Build() (context.Context, context.CancelFunc) {
	var ctx context.Context
	var cancel context.CancelFunc

	if o.Deadline.IsZero() && o.Timeout == 0 {
		ctx = context.Background()
	} else if !o.Deadline.IsZero() {
		ctx, cancel = context.WithDeadline(context.Background(), o.Deadline)
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), o.Timeout)
	}

	for k, v := range o.ValueKeys {
		ctx = context.WithValue(ctx, k, v)
	}

	return ctx, cancel
}

// WithContextValueOpt mutates a ContextOptions value directly.
func WithContextValueOpt(o *ContextOptions, key string, value any) *ContextOptions {
	if o.ValueKeys == nil {
		o.ValueKeys = make(map[contextValueKey]any)
	}
	o.ValueKeys[contextValueKey(key)] = value
	return o
}
