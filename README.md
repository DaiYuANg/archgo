# arcgo

`arcgo` is a modular Go toolkit for backend infrastructure.
It is package-oriented, supports incremental adoption, and allows inter-package composition.

English | [Chinese](./README_ZH.md)

> **Documentation**: Browse the [Hugo documentation site](./docs/) for a unified documentation experience.

## Packages

| Package | Description |
| --- | --- |
| `authx` | Opinionated security layer on Authboss + Casbin |
| `collectionx` | Generic collections and concurrent-safe structures |
| `configx` | Layered config loading and validation |
| `eventx` | In-memory typed event bus |
| `httpx` | Typed HTTP routing across adapters |
| `logx` | Structured logging with zerolog + slog bridge |
| `observabilityx` | Optional observability facade with OTel/Prometheus adapters |

## How To Choose Quickly

- You need container/data helpers: start with `collectionx`.
- You need opinionated auth/authz abstraction on Authboss + Casbin: start with `authx`.
- You need config from `.env` + file + env vars: start with `configx`.
- You need process-local pub/sub with typed payloads: start with `eventx`.
- You need unified typed HTTP handlers across frameworks: start with `httpx`.
- You need structured logs and rotation: start with `logx`.
- You need optional telemetry abstraction (OTel/Prometheus): start with `observabilityx`.

## Typical Stack Combinations

- API service baseline: `httpx + configx + logx`
- Domain-events in a monolith: `eventx + logx`
- Data-heavy utility/internal libs: `collectionx + configx`

## Build & QA

```bash
go tool task fmt
go tool task lint
go tool task test
go tool task check
```

## Docs Utility (Cross-Platform)

```bash
go run ./scripts/deploy-docs help
go run ./scripts/deploy-docs sync
go run ./scripts/deploy-docs build
go run ./scripts/deploy-docs serve
go run ./scripts/deploy-docs deploy
# optional:
# DOCS_REMOTE=origin DOCS_BRANCH=gh-pages go run ./scripts/deploy-docs deploy
```

## clientx Examples

```bash
go run ./clientx/examples/edge_http
go run ./clientx/examples/internal_rpc_tcp
go run ./clientx/examples/low_latency_udp
```

## Git Pre-Commit Hook

This repo uses `lefthook` (managed via `go tool`).

Install hooks (run once per clone):

```bash
go tool task git:hooks:install
```

Run hooks manually:

```bash
go tool task git:hooks:run
```

The `pre-commit` hook runs:

- `go tool task fmt`
- `go tool task lint`
```

## Notes

- Code comments are English-only.
- All documentation is available at the Hugo documentation site.
