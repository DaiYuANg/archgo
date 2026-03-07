package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/httpx/adapter"
	"github.com/DaiYuANg/arcgo/httpx/adapter/std"
	"github.com/DaiYuANg/arcgo/httpx/middleware"
	"github.com/DaiYuANg/arcgo/logx"
	"github.com/DaiYuANg/arcgo/pkg/randomport"
	"github.com/danielgtaylor/huma/v2"
)

type HealthOutput struct {
	Body struct {
		Status string `json:"status"`
	}
}

func main() {
	logger, err := logx.New(logx.WithConsole(true), logx.WithDebugLevel())
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Close() }()

	stdAdapter := std.New(adapter.HumaOptions{
		Title:       "ArcGo Monitoring API",
		Version:     "1.0.0",
		Description: "Monitoring API",
		DocsPath:    "/docs",
		OpenAPIPath: "/openapi.json",
	}).WithLogger(logx.NewSlog(logger))

	server := httpx.NewServer(
		httpx.WithAdapter(stdAdapter),
		httpx.WithLogger(logx.NewSlog(logger)),
		httpx.WithPrintRoutes(true),
	)

	httpx.MustGet(server, "/health", func(ctx context.Context, input *struct{}) (*HealthOutput, error) {
		out := &HealthOutput{}
		out.Body.Status = "ok"
		return out, nil
	}, huma.OperationTags("monitoring"))

	server.Adapter().Handle(httpx.MethodGet, "/metrics", func(
		ctx context.Context,
		w http.ResponseWriter,
		r *http.Request,
	) error {
		_ = ctx
		middleware.MetricsHandler().ServeHTTP(w, r)
		return nil
	})

	_ = middleware.PrometheusMiddleware(middleware.OpenTelemetryMiddleware(server.Handler()))

	port := randomport.MustFind()
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Monitoring server starting on %s\n", addr)
	fmt.Printf("Health:     http://localhost%s/health\n", addr)
	fmt.Printf("Metrics:    http://localhost%s/metrics\n", addr)
	fmt.Printf("OpenAPI:    http://localhost%s/openapi.json\n", addr)
	fmt.Printf("Docs:       http://localhost%s/docs\n", addr)

	if err := server.ListenAndServe(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
