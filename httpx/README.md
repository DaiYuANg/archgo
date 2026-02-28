# httpx

基于适配器模式的 Go HTTP 框架，支持通过 struct 方法命名/标签自动生成路由，可适配多种 Go HTTP 框架，并集成 Huma 自动生成 OpenAPI 3.0 文档。

## 特性

- ✅ **自动路由生成**：通过方法名、标签自动生成路由
- ✅ **适配器模式**：支持 `net/http`、`gin`、`echo`、`fiber` 等框架
- ✅ **Endpoint Object**：使用 struct 组织端点，代码更清晰
- ✅ **中间件支持**：统一的中间件接口
- ✅ **路由管理**：支持路由存储、查询、打印
- ✅ **slog 集成**：默认使用 `log/slog` 日志
- ✅ **函数式选项**：灵活的 `WithOption` 配置模式
- ✅ **OpenAPI 文档**：集成 Huma，自动生成 Swagger UI（`/docs`）
- ✅ **Prometheus 监控**：自动收集请求数、延迟、状态码等指标
- ✅ **OpenTelemetry 追踪**：支持分布式追踪，兼容 Jaeger、Zipkin 等

## 安装

```bash
go get github.com/DaiYuANg/toolkit4go/httpx
```

## 快速开始

### 基本用法

```go
package main

import (
    "context"
    "net/http"
    "github.com/DaiYuANg/toolkit4go/httpx"
    "github.com/DaiYuANg/toolkit4go/logx"
)

// UserEndpoint 用户端点
type UserEndpoint struct {
    httpx.BaseEndpoint
}

// GetUserList 获取用户列表
// 方法名自动转换为路由：GET /user/list
func (e *UserEndpoint) GetUserList(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
    e.Success(w, map[string]interface{}{
        "users": []string{"Alice", "Bob", "Charlie"},
    })
    return nil
}

// CreateUser 创建用户
// 方法名自动转换为路由：POST /user
func (e *UserEndpoint) CreateUser(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
    e.Success(w, map[string]interface{}{
        "message": "user created",
    })
    return nil
}

func main() {
    // 创建 logger
    logger, _ := logx.New(
        logx.WithConsole(true),
        logx.WithDebugLevel(),
    )
    defer logger.Close()
    slogLogger := logx.NewSlog(logger)

    // 创建服务器（启用 OpenAPI 文档）
    server := httpx.NewServer(
        httpx.WithLogger(slogLogger),
        httpx.WithMiddleware(httpx.MiddlewareLogger),
        httpx.WithMiddleware(httpx.MiddlewareRecovery),
        httpx.WithPrintRoutes(true),
        httpx.WithHuma(httpx.HumaOptions{
            Enabled: true,
            Title:   "My API",
            Version: "1.0.0",
        }),
    )

    // 注册端点（自动同步到 OpenAPI）
    _ = server.Register(&UserEndpoint{})

    // 启动服务器
    server.ListenAndServe(":8080")
}
```

### 访问 OpenAPI 文档

启动服务器后访问：
- **Swagger UI**: http://localhost:8080/docs
- **OpenAPI JSON**: http://localhost:8080/openapi.json

### 路由生成规则

#### 1. 方法命名方式（默认启用）

方法名前缀自动映射为 HTTP 方法和路径：

| 方法前缀 | HTTP 方法 | 路径转换 | 示例 |
|---------|----------|---------|------|
| `Get` | GET | 驼峰转斜杠 | `GetUserList` → `GET /user/list` |
| `List` | GET | 驼峰转斜杠 | `ListUsers` → `GET /users` |
| `Create` | POST | 驼峰转斜杠 | `CreateUser` → `POST /user` |
| `Update` | PUT | 驼峰转斜杠 | `UpdateUser` → `PUT /user` |
| `Patch` | PATCH | 驼峰转斜杠 | `PatchUser` → `PATCH /user/patch` |
| `Delete` | DELETE | 驼峰转斜杠 | `DeleteUser` → `DELETE /user` |

