# httpx Auth Example

This example demonstrates authentication-related OpenAPI configuration using the
current `httpx` API surface.

It shows:

- `WithSecurity(...)` for security scheme registration
- `RegisterGlobalHeader(...)` for request correlation headers
- group-level default security and tags
- typed header binding for `Authorization`, `X-API-Key`, and `X-Request-Id`

## Run

From the repo root:

```bash
go run ./httpx/examples/auth
```

## Routes

- `GET /api/health`
- `GET /api/secure/profile`

## Docs

- Docs UI: `http://localhost:8080/docs`
- OpenAPI JSON: `http://localhost:8080/openapi.json`

## Try It

Bearer example:

```bash
curl http://localhost:8080/api/secure/profile ^
  -H "Authorization: Bearer demo-token" ^
  -H "X-Request-Id: req-1"
```

API key example:

```bash
curl http://localhost:8080/api/secure/profile ^
  -H "X-API-Key: demo-key" ^
  -H "X-Request-Id: req-2"
```

## What To Look For

The `GET /api/secure/profile` operation is documented with:

- security alternatives: `BearerAuth` and `ApiKeyAuth`
- tag: `auth`
- global header: `X-Request-Id`
