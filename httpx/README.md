# httpx

`httpx` is a lightweight HTTP service organization layer built on top of Huma.

[Chinese](./README_ZH.md)

## What You Get

- Unified typed route registration across adapters (`Get`, `Post`, `Put`, `Patch`, `Delete`, ...)
- Adapter-based runtime integration (`std`, `gin`, `echo`, `fiber`)
- First-class OpenAPI and docs control
- Direct Huma escape hatches (`HumaAPI`, `OpenAPI`, `ConfigureOpenAPI`)
- Group-level Huma middleware and operation customization
- Optional request validation via `go-playground/validator`
- Route introspection APIs for tests and diagnostics

## Positioning

`httpx` is not a heavy web framework and it is not trying to replace Huma.
It provides a stable server/group/endpoint API surface while preserving direct access to Huma's advanced features.

Responsibilities are split as follows:

- `Huma`: typed operations, schemas, OpenAPI, docs, middleware model
- `adapter/*`: runtime, router integration, native middleware ecosystem
- `httpx`: unified service organization API and Huma capability exposure

## Minimal Setup

```go
package main

import (
    "context"

    "github.com/DaiYuANg/arcgo/httpx"
    "github.com/DaiYuANg/arcgo/httpx/adapter/std"
    "github.com/go-chi/chi/v5/middleware"
)

type HealthOutput struct {
    Body struct {
        Status string `json:"status"`
    }
}

func main() {
    a := std.New()
    a.Router().Use(middleware.Logger, middleware.Recoverer)

    s := httpx.NewServer(
        httpx.WithAdapter(a),
        httpx.WithBasePath("/api"),
        httpx.WithOpenAPIInfo("My API", "1.0.0", "Service API"),
    )

    _ = httpx.Get(s, "/health", func(ctx context.Context, input *struct{}) (*HealthOutput, error) {
        out := &HealthOutput{}
        out.Body.Status = "ok"
        return out, nil
    })

    _ = s.ListenAndServe(":8080")
}
```

## Core APIs

### Server

- `NewServer(...)`
- `WithAdapter(...)`
- `WithBasePath(...)`
- `WithValidation()` / `WithValidator(...)`
- `WithPanicRecover(...)`
- `WithAccessLog(...)`
- `HumaAPI()`
- `OpenAPI()`
- `ConfigureOpenAPI(...)`
- `PatchOpenAPI(...)`
- `UseHumaMiddleware(...)`

### Docs / OpenAPI

Construction-time docs config:

```go
s := httpx.NewServer(
    httpx.WithDocs(httpx.DocsOptions{
        Enabled:     true,
        DocsPath:    "/reference",
        OpenAPIPath: "/spec",
        SchemasPath: "/schemas",
        Renderer:    httpx.DocsRendererScalar,
    }),
)
```

Runtime docs config:

```go
s.ConfigureDocs(func(d *httpx.DocsOptions) {
    d.DocsPath = "/docs/internal"
    d.OpenAPIPath = "/openapi/internal"
})
```

OpenAPI patching:

```go
s.ConfigureOpenAPI(func(doc *huma.OpenAPI) {
    doc.Tags = append(doc.Tags, &huma.Tag{Name: "internal"})
})
```

Notes:

- `WithOpenAPIInfo(...)` and `WithOpenAPIDocs(...)` still work.
- `ConfigureDocs(...)` now updates adapter-managed docs routes as well.
- Supported built-in renderers:
  - `httpx.DocsRendererStoplightElements`
  - `httpx.DocsRendererScalar`
  - `httpx.DocsRendererSwaggerUI`

### Security / Components / Global Parameters

```go
s := httpx.NewServer(
    httpx.WithSecurity(httpx.SecurityOptions{
        Schemes: map[string]*huma.SecurityScheme{
            "bearerAuth": {
                Type:   "http",
                Scheme: "bearer",
            },
        },
        Requirements: []map[string][]string{
            {"bearerAuth": {}},
        },
    }),
)

s.RegisterComponentParameter("Locale", &huma.Param{
    Name: "locale",
    In:   "query",
    Schema: &huma.Schema{Type: "string"},
})

s.RegisterGlobalHeader(&huma.Param{
    Name:   "X-Request-Id",
    In:     "header",
    Schema: &huma.Schema{Type: "string"},
})
```

Available APIs:

- `RegisterSecurityScheme(...)`
- `SetDefaultSecurity(...)`
- `RegisterComponentParameter(...)`
- `RegisterComponentHeader(...)`
- `RegisterGlobalParameter(...)`
- `RegisterGlobalHeader(...)`
- `AddTag(...)`

### Groups

Basic grouping:

```go
api := s.Group("/v1")
_ = httpx.GroupGet(api, "/users/{id}", getUser)
_ = httpx.GroupPost(api, "/users", createUser)
```

Group-level Huma capabilities:

```go
api := s.Group("/admin")
api.UseHumaMiddleware(authMiddleware)
api.DefaultTags("admin")
api.DefaultSecurity(map[string][]string{"bearerAuth": {}})
api.DefaultParameters(&huma.Param{
    Name:   "X-Tenant",
    In:     "header",
    Schema: &huma.Schema{Type: "string"},
})
api.DefaultSummaryPrefix("Admin")
api.DefaultDescription("Administrative APIs")
```

Available group APIs:

