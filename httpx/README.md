# httpx

`httpx` is a typed HTTP abstraction layer for multiple Go web frameworks.

[Chinese](./README_ZH.md)

## What You Get

- Unified typed route registration (`Get`, `Post`, `Put`, `Patch`, `Delete`, ...)
- Adapter-based framework integration (`std`, `gin`, `fiber`, `echo`)
- Huma request/response handling and OpenAPI integration by default
- Optional global request validation via `go-playground/validator`
- Route grouping and route introspection APIs

## Core Architecture

- `Server`: central routing/runtime object
- `adapter/*`: framework bridges
- Typed route APIs: `httpx.Get/Post/...`
- Group route APIs: `httpx.GroupGet/...`

## 0-to-1 Runnable Example

- Quickstart directory: [httpx/examples/quickstart](./examples/quickstart)
- Run from repo root:

```bash
go run ./httpx/examples/quickstart
```

## Minimal Setup (std/chi)

```go
a := std.New()
a.Router().Use(middleware.Logger, middleware.Recoverer)

s := httpx.NewServer(
    httpx.WithAdapter(a),
    httpx.WithBasePath("/api"),
    httpx.WithOpenAPIInfo("My API", "1.0.0", "Service API"),
)

_ = httpx.Get(s, "/health", func(ctx context.Context, input *struct{}) (*HealthOutput, error) {
    return &HealthOutput{Body: struct{ Status string `json:"status"` }{Status: "ok"}}, nil
})

_ = s.ListenAndServe(":8080")
```

## Typed Input Patterns

### 1) Path params

```go
type GetUserInput struct {
    ID int `path:"id"`
}

_ = httpx.Get(s, "/users/{id}", func(ctx context.Context, in *GetUserInput) (*UserOutput, error) {
    // in.ID from path
    return out, nil
})
```

### 2) Query params

```go
type ListUsersInput struct {
    Page int `query:"page"`
    Size int `query:"size"`
}
```

### 3) Headers

```go
type SecureInput struct {
    RequestID string `header:"X-Request-Id"`
}
```

### 4) JSON body

```go
type CreateUserInput struct {
    Body struct {
        Name  string `json:"name" validate:"required,min=2,max=64"`
        Email string `json:"email" validate:"required,email"`
    }
}
```

## Route Groups

```go
api := s.Group("/v1")
_ = httpx.GroupGet(api, "/users/{id}", getUserHandler)
_ = httpx.GroupPost(api, "/users", createUserHandler)
```

## Validation Modes

### Built-in validator

```go
s := httpx.NewServer(
    httpx.WithAdapter(a),
    httpx.WithValidation(),
)
```

### Custom validator instance

```go
v := validator.New(validator.WithRequiredStructEnabled())
s := httpx.NewServer(
    httpx.WithAdapter(a),
    httpx.WithValidator(v),
)
```

Validation failures are converted to HTTP 400 with structured error output through Huma.

## OpenAPI / Docs Control

```go
s := httpx.NewServer(
    httpx.WithAdapter(a),
    httpx.WithOpenAPIInfo("My API", "1.0.0", "Public API"),
    httpx.WithOpenAPIDocs(false), // disable /docs and /openapi.* in production
)
```

## Error Mapping

Handlers can return standard errors or `httpx.Error`:

```go
return nil, httpx.NewError(http.StatusForbidden, "forbidden")
```

`httpx.Error` is converted to Huma HTTP errors with your status code.

## Framework Middleware Integration

Register middleware directly through native adapter objects:

- `std`: `adapter.Router().Use(...)`
- `gin`: `adapter.Router().Use(...)`
- `fiber`: `adapter.Router().Use(...)`
- `echo`: `adapter.Router().Use(...)`

## Server Introspection APIs

- `GetRoutes()`
- `GetRoutesByMethod(method)`
- `GetRoutesByPath(prefix)`
- `HasRoute(method, path)`
- `RouteCount()`

Useful for runtime diagnostics or test assertions.

## Testing Patterns

```go
req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
rec := httptest.NewRecorder()
s.ServeHTTP(rec, req)

if rec.Code != http.StatusOK { t.Fatal(rec.Code) }
```

Recommended test setup:

- Use `std` adapter for simple in-process tests.
- Assert route registration with `HasRoute` and `GetRoutes`.
- Test both validation-success and validation-failure paths.

## Migration Notes

- Legacy endpoint wrappers were removed.
- Use typed route APIs directly (`Get/Post/Group*`).
- Use Huma-style input structs consistently.

## Advanced Option Builder

You can build server options through `httpx/options` package when assembling config-driven apps.

```go
opts := options.DefaultServerOptions()
opts.BasePath = "/api"
opts.EnableValidation = true

s := httpx.NewServer(append(opts.Build(), httpx.WithAdapter(a))...)
```

## FAQ

### Do I have to use Huma-style input structs?

Yes for typed route handlers in this package.  
`httpx` standardizes on Huma input/output modeling to keep behavior consistent across adapters.

### Can I disable OpenAPI docs in production?

Yes. Use `WithOpenAPIDocs(false)` to disable `/docs` and `/openapi.*` endpoints.

### Should I use adapter middleware wrappers from `httpx`?

No. Register middleware natively on each framework adapter engine/app/router.

### How do I access path/query/header parameters?

Declare them in input struct tags (`path`, `query`, `header`).  
Huma parsing binds values into the typed input object automatically.

## Troubleshooting

### Route returns 400 unexpectedly

Common causes:

- Input tag mismatch (`path:\"id\"` but route uses `{userId}`).
- Validation failure (`validate` tags fail).
- Request body shape does not match `Body` struct.

### Docs endpoints not visible

Check:

- `WithOpenAPIDocs(true)` is enabled.
- Reverse proxy/path prefix settings are correct.
- Adapter has Huma enabled (default server path does this automatically).

### Framework middleware not firing

Middleware must be attached to the framework adapter itself (for example `ginAdapter.Router().Use(...)`).  
Adding middleware in unrelated HTTP stack layers will not affect adapter route chain.

### Different adapters behave differently for edge HTTP semantics

Some low-level framework behaviors differ (error propagation, request body reuse, context internals).  
Keep handler logic adapter-agnostic and avoid depending on framework-specific side effects.

## Anti-Patterns

- Mixing manual framework binding with typed Huma input for the same route.
- Returning ad-hoc response types without stable schema for public APIs.
- Treating validation as optional for externally-facing write endpoints.
- Building a single giant route file instead of grouped bounded contexts.
- Coupling business logic directly to adapter-specific request objects.
