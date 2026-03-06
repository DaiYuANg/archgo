// Package adapter provides package-level APIs.
package adapter

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

// HandlerFunc documents related behavior.
type HandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// MiddlewareFunc documents related behavior.
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

type routeParamsCtxKey struct{}

// Adapter documents related behavior.
// Note.
// Note.
// Note.
// Note.
type Adapter interface {
	// Name returns related data.
	Name() string

	// Handle registers related handlers.
	Handle(method, path string, handler HandlerFunc)

	// Group creates related functionality.
	Group(prefix string) Adapter

	// ServeHTTP documents related behavior.
	ServeHTTP(w http.ResponseWriter, r *http.Request)

	// HumaAPI returns related data.
	HumaAPI() huma.API

	// HasHuma enables related functionality.
	HasHuma() bool
}

// RouterAdapter exposes router objects with strong typing.
type RouterAdapter[R any] interface {
	Adapter
	Router() R
}

// HumaConfigurator configures related behavior.
type HumaConfigurator interface {
	ConfigureHuma(opts HumaOptions)
}

// ListenableAdapter starts related services.
type ListenableAdapter interface {
	Listen(addr string) error
}

// ContextListenableAdapter starts related services.
type ContextListenableAdapter interface {
	ListenContext(ctx context.Context, addr string) error
}

// HumaOptions documents related behavior.
type HumaOptions struct {
	// Title documents related behavior.
	Title string
	// Version documents related behavior.
	Version string
	// Description documents related behavior.
	Description string
	// DocsPath provides default behavior.
	DocsPath string
	// OpenAPIPath provides default behavior.
	OpenAPIPath string
	// DisableDocsRoutes closes related resources.
	DisableDocsRoutes bool
}

// DefaultHumaOptions provides default behavior.
func DefaultHumaOptions() HumaOptions {
	return HumaOptions{
		Title:       "My API",
		Version:     "1.0.0",
		Description: "API Documentation",
		DocsPath:    "/docs",
		OpenAPIPath: "/openapi",
	}
}

// ApplyHumaConfig documents related behavior.
func ApplyHumaConfig(cfg *huma.Config, opts HumaOptions) {
	if cfg == nil {
		return
	}

	cfg.Info.Description = opts.Description

	if opts.DisableDocsRoutes {
		cfg.DocsPath = ""
		cfg.OpenAPIPath = ""
		return
	}

	cfg.DocsPath = normalizeDocsPath(opts.DocsPath)
	cfg.OpenAPIPath = normalizeOpenAPIPath(opts.OpenAPIPath)
}

func normalizeDocsPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "/docs"
	}
	if !strings.HasPrefix(trimmed, "/") {
		return "/" + trimmed
	}
	return trimmed
}

func normalizeOpenAPIPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "/openapi"
	}

	trimmed = strings.TrimSuffix(trimmed, ".json")
	trimmed = strings.TrimSuffix(trimmed, ".yaml")
	trimmed = strings.TrimSuffix(trimmed, "/")
	if trimmed == "" {
		return "/openapi"
	}
	if !strings.HasPrefix(trimmed, "/") {
		return "/" + trimmed
	}
	return trimmed
}

// WithRouteParams documents related behavior.
func WithRouteParams(ctx context.Context, params map[string]string) context.Context {
	if len(params) == 0 {
		return ctx
	}
	return context.WithValue(ctx, routeParamsCtxKey{}, params)
}

// RouteParam documents related behavior.
func RouteParam(ctx context.Context, name string) string {
	if ctx == nil || name == "" {
		return ""
	}

	params, ok := ctx.Value(routeParamsCtxKey{}).(map[string]string)
	if !ok {
		return ""
	}
	return params[name]
}
