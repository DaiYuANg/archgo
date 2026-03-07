package adapter

import (
	"encoding/json"
	"net/http"
	"path"
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
	if !isZeroHumaOptions(c.current) {
		c.stale = append(c.stale, c.current)
	}
	c.current = next
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
	return requestPath == normalizeOpenAPIPath(opts.OpenAPIPath)+suffix
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

func renderDocsHTML(openAPI *huma.OpenAPI, opts HumaOptions) []byte {
	title := "API Reference"
	if openAPI != nil && openAPI.Info != nil && openAPI.Info.Title != "" {
		title = openAPI.Info.Title + " Reference"
	}

	openAPIPath := normalizeOpenAPIPath(opts.OpenAPIPath)
	if prefix := openAPIPrefix(openAPI); prefix != "" {
		openAPIPath = path.Join(prefix, openAPIPath)
	}

	renderer := opts.DocsRenderer
	if renderer == "" {
		renderer = huma.DocsRendererStoplightElements
	}

	switch renderer {
	case huma.DocsRendererScalar:
		return []byte(`<!doctype html>
<html lang="en">
  <head>
    <title>` + title + `</title>
    <meta charset="utf-8">
    <meta content="width=device-width,initial-scale=1" name="viewport">
  </head>
  <body>
    <script data-url="` + openAPIPath + `.yaml" id="api-reference"></script>
    <script>let apiReference = document.getElementById("api-reference")</script>
    <script src="https://unpkg.com/@scalar/api-reference@1.44.18/dist/browser/standalone.js"></script>
  </body>
</html>`)
	case huma.DocsRendererSwaggerUI:
		return []byte(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>` + title + `</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.31.0/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.31.0/swagger-ui-bundle.js" crossorigin></script>
    <script>
      window.onload = () => {
        window.ui = SwaggerUIBundle({
          url: '` + openAPIPath + `.json',
          dom_id: '#swagger-ui',
        });
      };
    </script>
  </body>
</html>`)
	default:
		return []byte(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="referrer" content="same-origin" />
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no" />
    <title>` + title + `</title>
    <link href="https://unpkg.com/@stoplight/elements@9.0.0/styles.min.css" rel="stylesheet" />
    <script src="https://unpkg.com/@stoplight/elements@9.0.0/web-components.min.js" crossorigin="anonymous"></script>
  </head>
  <body style="height: 100vh;">
    <elements-api apiDescriptionUrl="` + openAPIPath + `.yaml" router="hash" layout="sidebar" tryItCredentialsPolicy="same-origin" />
  </body>
</html>`)
	}
}

func openAPIPrefix(openAPI *huma.OpenAPI) string {
	if openAPI == nil || len(openAPI.Servers) == 0 || openAPI.Servers[0] == nil {
		return ""
	}
	serverURL := strings.TrimSpace(openAPI.Servers[0].URL)
	if serverURL == "" {
		return ""
	}
	if strings.HasPrefix(serverURL, "http://") || strings.HasPrefix(serverURL, "https://") {
		parts := strings.SplitN(serverURL, "/", 4)
		if len(parts) < 4 {
			return ""
		}
		return "/" + strings.Trim(parts[3], "/")
	}
	if !strings.HasPrefix(serverURL, "/") {
		serverURL = "/" + serverURL
	}
	return strings.TrimRight(serverURL, "/")
}

func isZeroHumaOptions(opts HumaOptions) bool {
	return opts.Title == "" &&
		opts.Version == "" &&
		opts.Description == "" &&
		opts.DocsPath == "" &&
		opts.OpenAPIPath == "" &&
		opts.SchemasPath == "" &&
		opts.DocsRenderer == "" &&
		!opts.DisableDocsRoutes
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
