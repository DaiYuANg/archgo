# httpx Organization Example

This example demonstrates the newer `httpx` organization APIs in one runnable service:

- custom docs paths and docs renderer
- OpenAPI security registration
- global request headers
- group-level tags, security, parameters, summary/description defaults
- group-level external docs and OpenAPI extensions

## Run

From the repo root:

```bash
go run ./httpx/examples/organization
```

## Routes

- `GET /api/health`
- `GET /api/admin/tenants/{id}`

## Docs

- Docs UI: `http://localhost:8080/reference`
- OpenAPI JSON: `http://localhost:8080/spec.json`
- OpenAPI YAML: `http://localhost:8080/spec.yaml`

## What To Look For

The `/api/admin/tenants/{id}` operation is registered through a group and inherits:

- tags: `admin`, `tenants`
- security: `bearerAuth`
- header parameter: `X-Tenant`
- summary prefix: `Admin`
- shared description
- external docs metadata
- extension: `x-owner=platform`

The server also registers a global request header:

- `X-Request-Id`