#### 2. 标签方式

使用 `http` 或 `route` 标签定义路由：

```go
type UserEndpoint struct {
    httpx.BaseEndpoint

    // 指定方法和路径
    GetUsers    func() `http:"GET /api/users"`
    
    // 只指定路径，默认 GET
    GetUser     func() `http:"/api/user/:id"`
    
    // 使用 route 标签
    CreateUser  func() `route:"POST /api/users"`
    UpdateUser  func() `route:"PUT /api/user/:id"`
}
```

### 使用 Gin 适配器

```go
package main

import (
    "github.com/DaiYuANg/toolkit4go/httpx"
    "github.com/gin-gonic/gin"
)

func main() {
    // 创建 Gin 适配器
    adapter := httpx.NewGinAdapter()

    // 创建服务器
    server := httpx.NewServer(
        httpx.WithAdapter(adapter),
    )

    // 注册端点
    _ = server.Register(&UserEndpoint{})

    // 启动服务器
    adapter.Engine().Run(":8080")
}
```

### 使用 Echo 适配器

```go
package main

import (
    "github.com/DaiYuANg/toolkit4go/httpx"
    "github.com/labstack/echo/v4"
)

func main() {
    // 创建 Echo 适配器
    adapter := httpx.NewEchoAdapter()

    // 创建服务器
    server := httpx.NewServer(
        httpx.WithAdapter(adapter),
    )

    // 注册端点
    _ = server.Register(&UserEndpoint{})

    // 启动服务器
    adapter.Engine().Start(":8080")
}
```

### 使用 Fiber 适配器

```go
package main

import (
    "github.com/DaiYuANg/toolkit4go/httpx"
    "github.com/gofiber/fiber/v2"
)

func main() {
    // 创建 Fiber 适配器
    adapter := httpx.NewFiberAdapter()

    // 创建服务器
    server := httpx.NewServer(
        httpx.WithAdapter(adapter),
    )

    // 注册端点
    _ = server.Register(&UserEndpoint{})

    // 启动服务器
    adapter.App().Listen(":8080")
}
```

### Huma OpenAPI 集成

