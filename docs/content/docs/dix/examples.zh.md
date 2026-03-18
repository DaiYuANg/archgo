---
title: 'dix 示例'
linkTitle: 'examples'
description: 'dix 的可运行示例'
weight: 10
---

## dix 示例

这一页汇总 `examples/dix` 里的可运行程序，并说明每个示例主要覆盖的 API 场景。

## 本地运行

从 `examples/dix` 模块执行：

```bash
cd examples/dix
go run ./basic
go run ./runtime_scope
go run ./inspect
```

## 核心示例

| 示例 | 关注点 | 目录 |
| --- | --- | --- |
| `basic` | 不可变 app spec、build/start/stop、health check、`logx` 集成 | [examples/dix/basic](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/basic) |
| `aggregate_params` | 多个 typed dependency 的 provider graph 组合 | [examples/dix/aggregate_params](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/aggregate_params) |
| `build_runtime` | 显式 `Build()` 到 `Runtime` 的流程 | [examples/dix/build_runtime](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/build_runtime) |
| `build_failure` | validation/build 失败路径 | [examples/dix/build_failure](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/build_failure) |

## 高级示例

| 示例 | 关注点 | 目录 |
| --- | --- | --- |
| `advanced_do_bridge` | 显式 `do` bridge setup | [examples/dix/advanced_do_bridge](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/advanced_do_bridge) |
| `named_alias` | named service 与 typed alias 绑定 | [examples/dix/named_alias](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/named_alias) |
| `runtime_scope` | 请求级 runtime scope 与 scoped provider | [examples/dix/runtime_scope](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/runtime_scope) |
| `transient` | transient provider 语义 | [examples/dix/transient](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/transient) |
| `override` | 结构化 override | [examples/dix/override](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/override) |
| `inspect` | runtime inspection 与诊断 | [examples/dix/inspect](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/inspect) |

## 示例：基础应用组装

```go
app := dix.New(
    "basic",
    dix.WithLogger(logger),
    dix.WithModule(
        dix.NewModule("config",
            dix.WithModuleProviders(
                dix.Provider0(func() Config { return Config{Port: 8080} }),
            ),
        ),
    ),
)

if err := app.Validate(); err != nil {
    panic(err)
}

rt, err := app.Build()
if err != nil {
    panic(err)
}
defer func() {
    _, _ = rt.StopWithReport(context.Background())
}()

if err := rt.Start(context.Background()); err != nil {
    panic(err)
}
```

## 示例：Runtime Scope

```go
requestScope := advanced.Scope(rt, "request-42", func(injector do.Injector) {
    advanced.ProvideScopedValue(injector, RequestContext{RequestID: "req-42"})
    advanced.ProvideScoped2(injector, func(cfg AppConfig, req RequestContext) ScopedService {
        return ScopedService{Config: cfg, Request: req}
    })
})

svc, err := advanced.ResolveScopedAs[ScopedService](requestScope)
if err != nil {
    panic(err)
}
fmt.Println(svc.Request.RequestID)
```

## 示例：细粒度 Inspection

```go
provided := advanced.ListProvidedServices(rt)
deps := advanced.ExplainNamedDependencies(rt, "tenant.default")

fmt.Println("provided services:", len(provided))
fmt.Println("tenant graph known:", deps["tenant.default"] != "")
```

如果你只需要某一个诊断视图，优先使用这些细粒度 helper。
`InspectRuntime(...)` 依然方便，但它是更重的聚合路径。
