package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/httpx/adapter/std"
	"github.com/DaiYuANg/arcgo/httpx/options"
	"github.com/DaiYuANg/arcgo/logx"
	"github.com/DaiYuANg/arcgo/pkg/randomport"
	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5/middleware"
)

type UserOutput struct {
	Body struct {
		Users []string `json:"users"`
	}
}

func main() {
	logger, _ := logx.New(logx.WithConsole(true))
	defer func() { _ = logger.Close() }()
	slogLogger := logx.NewSlog(logger)

	fmt.Println("=== Example 1: Using ServerOptions + adapter construction options ===")
	serverOpts := options.DefaultServerOptions()
	serverOpts.Logger = slogLogger
	serverOpts.BasePath = "/api"
	serverOpts.PrintRoutes = true
	serverOpts.EnableValidation = true
	serverOpts.OpenAPIDocsEnabled = true
	serverOpts.HumaTitle = "ArcGo API"
	serverOpts.HumaVersion = "1.0.0"
	serverOpts.HumaDescription = "API Documentation"
	serverOpts.DocsPath = "/docs"
	serverOpts.OpenAPIPath = "/openapi.json"
	serverOpts.EnablePanicRecover = true
	serverOpts.EnableAccessLog = true

	// httpx server logs and adapter bridge logs are configured separately.
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
	stdAdapter.Router().Use(middleware.Logger, middleware.Recoverer, middleware.RequestID)

	server := httpx.NewServer(append(serverOpts.Build(), httpx.WithAdapter(stdAdapter))...)
	httpx.MustGet(server, "/users", func(ctx context.Context, input *struct{}) (*UserOutput, error) {
		out := &UserOutput{}
		out.Body.Users = []string{"Alice", "Bob", "Charlie"}
		return out, nil
	}, huma.OperationTags("users"))

	fmt.Println("=== Example 2: Using HTTP Client Options ===")
	clientOpts := &options.HTTPClientOptions{Timeout: 30 * time.Second}
	client := clientOpts.Build()
	fmt.Printf("HTTP Client Timeout: %v\n", client.Timeout)

	fmt.Println("=== Example 3: Using Context Options ===")
	ctxOpts := &options.ContextOptions{Timeout: 5 * time.Second}
	ctxOpts = options.WithContextValueOpt(ctxOpts, "request_id", "12345")
	ctx, cancel := ctxOpts.Build()
	defer cancel()
	fmt.Printf("Context value request_id: %v\n", ctx.Value("request_id"))

	port := randomport.MustFind()
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Config example server running on %s\n", addr)
	fmt.Printf("GET  /api/users\n")
	fmt.Printf("OpenAPI: http://localhost%s/openapi.json\n", addr)
	fmt.Printf("Docs:    http://localhost%s/docs\n", addr)
	fmt.Println(server.GetRoutes())

	if err := server.ListenAndServe(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
