package adapter

import (
	"path"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

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
