---
title: 'dix'
linkTitle: 'dix'
description: '基于 do 的强类型模块化应用框架'
weight: 6
---

## dix

`dix` 是构建在 `do` 之上的强类型、模块化应用框架。
它提供不可变 `App` 规格、typed provider/invoke、生命周期 hook、构建期校验，
以及独立的运行时模型，同时默认路径不会强迫业务代码直接接触 `do`。

## API 状态

- 当前对外 API 已进入收口阶段。
- 后续迭代应以增量能力、更多集成、以及文档和示例完善为主。
- 默认路径面向正常业务应用，后续应尽量保持稳定和易用。

## 核心建模

- `App`: 不可变应用规格
- `Module`: 不可变模块描述
- `ProviderN`: 强类型服务注册
- `InvokeN`: 强类型预热初始化
- `HookFunc`: 强类型启动和停止 hook
- `Build()`: 将规格编译为运行时
- `Runtime`: 生命周期、容器访问、健康检查、诊断入口

## 默认路径

大多数应用直接使用 `dix` 包即可：

- `dix.New(...)`
- `dix.NewModule(...)`
- `dix.WithModuleProviders(...)`
- `dix.WithModuleSetups(...)` / `dix.WithModuleSetup(...)`
- `dix.WithModuleHooks(...)`
- `app.Validate()`
- `app.Build()`
- `runtime.Start(ctx)` / `runtime.Stop(ctx)` / `runtime.StopWithReport(ctx)`

这条默认路径保持了显式建模、泛型优先和较强校验，同时避免了原始容器操作。

## 快速开始

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

## 校验模型

`app.Validate()` 会对 typed graph 做静态校验，覆盖：

- typed provider
- typed invoke
- lifecycle hook
- 结构化 setup
- advanced 路径中的 alias、override、named provider 等结构化绑定

当你使用显式逃生口时，校验会变得更保守：

- `dix.RawProvider(...)`
- `dix.RawInvoke(...)`
- `advanced.DoSetup(...)`

这些 API 仍然保留，但它们会用更高灵活性换取更弱的静态保证。

## 高级路径

当你需要显式容器能力时，使用 `github.com/DaiYuANg/arcgo/dix/advanced`：

- named service
- alias 绑定
- transient provider
- override / transient override
- runtime scope
- inspection helper
- 原生 `do` bridge setup

常见高级 API：

- `advanced.NamedProvider1(...)`
- `advanced.BindAlias[...]()`
- `advanced.TransientProvider0(...)`
- `advanced.Override0(...)`
- `advanced.OverrideTransient0(...)`
- `advanced.Scope(...)`
- `advanced.InspectRuntime(...)`
- `advanced.ExplainNamedDependencies(...)`

## Runtime Scope 示例

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

需要停机诊断时，优先使用 `runtime.StopWithReport(ctx)`。
它会聚合：

- stop hook 错误
- `do` 容器 shutdown 错误

如果调用方需要对 teardown 失败做更细粒度处理，这个 API 比单纯的 `Stop(ctx)` 更合适。

## 示例

- 示例总览页：[dix examples](./examples)
- 仓库中的可运行示例：
  - [examples/dix/basic](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/basic)
  - [examples/dix/build_runtime](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/build_runtime)
  - [examples/dix/build_failure](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/build_failure)
  - [examples/dix/runtime_scope](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/runtime_scope)
  - [examples/dix/transient](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/transient)
  - [examples/dix/override](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/override)
  - [examples/dix/inspect](https://github.com/DaiYuANg/arcgo/tree/main/examples/dix/inspect)

## 测试与 Benchmark

```bash
go test ./dix/...
go test ./dix -run ^$ -bench . -benchmem
```

对 benchmark 的实际理解：

- typed resolve 路径足够轻，可以进入热路径
- `ResolveAssignableAs` 比 typed alias 绑定更慢
- inspection API 是诊断路径，不应当按请求热路径去使用
