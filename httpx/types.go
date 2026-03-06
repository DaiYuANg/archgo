package httpx

import (
	"context"
	"net/http"

	"github.com/DaiYuANg/arcgo/httpx/adapter"
	"github.com/danielgtaylor/huma/v2"
)

// HTTPMethod documents related behavior.
const (
	MethodGet     = http.MethodGet
	MethodPost    = http.MethodPost
	MethodPut     = http.MethodPut
	MethodDelete  = http.MethodDelete
	MethodPatch   = http.MethodPatch
	MethodHead    = http.MethodHead
	MethodOptions = http.MethodOptions
)

// RouteInfo documents related behavior.
type RouteInfo struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	HandlerName string   `json:"handler_name"`
	Comment     string   `json:"comment,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// String returns related data.
func (r RouteInfo) String() string {
	return r.Method + " " + r.Path + " -> " + r.HandlerName
}

// TypedHandler documents related behavior.
// Note.
type TypedHandler[I, O any] func(ctx context.Context, input *I) (*O, error)

// OperationOption documents related behavior.
type OperationOption func(*huma.Operation)

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

// ToAdapterHumaOptions converts related values.
func ToAdapterHumaOptions(opts HumaOptions) adapter.HumaOptions {
	return adapter.HumaOptions(opts)
}