- `HumaGroup()`
- `UseHumaMiddleware(...)`
- `UseOperationModifier(...)`
- `UseSimpleOperationModifier(...)`
- `UseResponseTransformer(...)`
- `DefaultTags(...)`
- `DefaultSecurity(...)`
- `DefaultParameters(...)`
- `DefaultSummaryPrefix(...)`
- `DefaultDescription(...)`

## Typed Input Patterns

```go
type GetUserInput struct {
    ID int `path:"id"`
}

type ListUsersInput struct {
    Page int `query:"page"`
    Size int `query:"size"`
}

type SecureInput struct {
    RequestID string `header:"X-Request-Id"`
}

type CreateUserInput struct {
    Body struct {
        Name  string `json:"name" validate:"required,min=2,max=64"`
        Email string `json:"email" validate:"required,email"`
    }
}
```

## Middleware Model

`httpx` uses a two-layer middleware model:

- Adapter-native middleware: register directly on the adapter router/engine/app
- Huma middleware: register via `Server.UseHumaMiddleware(...)` or `Group.UseHumaMiddleware(...)`

Adapter middleware should stay adapter-native:

- `std`: `adapter.Router().Use(...)`
- `gin`: `adapter.Router().Use(...)`
- `echo`: `adapter.Router().Use(...)`
- `fiber`: `adapter.Router().Use(...)`

Typed-handler operational controls live at the `httpx` layer:

- `WithPanicRecover(...)` controls panic recovery for typed `httpx` handlers
- `WithAccessLog(...)` controls request logging through the server logger

Runtime listener settings such as read/write/idle timeouts and max header bytes are adapter concerns and should be configured on the adapter or underlying server library, not through `httpx/options.ServerOptions`.

## Logging

`httpx` logger behavior is intentionally split across layers:

- `httpx.WithLogger(...)` configures the `httpx.Server` logger
- adapter loggers configure bridge-layer errors emitted by `adapter/std`, `adapter/gin`, `adapter/echo`, and `adapter/fiber`
- framework-native loggers and logging middleware remain framework concerns

In practice this means:

- use `httpx.WithLogger(...)` for `httpx` route/access-log/route-registration output
- configure the adapter logger explicitly when you want adapter bridge errors to use the same logger
- continue configuring `chi` / `gin` / `echo` / `fiber` logging middleware on the adapter router or engine

`httpx` does not currently promise to fully replace framework-native loggers.

## Adapter Construction

Listener and bridge-layer configuration belongs to the adapter, not `httpx.ServerOptions`.

For `net/http`-based adapters such as `std`, `gin`, and `echo`, use construction-time adapter options:

```go
stdAdapter := std.NewWithOptions(std.Options{
    Logger: slogLogger,
    Server: std.ServerOptions{
        ReadTimeout:     15 * time.Second,
        WriteTimeout:    15 * time.Second,
        IdleTimeout:     60 * time.Second,
        ShutdownTimeout: 5 * time.Second,
        MaxHeaderBytes:  1 << 20,
    },
})
```

For `fiber`, timeout settings belong to the app config used when the adapter creates the app:

```go
fiberAdapter := fiber.NewWithOptions(nil, fiber.Options{
    Logger: slogLogger,
    App: fiber.AppOptions{
        ReadTimeout:     15 * time.Second,
        WriteTimeout:    15 * time.Second,
        IdleTimeout:     60 * time.Second,
        ShutdownTimeout: 5 * time.Second,
    },
})
```

If you pass an already-created framework object, that framework object's own config remains authoritative.

## Introspection APIs

- `GetRoutes()`
- `GetRoutesByMethod(method)`
- `GetRoutesByPath(prefix)`
- `HasRoute(method, path)`
- `RouteCount()`

## Option Builder

You can build server options through `httpx/options`:

```go
opts := options.DefaultServerOptions()
opts.BasePath = "/api"
opts.HumaTitle = "Arc API"
opts.DocsPath = "/reference"
opts.DocsRenderer = httpx.DocsRendererSwaggerUI
opts.EnablePanicRecover = true
opts.EnableAccessLog = true

s := httpx.NewServer(append(opts.Build(), httpx.WithAdapter(a))...)
```

Use adapter construction options separately for listener timeout and adapter logger configuration.

## Testing Pattern

```go
req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
rec := httptest.NewRecorder()
s.ServeHTTP(rec, req)

if rec.Code != http.StatusOK {
    t.Fatal(rec.Code)
}
```

## FAQ

### Do I have to use Huma-style input structs?

Yes for typed route handlers in this package.

### Can I still access raw Huma APIs?

Yes. Use `HumaAPI()`, `OpenAPI()`, and `HumaGroup()`.

### Should `httpx` wrap adapter middleware too?

No. Keep adapter-native middleware on the adapter itself and use `httpx` for Huma-side middleware and service organization.

## Examples

- Quickstart: `go run ./httpx/examples/quickstart`
  - Minimal typed routes + validation + base path
- Auth: `go run ./httpx/examples/auth`
  - Security schemes, global headers, and typed auth header binding
  - See [`httpx/examples/auth/README.md`](./examples/auth/README.md)
- Organization: `go run ./httpx/examples/organization`
  - Docs paths, security, global headers, and group defaults
  - See [`httpx/examples/organization/README.md`](./examples/organization/README.md)
