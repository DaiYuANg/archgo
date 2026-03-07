# Endpoint Pattern Example

This example shows the optional `Endpoint` helper layer in `httpx`.

It is useful when you want to split a service into business-oriented route modules
such as `HealthEndpoint`, `UserEndpoint`, and `OrderEndpoint`, while still using
the same typed route APIs under the hood.

## Run

```bash
go run ./httpx/examples/endpoint
```

## What It Demonstrates

- organizing routes by endpoint struct
- registering multiple endpoints with `RegisterOnly(...)`
- using typed route APIs inside each endpoint
- mixing top-level routes and grouped routes

## Endpoint Shape

```go
type UserEndpoint struct {
    httpx.BaseEndpoint
}

func (e *UserEndpoint) RegisterRoutes(server *httpx.Server) {
    api := server.Group("/api/v1/users")

    httpx.MustGroupGet(api, "", listUsersHandler)
    httpx.MustGroupGet(api, "/{id}", getUserHandler)
    httpx.MustGroupPost(api, "", createUserHandler)
}
```

## Registration

```go
server.RegisterOnly(
    &HealthEndpoint{},
    &UserEndpoint{},
    &OrderEndpoint{},
)
```

## API Endpoints

- `GET /health`
- `GET /api/v1/users`
- `GET /api/v1/users/{id}`
- `POST /api/v1/users`
- `POST /api/v1/orders`

## Docs

- Docs UI: `http://localhost:8080/docs`
- OpenAPI JSON: `http://localhost:8080/openapi.json`
