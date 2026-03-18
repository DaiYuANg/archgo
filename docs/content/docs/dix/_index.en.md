---
title: 'dix'
linkTitle: 'dix'
description: 'Strongly typed modular app framework built on do'
weight: 6
---

## dix

`dix` is a strongly typed, module-oriented application framework built on top of `do`.
It provides an immutable app spec, typed providers and invokes, lifecycle hooks, validation,
and a runtime model without forcing most users to deal with `do` directly.

## API Status

- The public API is now in convergence.
- Future iterations should prioritize additive capabilities, integrations, and documentation.
- The default path is intended to remain stable and ergonomic for normal business applications.

## Core Model

- `App`: immutable application spec
- `Module`: immutable composition unit
- `ProviderN`: typed service registration
- `InvokeN`: typed eager initialization
- `HookFunc`: typed start and stop hooks
- `Build()`: compile spec into runtime
- `Runtime`: lifecycle, container access, health, and diagnostics

## Default Path

Use the `dix` package for most applications:

- `dix.New(...)`
- `dix.NewModule(...)`
- `dix.WithModuleProviders(...)`
- `dix.WithModuleSetups(...)` / `dix.WithModuleSetup(...)`
- `dix.WithModuleHooks(...)`
- `app.Validate()`
- `app.Build()`
- `runtime.Start(ctx)` / `runtime.Stop(ctx)` / `runtime.StopWithReport(ctx)`

The default path keeps the application model explicit and generic-first, while avoiding raw container mutation.

## Quick Start

```go
type Config struct {
    Port int
}

type Server struct {
    Logger *slog.Logger
    Config Config
}

configModule := dix.NewModule("config",
    dix.WithModuleProviders(
        dix.Provider0(func() Config { return Config{Port: 8080} }),
    ),
)

serverModule := dix.NewModule("server",
    dix.WithModuleImports(configModule),
    dix.WithModuleProviders(
        dix.Provider2(func(logger *slog.Logger, cfg Config) *Server {
            return &Server{Logger: logger, Config: cfg}
        }),
    ),
    dix.WithModuleHooks(
        dix.OnStart(func(ctx context.Context, srv *Server) error {
            srv.Logger.Info("server starting", "port", srv.Config.Port)
            return nil
        }),
        dix.OnStop(func(ctx context.Context, srv *Server) error {
            srv.Logger.Info("server stopping", "port", srv.Config.Port)
            return nil
        }),
    ),
)

logger, _ := logx.NewDevelopment()

app := dix.New(
    "demo",
    dix.WithProfile(dix.ProfileDev),
    dix.WithLogger(logger),
    dix.WithModules(configModule, serverModule),
)

if err := app.Validate(); err != nil {
    panic(err)
}

rt, err := app.Build()
if err != nil {
    panic(err)
}

if err := rt.Start(context.Background()); err != nil {
    panic(err)
}
defer func() {
    _, _ = rt.StopWithReport(context.Background())
}()
```

## Validation Model

`app.Validate()` performs typed graph validation for:

- typed providers
- typed invokes
- lifecycle hooks
- structured setup steps
- advanced structured bindings such as aliases, overrides, and named providers

Validation becomes conservative when you use explicit escape hatches:

- `dix.RawProvider(...)`
- `dix.RawInvoke(...)`
- `advanced.DoSetup(...)`

Those APIs remain available, but they intentionally trade validation strength for flexibility.

## Advanced Path

Use `github.com/DaiYuANg/arcgo/dix/advanced` when you need explicit container features:

- named services
- alias binding
- transient providers
- overrides and transient overrides
- runtime scopes
- inspection helpers
- raw `do` bridge setup

Common advanced APIs:

- `advanced.NamedProvider1(...)`
- `advanced.BindAlias[...]()`
- `advanced.TransientProvider0(...)`
- `advanced.Override0(...)`
- `advanced.OverrideTransient0(...)`
- `advanced.Scope(...)`
- `advanced.InspectRuntime(...)`
- `advanced.ExplainNamedDependencies(...)`

## Runtime Scope Example

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
_ = svc
```

## Stop Report

Use `runtime.StopWithReport(ctx)` when you need shutdown diagnostics.
It aggregates:

- lifecycle stop hook errors
- container shutdown errors from `do`

This is the preferred shutdown API when the caller needs visibility into teardown failures.

## Examples

- Example guide: [dix examples](./examples)
- Runnable examples in repository:
  - [examples/dix/basic](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/basic)
  - [examples/dix/build_runtime](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/build_runtime)
  - [examples/dix/build_failure](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/build_failure)
  - [examples/dix/runtime_scope](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/runtime_scope)
  - [examples/dix/transient](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/transient)
  - [examples/dix/override](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/override)
  - [examples/dix/inspect](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/inspect)

## Testing and Benchmarks

```bash
go test ./dix/...
go test ./dix -run ^$ -bench . -benchmem
```

Practical benchmark reading:

- typed resolve paths are cheap and suitable for hot paths
- `ResolveAssignableAs` is slower than typed alias binding
- inspection APIs are diagnostic paths and should not be treated as request hot paths
