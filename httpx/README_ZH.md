# HTTPX

`httpx` 是一个面向多框架的 HTTP 适配层，核心目标是：

- 用统一的泛型强类型 API 注册路由（`Get/Post/...`）
- 通过 adapter 子包接入 `gin/fiber/echo/std(chi)` 生态
- 默认集成 Huma OpenAPI 文档
- 可选开启全局请求参数校验（`go-playground/validator`）

## 核心设计

- 路由注册：`httpx.Get/Post/Put/Patch/Delete/...`
- 路由分组：`server.Group("/prefix")` + `httpx.GroupGet/...`
- 路径参数：统一使用 `{id}` 风格（Huma 标准）
- 框架中间件：统一使用各适配器的 `Router()`
- OpenAPI：`httpx.WithOpenAPIInfo(...)` + `httpx.WithOpenAPIDocs(...)`
- 校验：`httpx.WithValidation()` 或 `httpx.WithValidator(...)`

`httpx` 本身不强绑定某个框架中间件体系，避免重复造轮子。

## 安装

按需引入 adapter：

```bash
go get github.com/DaiYuANg/arcgo/httpx/adapter/std
go get github.com/DaiYuANg/arcgo/httpx/adapter/gin
go get github.com/DaiYuANg/arcgo/httpx/adapter/fiber
go get github.com/DaiYuANg/arcgo/httpx/adapter/echo
```

## 快速开始（std/chi）

```go
package main

import (
	"context"

	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/httpx/adapter/std"
	"github.com/go-playground/validator/v10"
	"github.com/go-chi/chi/v5/middleware"
)

type CreateInput struct {
	Name string `json:"name" validate:"required,min=2,max=64"`
}

type HealthOutput struct {
	Body struct {
		Status string `json:"status"`
	} `json:"body"`
}

func main() {
	a := std.New()
	a.Router().Use(middleware.Logger, middleware.Recoverer)

	s := httpx.NewServer(
		httpx.WithAdapter(a),
		httpx.WithBasePath("/api"),
		httpx.WithPrintRoutes(true),
		httpx.WithValidator(validator.New(validator.WithRequiredStructEnabled())),
	)

	_ = httpx.Get(s, "/health", func(ctx context.Context, input *struct{}) (*HealthOutput, error) {
		out := &HealthOutput{}
		out.Body.Status = "ok"
		return out, nil
	})

	_ = httpx.Post(s, "/users", func(ctx context.Context, input *CreateInput) (*CreateInput, error) {
		return input, nil
	})

	_ = s.ListenAndServe(":8080")
}
```

## 分组路由

```go
api := server.Group("/v1")
_ = httpx.GroupGet(api, "/users", func(ctx context.Context, input *ListUsersInput) (*ListUsersOutput, error) {
	// ...
	return out, nil
})
```

## OpenAPI（Huma）

```go
server := httpx.NewServer(
	httpx.WithAdapter(adapter),
	httpx.WithOpenAPIInfo("My API", "1.0.0", "My service API"),
	httpx.WithOpenAPIDocs(false), // 生产环境可关闭 /docs 与 /openapi.*
	httpx.WithValidation(),
)
```

## 请求校验（validator）

- 默认关闭，开启方式：
  - `httpx.WithValidation()`：启用内置 validator
  - `httpx.WithValidator(v)`：注入自定义 validator
- 校验入口：输入结构体上的 `validate:"..."` tag
- 生效范围：泛型路由（默认 Huma 处理链）

示例：

```go
type CreateUserInput struct {
	Body struct {
		Name  string `json:"name" validate:"required,min=2,max=64"`
		Email string `json:"email" validate:"required,email"`
	} `json:"body"`
}
```

## 适配器与中间件

- Gin: `adapter/gin`
- Fiber: `adapter/fiber`
- Echo: `adapter/echo`
- Std(chi): `adapter/std`

示例里请直接在底层适配器对象上挂原生中间件：

- Gin: `ginAdapter.Router().Use(...)`
- Fiber: `fiberAdapter.Router().Use(...)`
- Echo: `echoAdapter.Router().Use(...)`
- Std(chi): `stdAdapter.Router().Use(...)`

## options 包

可通过 `httpx/options` 构造统一配置并转换成 `[]httpx.ServerOption`：

```go
opts := options.DefaultServerOptions()
opts.BasePath = "/api"
opts.PrintRoutes = true

server := httpx.NewServer(append(opts.Build(), httpx.WithAdapter(adapter))...)
```

## 示例代码

完整可运行示例见：

- `httpx/examples/main.go`
- `httpx/examples/std/main.go`
- `httpx/examples/gin/main.go`
- `httpx/examples/fiber/main.go`
- `httpx/examples/echo/main.go`
- `httpx/examples/config/main.go`
- `httpx/examples/monitoring/main.go`

## 兼容说明

`BaseEndpoint` / `HandlerEndpoint` 历史兼容 API 已移除，统一使用泛型路由 API（`Get/Post/Group*`）与显式输入输出结构体。
