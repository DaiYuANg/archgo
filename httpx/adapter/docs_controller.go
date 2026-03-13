package adapter

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/danielgtaylor/huma/v2"
)

var schemaRefPattern = regexp.MustCompile(`"#/components/schemas/([^"]+)"`)

// HumaOptionsConfigurer updates adapter-managed Huma/docs behavior after construction.
type HumaOptionsConfigurer interface {
	ConfigureHumaOptions(opts HumaOptions)
}

// DocsController handles docs/openapi/schema routes outside the router's static registration.
type DocsController struct {
	mu      sync.RWMutex
	api     huma.API
	current HumaOptions
	stale   []HumaOptions
}

// NewDocsController creates a docs controller for adapter-managed docs routes.
func NewDocsController(api huma.API, opts HumaOptions) *DocsController {
	return &DocsController{
		api:     api,
		current: MergeHumaOptions(opts),
	}
}

// Configure updates the active docs config and invalidates previous docs routes.
func (c *DocsController) Configure(opts HumaOptions) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	next := MergeHumaOptions(opts)
	if humaOptionsEqual(c.current, next) {
		return
	}
	if !pathsEqual(c.current, next) {
		c.stale = append(c.stale, c.current)
	}
	c.current = next
}

// pathsEqual checks if the path-related fields are equal between two HumaOptions.
func pathsEqual(a, b HumaOptions) bool {
	return a.DocsPath == b.DocsPath &&
		normalizeOpenAPIPath(a.OpenAPIPath) == normalizeOpenAPIPath(b.OpenAPIPath) &&
		a.SchemasPath == b.SchemasPath
}

// ServeHTTP handles docs/OpenAPI/schema requests and reports whether it wrote a response.
func (c *DocsController) ServeHTTP(w http.ResponseWriter, r *http.Request) bool {
	if c == nil || r == nil || w == nil {
		return false
	}
	if r.Method != http.MethodGet {
		return false
	}

	c.mu.RLock()
	current := c.current
	stale := append([]HumaOptions(nil), c.stale...)
	c.mu.RUnlock()

	for _, opts := range stale {
		if matchDocsRoute(opts, r.URL.Path) {
			http.NotFound(w, r)
			return true
		}
	}

	if !matchDocsRoute(current, r.URL.Path) {
		return false
	}
	if current.DisableDocsRoutes {
		http.NotFound(w, r)
		return true
	}

	openAPI := c.api.OpenAPI()
	switch {
	case isDocsPath(current, r.URL.Path):
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(renderDocsHTML(openAPI, current))
		return true
	case isOpenAPIPath(current, r.URL.Path, ""):
		w.Header().Set("Content-Type", "application/openapi+json")
		body, _ := json.Marshal(openAPI)
		_, _ = w.Write(body)
		return true
	case isOpenAPIPath(current, r.URL.Path, ".json"):
		w.Header().Set("Content-Type", "application/openapi+json")
		body, _ := json.Marshal(openAPI)
		_, _ = w.Write(body)
		return true
	case isOpenAPIPath(current, r.URL.Path, "-3.0.json"):
		w.Header().Set("Content-Type", "application/openapi+json")
		body, _ := openAPI.Downgrade()
		_, _ = w.Write(body)
		return true
	case isOpenAPIPath(current, r.URL.Path, ".yaml"):
		w.Header().Set("Content-Type", "application/openapi+yaml")
		body, _ := openAPI.YAML()
		_, _ = w.Write(body)
		return true
	case isOpenAPIPath(current, r.URL.Path, "-3.0.yaml"):
		w.Header().Set("Content-Type", "application/openapi+yaml")
		body, _ := openAPI.DowngradeYAML()
		_, _ = w.Write(body)
		return true
	case isSchemaPath(current, r.URL.Path):
		w.Header().Set("Content-Type", "application/json")
		schemaName := strings.TrimPrefix(r.URL.Path, normalizeSchemasPath(current.SchemasPath)+"/")
		schemaName = strings.TrimSuffix(schemaName, ".json")
		var body []byte
		if openAPI.Components != nil {
			body, _ = json.Marshal(openAPI.Components.Schemas.Map()[schemaName])
		}
		body = schemaRefPattern.ReplaceAll(body, []byte(`"`+normalizeSchemasPath(current.SchemasPath)+`/$1.json"`))
		_, _ = w.Write(body)
		return true
	default:
		return false
	}
}

func matchDocsRoute(opts HumaOptions, requestPath string) bool {
	return isDocsPath(opts, requestPath) ||
		isOpenAPIPath(opts, requestPath, "") ||
		isOpenAPIPath(opts, requestPath, ".json") ||
		isOpenAPIPath(opts, requestPath, ".yaml") ||
		isOpenAPIPath(opts, requestPath, "-3.0.json") ||
		isOpenAPIPath(opts, requestPath, "-3.0.yaml") ||
		isSchemaPath(opts, requestPath)
}

func isDocsPath(opts HumaOptions, requestPath string) bool {
	if opts.DisableDocsRoutes {
		return matchAnyDocsPath(opts, requestPath)
	}
	return requestPath == normalizeDocsPath(opts.DocsPath)
}

func isOpenAPIPath(opts HumaOptions, requestPath, suffix string) bool {
	normalizedPath := normalizeOpenAPIPath(opts.OpenAPIPath)
	normalizedRequest := normalizeOpenAPIPath(requestPath)
	return normalizedRequest+suffix == normalizedPath+suffix
}

func isSchemaPath(opts HumaOptions, requestPath string) bool {
	prefix := normalizeSchemasPath(opts.SchemasPath) + "/"
	return strings.HasPrefix(requestPath, prefix)
}

func matchAnyDocsPath(opts HumaOptions, requestPath string) bool {
	return requestPath == normalizeDocsPath(opts.DocsPath) ||
		strings.HasPrefix(requestPath, normalizeSchemasPath(opts.SchemasPath)+"/") ||
		strings.HasPrefix(requestPath, normalizeOpenAPIPath(opts.OpenAPIPath))
}

func humaOptionsEqual(a, b HumaOptions) bool {
	return a.Title == b.Title &&
		a.Version == b.Version &&
		a.Description == b.Description &&
		a.DocsPath == b.DocsPath &&
		a.OpenAPIPath == b.OpenAPIPath &&
		a.SchemasPath == b.SchemasPath &&
		a.DocsRenderer == b.DocsRenderer &&
		a.DisableDocsRoutes == b.DisableDocsRoutes
}
