---
title: 'dix examples'
linkTitle: 'examples'
description: 'Runnable examples for dix'
weight: 10
---

## dix Examples

This page collects the runnable `examples/dix` programs and maps them to the API surface they demonstrate.

## Run Locally

Run from the `examples/dix` module:

```bash
cd examples/dix
go run ./basic
go run ./runtime_scope
go run ./inspect
```

## Core Examples

| Example | Focus | Directory |
| --- | --- | --- |
| `basic` | immutable app spec, build/start/stop, health checks, `logx` integration | [examples/dix/basic](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/basic) |
| `aggregate_params` | provider graph composition with multiple typed dependencies | [examples/dix/aggregate_params](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/aggregate_params) |
| `build_runtime` | explicit `Build()` to `Runtime` flow | [examples/dix/build_runtime](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/build_runtime) |
| `build_failure` | validation/build failure behavior | [examples/dix/build_failure](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/build_failure) |

## Advanced Examples

| Example | Focus | Directory |
| --- | --- | --- |
| `advanced_do_bridge` | explicit `do` bridge setup | [examples/dix/advanced_do_bridge](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/advanced_do_bridge) |
| `named_alias` | named services and typed alias binding | [examples/dix/named_alias](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/named_alias) |
| `runtime_scope` | request-like runtime scope and scoped providers | [examples/dix/runtime_scope](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/runtime_scope) |
| `transient` | transient provider semantics | [examples/dix/transient](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/transient) |
| `override` | structured overrides | [examples/dix/override](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/override) |
| `inspect` | runtime inspection and diagnostics | [examples/dix/inspect](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/inspect) |

## Example: Basic App Composition

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

## Example: Runtime Scope

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

## Example: Fine-Grained Inspection

```go
provided := advanced.ListProvidedServices(rt)
deps := advanced.ExplainNamedDependencies(rt, "tenant.default")

fmt.Println("provided services:", len(provided))
fmt.Println("tenant graph known:", deps["tenant.default"] != "")
```

Use the fine-grained inspection helpers when you only need one diagnostic view.
`InspectRuntime(...)` remains convenient, but it is the heavier aggregation path.
