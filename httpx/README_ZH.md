# HTTPX

`httpx` 是一个构建在 Huma 之上的轻量 HTTP 服务组织层。

## 你能得到什么

- 统一的强类型路由注册 API（`Get/Post/Put/Patch/Delete/...`）
- 基于 adapter 的运行时集成（`std`、`gin`、`echo`、`fiber`）
- 一等公民的 OpenAPI 与 docs 配置能力
- 保留 Huma 原生出口（`HumaAPI`、`OpenAPI`、`ConfigureOpenAPI`）
- Group 级 Huma middleware 与 operation 定制能力
- 可选的 `go-playground/validator` 请求校验
- 用于测试和诊断的路由自省 API

## 定位

`httpx` 不是重型 web framework，也不是要替代 Huma。
它的目标是在保留 Huma 原生能力的前提下，提供一套稳定的 `server/group/endpoint` 组织 API。

职责划分：

- `Huma`：typed operation、schema、OpenAPI、docs、middleware 模型
- `adapter/*`：运行时承载、router 集成、原生 middleware 生态
- `httpx`：统一服务组织 API，以及对 Huma 能力域的正式暴露

## 最小示例

```go
package main

import (
    "context"

    "github.com/DaiYuANg/arcgo/httpx"
    "github.com/DaiYuANg/arcgo/httpx/adapter/std"
    "github.com/go-chi/chi/v5/middleware"
)

type HealthOutput struct {
    Body struct {
        Status string `json:"status"`
    }
}

func main() {
    a := std.New()
    a.Router().Use(middleware.Logger, middleware.Recoverer)

    s := httpx.NewServer(
        httpx.WithAdapter(a),
        httpx.WithBasePath("/api"),
        httpx.WithOpenAPIInfo("My API", "1.0.0", "Service API"),
    )

    _ = httpx.Get(s, "/health", func(ctx context.Context, input *struct{}) (*HealthOutput, error) {
        out := &HealthOutput{}
        out.Body.Status = "ok"
        return out, nil
    })

    _ = s.ListenAndServe(":8080")
}
```

## 核心 API

### Server

- `NewServer(...)`
- `WithAdapter(...)`
- `WithBasePath(...)`
- `WithValidation()` / `WithValidator(...)`
- `WithPanicRecover(...)`
- `WithAccessLog(...)`
- `HumaAPI()`
- `OpenAPI()`
- `ConfigureOpenAPI(...)`
- `PatchOpenAPI(...)`
- `UseHumaMiddleware(...)`

### Docs / OpenAPI

构建期 docs 配置：

```go
s := httpx.NewServer(
    httpx.WithDocs(httpx.DocsOptions{
        Enabled:     true,
        DocsPath:    "/reference",
        OpenAPIPath: "/spec",
        SchemasPath: "/schemas",
        Renderer:    httpx.DocsRendererScalar,
    }),
)
```

运行期 docs 配置：

```go
s.ConfigureDocs(func(d *httpx.DocsOptions) {
    d.DocsPath = "/docs/internal"
    d.OpenAPIPath = "/openapi/internal"
})
```

OpenAPI patch：

```go
s.ConfigureOpenAPI(func(doc *huma.OpenAPI) {
    doc.Tags = append(doc.Tags, &huma.Tag{Name: "internal"})
})
```

说明：

- `WithOpenAPIInfo(...)` 和 `WithOpenAPIDocs(...)` 仍可继续使用
- `ConfigureDocs(...)` 现在会同步更新 adapter 管理的 docs 路由
- 内置 renderer：
  - `httpx.DocsRendererStoplightElements`
  - `httpx.DocsRendererScalar`
  - `httpx.DocsRendererSwaggerUI`

### Security / Components / 全局参数

```go
s := httpx.NewServer(
    httpx.WithSecurity(httpx.SecurityOptions{
        Schemes: map[string]*huma.SecurityScheme{
            "bearerAuth": {
                Type:   "http",
                Scheme: "bearer",
            },
        },
        Requirements: []map[string][]string{
            {"bearerAuth": {}},
        },
    }),
)

s.RegisterComponentParameter("Locale", &huma.Param{
    Name: "locale",
    In:   "query",
    Schema: &huma.Schema{Type: "string"},
})

s.RegisterGlobalHeader(&huma.Param{
    Name:   "X-Request-Id",
    In:     "header",
    Schema: &huma.Schema{Type: "string"},
})
```

可用 API：

- `RegisterSecurityScheme(...)`
- `SetDefaultSecurity(...)`
- `RegisterComponentParameter(...)`
- `RegisterComponentHeader(...)`
- `RegisterGlobalParameter(...)`
- `RegisterGlobalHeader(...)`
- `AddTag(...)`

### Group

基础分组：

```go
api := s.Group("/v1")
_ = httpx.GroupGet(api, "/users/{id}", getUser)
_ = httpx.GroupPost(api, "/users", createUser)
```

Group 级 Huma 能力：

```go
api := s.Group("/admin")
api.UseHumaMiddleware(authMiddleware)
api.DefaultTags("admin")
api.DefaultSecurity(map[string][]string{"bearerAuth": {}})
api.DefaultParameters(&huma.Param{
    Name:   "X-Tenant",
    In:     "header",
    Schema: &huma.Schema{Type: "string"},
})
api.DefaultSummaryPrefix("Admin")
api.DefaultDescription("Administrative APIs")
```

可用 Group API：

- `HumaGroup()`
- `UseHumaMiddleware(...)`
- `UseOperationModifier(...)`
- `UseSimpleOperationModifier(...)`
- `UseResponseTransformer(...)`
- `DefaultTags(...)`
- `DefaultSecurity(...)`
- `DefaultParameters(...)`
- `DefaultSummaryPrefix(...)`
- `DefaultDescription(...)`

