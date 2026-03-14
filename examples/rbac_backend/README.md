# RBAC Backend Scaffold (fiber + httpx + authx + eventx + observabilityx + bun + fx)

A reusable backend scaffold example with:

- HTTP stack: `httpx` + `fiber` adapter
- DI and lifecycle: `go.uber.org/fx`
- Config loading: `configx` (`.env` + env + defaults)
- Logging: `logx`
- Events: `eventx` (async)
- Observability: `observabilityx` + Prometheus metrics
- AuthN: JWT (HS256)
- AuthZ: authx engine + RBAC tables via bun (`sqlite/mysql/postgres`)

## Run

```bash
go run ./examples/rbac_backend
```

Default address: `:18080`

- health: `http://127.0.0.1:18080/health`
- docs: `http://127.0.0.1:18080/docs`
- openapi: `http://127.0.0.1:18080/openapi.json`
- metrics: `http://127.0.0.1:18080/metrics`
- api base path: `http://127.0.0.1:18080/api/v1`

## Seeded Users

- admin: `alice / admin123`
- user: `bob / user123`

## RBAC Model

Tables:

- `rbac_users`
- `rbac_roles`
- `rbac_permissions`
- `rbac_user_roles`
- `rbac_role_permissions`
- `rbac_books`

Seed permissions:

- admin: `query:book`, `create:book`, `delete:book`
- user: `query:book`

## API Quick Try

Login and get JWT:

```bash
curl -X POST http://127.0.0.1:18080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"admin123"}'
```

Use returned token:

```bash
export TOKEN=<jwt-token>

curl http://127.0.0.1:18080/api/v1/books \
  -H "Authorization: Bearer ${TOKEN}"
```

Create book (admin allowed):

```bash
curl -X POST http://127.0.0.1:18080/api/v1/books \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"title":"New Book","author":"Someone"}'
```

Delete book (admin allowed):

```bash
curl -X DELETE http://127.0.0.1:18080/api/v1/books/1 \
  -H "Authorization: Bearer ${TOKEN}"
```

## Optional Env

- `RBAC_HTTP_ADDR` (default `:18080`)
- `RBAC_BASE_PATH` (default `/api/v1`)
- `RBAC_DOCS_PATH` (default `/docs`)
- `RBAC_OPENAPI_PATH` (default `/openapi.json`)
- `RBAC_METRICS_PATH` (default `/metrics`)
- `RBAC_DB_DRIVER` (default `sqlite`, optional: `mysql`, `postgres`)
- `RBAC_DB_DSN` (default `file:rbac_basic.db?cache=shared`)
- `RBAC_VERSION` (default `0.4.0`)
- `RBAC_JWT_SECRET` (default `change-me-in-production`)
- `RBAC_JWT_ISSUER` (default `arcgo-rbac-example`)
- `RBAC_JWT_EXPIRES_MINUTES` (default `120`)
- `RBAC_EVENT_WORKERS` (default `8`)
- `RBAC_EVENT_PARALLEL` (default `true`)

## Database DSN Examples

- sqlite: `RBAC_DB_DRIVER=sqlite`, `RBAC_DB_DSN=file:rbac_basic.db?cache=shared`
- mysql: `RBAC_DB_DRIVER=mysql`, `RBAC_DB_DSN=user:pass@tcp(127.0.0.1:3306)/rbac?parseTime=true`
- postgres: `RBAC_DB_DRIVER=postgres`, `RBAC_DB_DSN=postgres://user:pass@127.0.0.1:5432/rbac?sslmode=disable`
