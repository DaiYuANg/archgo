# observability

`observability` provides an optional, unified facade for logging/tracing/metrics.

[Chinese](./README_ZH.md)

## Why

- Keep `authx`, `eventx`, `configx` APIs stable.
- Make observability backend optional.
- Avoid forcing business code to use one telemetry stack.

## Backends

- `observability.Nop()` - default no-op backend.
- `observability/otel` - OpenTelemetry backend (trace + metric).
- `observability/prometheus` - Prometheus backend (metrics + `/metrics` handler).

## Compose Multiple Backends

```go
otelObs := otelobs.New()
promObs := promobs.New()

obs := observability.Multi(otelObs, promObs)
```

## Wire Into Packages

```go
manager, _ := authx.NewManager(
    authx.WithObservability(obs),
    authx.WithProvider(provider),
)

bus := eventx.New(
    eventx.WithObservability(obs),
)

var cfg AppConfig
_ = configx.Load(&cfg,
    configx.WithObservability(obs),
    configx.WithFiles("config.yaml"),
)
```

## Prometheus Metrics Endpoint

```go
promObs := promobs.New()

metricsServer := httpx.NewServer(
    httpx.WithAdapter(std.New()),
    httpx.WithOpenAPIDocs(false),
)
metricsServer.Adapter().Handle(httpx.MethodGet, "/metrics", func(
    ctx context.Context,
    w http.ResponseWriter,
    r *http.Request,
) error {
    promObs.Handler().ServeHTTP(w, r)
    return nil
})
```
