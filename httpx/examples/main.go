package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DaiYuANg/toolkit4go/httpx"
	"github.com/DaiYuANg/toolkit4go/logx"
)

// UserEndpoint 用户相关端点
type UserEndpoint struct {
	httpx.BaseEndpoint
}

// ListUsers 获取用户列表
func (e *UserEndpoint) ListUsers(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	e.Success(w, map[string]interface{}{
		"users": []string{"Alice", "Bob", "Charlie"},
	})
	return nil
}

// GetUser 获取单个用户
func (e *UserEndpoint) GetUser(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := e.GetQuery(r, "id", "1")
	e.Success(w, map[string]interface{}{
		"user": map[string]string{
			"id":   id,
			"name": "User" + id,
		},
	})
	return nil
}

// CreateNewUser 创建用户
func (e *UserEndpoint) CreateNewUser(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	e.Success(w, map[string]interface{}{
		"message": "user created successfully",
	})
	return nil
}

// UpdateUserInfo 更新用户信息
func (e *UserEndpoint) UpdateUserInfo(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	e.Success(w, map[string]interface{}{
		"message": "user updated successfully",
	})
	return nil
}

// DeleteUserAccount 删除用户账户
func (e *UserEndpoint) DeleteUserAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	e.Success(w, map[string]interface{}{
		"message": "user deleted successfully",
	})
	return nil
}

func main() {
	// 创建 logx logger
	logger, err := logx.New(
		logx.WithConsole(true),
		logx.WithDebugLevel(),
	)
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	// 创建 slog logger
	slogLogger := logx.NewSlog(logger)

	// 创建端点实例
	userEndpoint := &UserEndpoint{}

	// 创建服务器（使用默认的 net/http 适配器）
	// 启用 Huma OpenAPI 文档
	server := httpx.NewServer(
		httpx.WithLogger(slogLogger),
		httpx.WithMiddleware(httpx.MiddlewareLogger),
		httpx.WithMiddleware(httpx.MiddlewareRecovery),
		httpx.WithMiddleware(httpx.MiddlewareCORS("*")),
		httpx.WithPrintRoutes(true),
		httpx.WithHuma(httpx.HumaOptions{
			Enabled:     true,
			Title:       "My API",
			Version:     "1.0.0",
			Description: "API built with httpx",
		}),
	)

	// 注册端点
	_ = server.Register(userEndpoint)

	// 注册带前缀的端点
	_ = server.RegisterWithPrefix("/api/v1", userEndpoint)

	// 打印路由信息
	fmt.Println("\n=== Registered Routes ===")
	fmt.Printf("Total routes: %d\n\n", server.RouteCount())

	// 按方法分组打印
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}
	for _, method := range methods {
		routes := server.GetRoutesByMethod(method)
		if len(routes) > 0 {
			fmt.Printf("%s:\n", method)
			for _, route := range routes {
				fmt.Printf("  %-30s -> %s\n", route.Path, route.HandlerName)
			}
			fmt.Println()
		}
	}

	// 打印 OpenAPI 信息
	if server.HasHuma() {
		fmt.Println("=== OpenAPI Documentation ===")
		fmt.Printf("OpenAPI JSON: http://localhost:8080/openapi.json\n")
		fmt.Printf("Swagger UI:   http://localhost:8080/docs\n")
		fmt.Println()
	}

	fmt.Println("Server starting on :8080")

	// 启动服务器
	err = server.ListenAndServe(":8080")
	if err != nil {
		panic(err)
	}
}