httpx 集成了 [Huma](https://huma.rocks/) 库，支持自动生成 OpenAPI 3.0 文档和 Swagger UI。

**工作原理**：
- 使用 `WithHuma` 选项启用 OpenAPI 文档
- 每个适配器独立集成 Huma 官方适配器（humagin、humaecho、humafiber、humachi）
- 注册路由时自动同步到 Huma OpenAPI
- 启动时自动提供 `/docs` (Swagger UI) 和 `/openapi.json`

```go
package main

import (
    "github.com/DaiYuANg/toolkit4go/httpx"
)

func main() {
    // 创建服务器，启用 OpenAPI 文档
    server := httpx.NewServer(
        httpx.WithHuma(httpx.HumaOptions{
            Enabled:     true,
            Title:       "My API",
            Version:     "1.0.0",
            Description: "API Documentation",
        }),
    )

    // 注册端点（自动同步到 OpenAPI）
    _ = server.Register(&UserEndpoint{})

    // 启动服务器
    server.ListenAndServe(":8080")
}

// 访问：
// - Swagger UI:   http://localhost:8080/docs
// - OpenAPI JSON: http://localhost:8080/openapi.json
```

**注意**：
- 所有适配器（std、Gin、Echo、Fiber）都支持 Huma OpenAPI
- 每个适配器使用对应的 Huma 官方适配器（如 Gin 使用 humagin）
- 必须使用 `server.ListenAndServe()` 启动服务器以支持 OpenAPI 文档

### 中间件

#### 内置中间件

```go
server := httpx.NewServer(
    httpx.WithMiddleware(httpx.MiddlewareLogger),     // 日志中间件
    httpx.WithMiddleware(httpx.MiddlewareRecovery),   // 恢复中间件
    httpx.WithMiddleware(httpx.MiddlewareCORS("*")),  // CORS 中间件
)
```

#### 自定义中间件

```go
// 认证中间件
func AuthMiddleware(next httpx.HandlerFunc) httpx.HandlerFunc {
    return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return nil
        }
        // 验证 token...
        return next(ctx, w, r)
    }
}

server := httpx.NewServer(
    httpx.WithMiddleware(AuthMiddleware),
)
```

### 路由管理

#### 打印路由

```go
server := httpx.NewServer(
    httpx.WithPrintRoutes(true),
)
```

输出：
```
INFO Registered routes count=5
INFO   GET /user/list -> GetUserList
INFO   POST /user -> CreateUser
```

#### 获取路由

```go
// 获取所有路由
routes := server.GetRoutes()

// 按方法过滤
getRoutes := server.GetRoutesByMethod(http.MethodGet)

// 按路径前缀过滤
apiRoutes := server.GetRoutesByPath("/api")

// 检查路由是否存在
exists := server.HasRoute(http.MethodGet, "/user/list")

// 路由数量
count := server.RouteCount()
```

### 路由前缀和分组

#### 基础路径

```go
server := httpx.NewServer(
    httpx.WithBasePath("/api/v1"),
)
```

#### 注册时添加前缀

```go
// 所有路由会添加 /api/v1 前缀
_ = server.RegisterWithPrefix("/api/v1", &UserEndpoint{})
```

### BaseEndpoint 辅助方法

`httpx.BaseEndpoint` 提供常用的辅助方法：

```go
type MyEndpoint struct {
    httpx.BaseEndpoint
}

func (e *MyEndpoint) Handler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
    // JSON 响应
    e.JSON(w, map[string]interface{}{"data": "value"}, http.StatusOK)

    // 成功响应
    e.Success(w, map[string]interface{}{"users": []string{"Alice", "Bob"}})

    // 错误响应
    e.Error(w, "error message", http.StatusBadRequest)

    // 获取请求头
    token := e.GetHeader(r, "Authorization")

    // 获取查询参数
    id := e.GetQuery(r, "id", "default")
    id = e.GetQueryOrDefault(r, "id", "123")

    return nil
}
```

## API 参考

### Server 选项

| 函数 | 说明 |
|------|------|
| `WithAdapter(adapter Adapter)` | 设置适配器 |
| `WithAdapterName(name string)` | 通过名称设置适配器 |
| `WithGenerator(gen *RouterGenerator)` | 设置路由生成器 |
| `WithBasePath(path string)` | 设置基础路径 |
| `WithMiddleware(mws ...MiddlewareFunc)` | 注册中间件 |
| `WithLogger(logger *slog.Logger)` | 设置日志记录器 |
| `WithPrintRoutes(enabled bool)` | 启用路由打印 |

### Server 方法

| 方法 | 说明 |
|------|------|
| `Register(endpoints ...interface{})` | 注册端点 |
| `RegisterWithPrefix(prefix string, endpoints ...interface{})` | 注册端点并添加前缀 |
| `GetRoutes() []RouteInfo` | 获取所有路由 |
| `GetRoutesByMethod(method string) []RouteInfo` | 按方法过滤路由 |
| `GetRoutesByPath(prefix string) []RouteInfo` | 按路径过滤路由 |
| `HasRoute(method, path string) bool` | 检查路由是否存在 |
| `RouteCount() int` | 路由数量 |
| `Logger() *slog.Logger` | 获取日志记录器 |
| `Adapter() Adapter` | 获取适配器 |

### RouterGenerator 选项

```go
gen := httpx.NewRouterGenerator(httpx.GeneratorOptions{
    BasePath:       "/api",        // 基础路径
    UseComment:     false,         // 是否使用注释解析
    UseTag:         true,          // 是否使用标签解析
    UseNaming:      true,          // 是否使用方法名解析
    TagKey:         "route",       // 标签键名
    MethodPrefixes: map[string]string{
        "Get":    http.MethodGet,
        "Create": http.MethodPost,
        // ...
    },
})
```

### 适配器

| 适配器 | 名称 | Huma 官方适配器 | 说明 |
|-------|------|---------------|------|
| `StdHTTPAdapter` | `std` | `humachi` | 标准 net/http 库 |
| `GinAdapter` | `gin` | `humagin` | Gin 框架 |
| `EchoAdapter` | `echo` | `humaecho` | Echo 框架 |
| `FiberAdapter` | `fiber` | `humafiber` | Fiber v2 框架 |

**OpenAPI 文档**：所有适配器都支持通过 `WithHuma` 选项启用 OpenAPI 文档生成。

### 已注册的适配器

```go
// 获取所有已注册的适配器名称
adapters := httpx.RegisteredAdapters()
```

## 示例

完整示例请参考 `examples/` 目录：

```bash
# 标准库适配器 (net/http) + OpenAPI
go run github.com/DaiYuANg/toolkit4go/httpx/examples/std

# Gin 适配器 + OpenAPI
go run github.com/DaiYuANg/toolkit4go/httpx/examples/gin

# Echo 适配器 + OpenAPI
go run github.com/DaiYuANg/toolkit4go/httpx/examples/echo

# Fiber 适配器 + OpenAPI
go run github.com/DaiYuANg/toolkit4go/httpx/examples/fiber
```

所有示例都启用了 Huma OpenAPI 文档，启动后访问：
- **Swagger UI**: http://localhost:8080/docs
- **OpenAPI JSON**: http://localhost:8080/openapi.json

### 运行输出示例

```
=== Registered Routes ===
Total routes: 5

GET:
  /users                         -> ListUsers
  /user                          -> GetUser

POST:
  /new/user                      -> CreateNewUser

PUT:
  /user/info                     -> UpdateUserInfo

DELETE:
  /user/account                  -> DeleteUserAccount

=== OpenAPI Documentation ===
OpenAPI JSON: http://localhost:8080/openapi.json
Swagger UI:   http://localhost:8080/docs

Server starting on :8080
Adapter: Fiber
```

## License

MIT

## 监控与追踪

httpx 内置了 Prometheus 和 OpenTelemetry 支持，可以方便地集成到可观测性系统中。

### Prometheus 指标

内置中间件自动收集以下指标：

- `http_requests_total` - HTTP 请求总数（按 method、path、status 标签）
- `http_request_duration_seconds` - HTTP 请求延迟直方图
- `http_requests_in_flight` - 正在处理的请求数

### OpenTelemetry 追踪

内置中间件自动创建 span，记录：
- HTTP 方法、路径、URL
- 响应状态码
- 请求延迟

### 使用示例

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/DaiYuANg/toolkit4go/httpx"
    "github.com/DaiYuANg/toolkit4go/httpx/middleware"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
    ctx := context.Background()
    
    // 初始化 OpenTelemetry
    exporter, _ := otlptracehttp.New(ctx,
        otlptracehttp.WithEndpoint("localhost:4318"),
        otlptracehttp.WithInsecure(),
    )
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(resource.NewWithAttributes(
            semconv.ServiceName("my-service"),
        )),
    )
    otel.SetTracerProvider(tp)
    
    // 创建服务器
    server := httpx.NewServer()
    _ = server.Register(&UserEndpoint{})
    
    // 组合路由
    mux := http.NewServeMux()
    
    // 应用路由（带监控中间件）
    mux.Handle("/", 
        middleware.OpenTelemetryMiddleware(
            middleware.PrometheusMiddleware(server),
        ),
    )
    
    // Prometheus 指标
    mux.Handle("/metrics", promhttp.Handler())
    
    // 启动服务器
    http.ListenAndServe(":8080", mux)
}
```

### 访问端点

- **应用**: http://localhost:8080
- **Prometheus 指标**: http://localhost:8080/metrics
- **OpenAPI 文档**: http://localhost:8080/docs

### 运行 Jaeger 收集追踪

```bash
docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

访问 Jaeger UI: http://localhost:16686
