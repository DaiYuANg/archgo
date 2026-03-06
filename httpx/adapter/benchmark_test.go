package adapter

import (
	"context"
	"testing"

	"github.com/danielgtaylor/huma/v2"
)

func BenchmarkApplyHumaConfig(b *testing.B) {
	opts := HumaOptions{
		Description: "benchmark docs",
		DocsPath:    "/docs/v1",
		OpenAPIPath: "/openapi/v1",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cfg := huma.DefaultConfig("benchmark", "1.0.0")
		ApplyHumaConfig(&cfg, opts)
		if cfg.DocsPath == "" || cfg.OpenAPIPath == "" {
			b.Fatal("expected docs/openapi paths to be set")
		}
	}
}

func BenchmarkRouteParamLookup(b *testing.B) {
	ctx := WithRouteParams(context.Background(), map[string]string{
		"id":    "123",
		"token": "abc",
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if RouteParam(ctx, "id") == "" {
			b.Fatal("expected route param id")
		}
		_ = RouteParam(ctx, "token")
		_ = RouteParam(ctx, "missing")
	}
}
