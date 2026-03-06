# httpx quickstart

This is a complete 0-to-1 runnable example for `httpx` using the `std` adapter.

What it demonstrates:

- Server bootstrap with typed routing
- Base path and route groups
- Path param binding
- JSON body binding
- Validator-based request validation
- OpenAPI metadata and docs endpoint

## Run

From repository root:

```bash
go run ./httpx/examples/quickstart
```

Server starts on `:8080`.

## Try it

### Health check

```bash
curl http://localhost:8080/api/health
```

### Create user (valid)

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H \"Content-Type: application/json\" \
  -d '{\"name\":\"Alice\",\"email\":\"alice@example.com\"}'
```

### Create user (invalid)

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H \"Content-Type: application/json\" \
  -d '{\"name\":\"A\",\"email\":\"bad\"}'
```

### Get user by path parameter

```bash
curl http://localhost:8080/api/v1/users/42
```

### OpenAPI docs

- Swagger UI: `http://localhost:8080/docs`
- OpenAPI: `http://localhost:8080/openapi.json`