## Typed Input 模式

```go
type GetUserInput struct {
    ID int `path:"id"`
}

type ListUsersInput struct {
    Page int `query:"page"`
    Size int `query:"size"`
}

type SecureInput struct {
    RequestID string `header:"X-Request-Id"`
}

type CreateUserInput struct {
    Body struct {
        Name  string `json:"name" validate:"required,min=2,max=64"`
        Email string `json:"email" validate:"required,email"`
    }
}
```

## Middleware 模型

`httpx` 采用双层 middleware 模型：

- adapter-native middleware：直接注册到 adapter 的 router/engine/app
- Huma middleware：通过 `Server.UseHumaMiddleware(...)` 或 `Group.UseHumaMiddleware(...)` 注册

adapter 原生 middleware 仍然应该保留在 adapter 上：

- `std`：`adapter.Router().Use(...)`
- `gin`：`adapter.Router().Use(...)`
- `echo`：`adapter.Router().Use(...)`
- `fiber`：`adapter.Router().Use(...)`

typed handler 的运行控制保留在 `httpx` 层：

- `WithPanicRecover(...)` 控制 `httpx` typed handler 的 panic recover
- `WithAccessLog(...)` 控制通过 server logger 输出请求日志

像 read/write/idle timeout、max header bytes 这类 listener/runtime 配置属于 adapter 或底层 server library，不再由 `httpx/options.ServerOptions` 统一承载。

## 日志语义

`httpx` 的 logger 目前是分层的：

- `httpx.WithLogger(...)` 配置的是 `httpx.Server` 自己的 logger
- adapter logger 配置的是 `adapter/std`、`adapter/gin`、`adapter/echo`、`adapter/fiber` bridge 层错误日志
- 框架原生日志和 logging middleware 仍属于各自框架

实际使用时建议这样理解：

- `httpx.WithLogger(...)` 用于 `httpx` 路由注册日志、access log 等 `httpx` 层输出
- 如果你希望 adapter bridge 层错误也走同一个 logger，需要显式配置 adapter logger
- `chi` / `gin` / `echo` / `fiber` 的日志中间件仍然应该注册在 adapter 的 router 或 engine 上

`httpx` 目前不承诺完全替代各框架的原生日志系统。

## Adapter 构造期配置

listener 和 bridge 层配置属于 adapter，而不是 `httpx.ServerOptions`。

对于 `std`、`gin`、`echo` 这类基于 `net/http` 的 adapter，应该通过 adapter 构造期 `Options` 配置：

```go
stdAdapter := std.NewWithOptions(std.Options{
    Logger: slogLogger,
    Server: std.ServerOptions{
        ReadTimeout:     15 * time.Second,
        WriteTimeout:    15 * time.Second,
        IdleTimeout:     60 * time.Second,
        ShutdownTimeout: 5 * time.Second,
        MaxHeaderBytes:  1 << 20,
    },
})
```

对于 `fiber`，timeout 属于 adapter 创建 `fiber.App` 时使用的 app config：

```go
fiberAdapter := fiber.NewWithOptions(nil, fiber.Options{
    Logger: slogLogger,
    App: fiber.AppOptions{
        ReadTimeout:     15 * time.Second,
        WriteTimeout:    15 * time.Second,
        IdleTimeout:     60 * time.Second,
        ShutdownTimeout: 5 * time.Second,
    },
})
```

如果你传入的是已经创建好的 framework 对象，那么最终仍以该 framework 对象自身配置为准。

## 路由自省 API

- `GetRoutes()`
- `GetRoutesByMethod(method)`
- `GetRoutesByPath(prefix)`
- `HasRoute(method, path)`
- `RouteCount()`

## options 包

可以通过 `httpx/options` 构建配置：

```go
opts := options.DefaultServerOptions()
opts.BasePath = "/api"
opts.HumaTitle = "Arc API"
opts.DocsPath = "/reference"
opts.DocsRenderer = httpx.DocsRendererSwaggerUI
opts.EnablePanicRecover = true
opts.EnableAccessLog = true

s := httpx.NewServer(append(opts.Build(), httpx.WithAdapter(a))...)
```

listener timeout 和 adapter logger 需要单独通过 adapter 构造期 options 配置。

## 测试方式

```go
req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
rec := httptest.NewRecorder()
s.ServeHTTP(rec, req)

if rec.Code != http.StatusOK {
    t.Fatal(rec.Code)
}
```

## FAQ

### 必须使用 Huma 风格输入结构体吗？

是的，`httpx` 的 typed route handler 以 Huma 输入输出建模为基础。

### 还能直接访问 Huma 原生 API 吗？

可以，使用 `HumaAPI()`、`OpenAPI()`、`HumaGroup()`。

### `httpx` 应该统一封装 adapter middleware 吗？

不应该。adapter-native middleware 保持在 adapter 层，`httpx` 负责 Huma 侧 middleware 和服务组织能力。

## 示例

- Quickstart：`go run ./httpx/examples/quickstart`
  - 最小 typed routes + validation + base path
- Auth：`go run ./httpx/examples/auth`
  - 演示 security schemes、全局 header 与认证 header 绑定
  - 英文说明见 [`httpx/examples/auth/README.md`](./examples/auth/README.md)
  - 中文说明见 [`httpx/examples/auth/README_ZH.md`](./examples/auth/README_ZH.md)
- Organization：`go run ./httpx/examples/organization`
  - 演示 docs 路径、security、全局 header、group defaults
  - 英文说明见 [`httpx/examples/organization/README.md`](./examples/organization/README.md)
  - 中文说明见 [`httpx/examples/organization/README_ZH.md`](./examples/organization/README_ZH.md)
