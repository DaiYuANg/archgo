# observability

`observability` 提供可选的统一可观测性抽象（日志/追踪/指标）。

[English](./README.md) | Chinese

## 目标

- 让 `authx`、`eventx`、`configx` 的 API 保持稳定。
- 可观测性后端按需开启，不强制绑定。
- 业务代码可自由选择 telemetry 栈。

## 可选后端

- `observability.Nop()`：默认空实现。
- `observability/otel`：OpenTelemetry（trace + metric）。
- `observability/prometheus`：Prometheus（metric + `/metrics` handler）。

## 组合多个后端

```go
otelObs := otelobs.New()
promObs := promobs.New()

obs := observability.Multi(otelObs, promObs)
```

## 接入业务包

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

## 暴露 Prometheus `/metrics`

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
