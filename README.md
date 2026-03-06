# arcgo

`arcgo` is a modular Go toolkit for backend infrastructure.  
It is organized by independent packages so you can adopt only what you need.

English | [Chinese](./README_ZH.md)

## Package Guides

| Package | What it solves | English | Chinese | Runnable Quickstart |
| --- | --- | --- | --- | --- |
| `authx` | Opinionated security layer on Authboss + Casbin | [authx/README.md](./authx/README.md) | [authx/README_ZH.md](./authx/README_ZH.md) | [authx/examples/quickstart](./authx/examples/quickstart) |
| `collectionx` | Generic collections and concurrent-safe structures | [collectionx/README.md](./collectionx/README.md) | [collectionx/README_ZH.md](./collectionx/README_ZH.md) | [collectionx/examples/quickstart](./collectionx/examples/quickstart) |
| `configx` | Layered config loading and validation | [configx/README.md](./configx/README.md) | [configx/README_ZH.md](./configx/README_ZH.md) | - |
| `eventx` | In-memory typed event bus | [eventx/README.md](./eventx/README.md) | [eventx/README_ZH.md](./eventx/README_ZH.md) | - |
| `httpx` | Typed HTTP routing across adapters | [httpx/README.md](./httpx/README.md) | [httpx/README_ZH.md](./httpx/README_ZH.md) | [httpx/examples/quickstart](./httpx/examples/quickstart) |
| `logx` | Structured logging with zerolog + slog bridge | [logx/README.md](./logx/README.md) | [logx/README_ZH.md](./logx/README_ZH.md) | - |
| `observability` | Optional observability facade with OTel/Prometheus adapters | [observability/README.md](./observability/README.md) | [observability/README_ZH.md](./observability/README_ZH.md) | [observability/examples/multi](./observability/examples/multi) |

## How To Choose Quickly

- You need container/data helpers: start with `collectionx`.
- You need opinionated auth/authz abstraction on Authboss + Casbin: start with `authx`.
- You need config from `.env` + file + env vars: start with `configx`.
- You need process-local pub/sub with typed payloads: start with `eventx`.
- You need unified typed HTTP handlers across frameworks: start with `httpx`.
- You need structured logs and rotation: start with `logx`.
- You need optional telemetry abstraction (OTel/Prometheus): start with `observability`.

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

## Notes

- Code comments are now English-only.
- Chinese docs are kept as `README_ZH.md` files per package.
