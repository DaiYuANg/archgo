package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DaiYuANg/toolkit4go/httpx"
	"github.com/DaiYuANg/toolkit4go/httpx/middleware"
	"github.com/DaiYuANg/toolkit4go/logx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// UserEndpoint 用户端点
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
		"user": map[string]string{"id": id, "name": "User" + id},
	})
	return nil
}

// initOTLPTracer 初始化 OpenTelemetry Trace Exporter
func initOTLPTracer(ctx context.Context, serviceName string) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("localhost:4318"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

func main() {
	ctx := context.Background()

	// 初始化 OpenTelemetry
	tp, err := initOTLPTracer(ctx, "httpx-example")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = tp.Shutdown(ctx)
	}()

	// 创建 logger
	logger, err := logx.New(
		logx.WithConsole(true),
		logx.WithDebugLevel(),
	)
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	// 创建端点
	userEndpoint := &UserEndpoint{}

	// 创建服务器
	// 启用 Huma OpenAPI + Prometheus + OpenTelemetry
	server := httpx.NewServer(
		httpx.WithLogger(logx.NewSlog(logger)),
		httpx.WithPrintRoutes(true),
		httpx.WithHuma(httpx.HumaOptions{
			Enabled: true,
			Title:   "My API with Monitoring",
			Version: "1.0.0",
		}),
	)

	// 注册端点
	_ = server.Register(userEndpoint)

	// 创建 mux 组合所有路由
	mux := http.NewServeMux()

	// 注册应用路由（带监控中间件）
	mux.Handle("/", middleware.OpenTelemetryMiddleware(
		middleware.PrometheusMiddleware(server),
	))

	// 注册 Prometheus 指标路由
	mux.Handle("/metrics", promhttp.Handler())

	// 注册 OpenAPI 路由（Huma 已自动在 ListenAndServe 中处理）
	// 这里我们手动组合 mux 以支持 metrics

	fmt.Println("=== Server with Monitoring ===")
	fmt.Println("Application: http://localhost:8080")
	fmt.Println("Prometheus Metrics: http://localhost:8080/metrics")
	fmt.Println("OpenAPI Docs: http://localhost:8080/docs")
	fmt.Println("OpenAPI JSON: http://localhost:8080/openapi.json")
	fmt.Println()
	fmt.Println("Start Jaeger/OTLP collector to see traces:")
	fmt.Println("  docker run -d --name jaeger \\")
	fmt.Println("    -e COLLECTOR_OTLP_ENABLED=true \\")
	fmt.Println("    -p 16686:16686 \\")
	fmt.Println("    -p 4318:4318 \\")
	fmt.Println("    jaegertracing/all-in-one:latest")
	fmt.Println()

	// 启动服务器
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
