package options

import (
	"time"

	"testing"
)

func BenchmarkServerOptionsBuild(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		opts := DefaultServerOptions()
		Compose(
			WithBasePath("/api/v1"),
			WithValidation(true),
			WithOpenAPIInfo("ArcGo API", "1.0.0", "benchmark"),
			WithOpenAPIDocs(true),
		)(opts)

		compiled := opts.Build()
		if len(compiled) == 0 {
			b.Fatal("expected non-empty server options")
		}
	}
}

func BenchmarkHTTPClientOptionsBuild(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		opts := DefaultHTTPClientOptions()
		WithHTTPTimeout(5 * time.Second)(opts)
		client := opts.Build()
		if client.Timeout != 5*time.Second {
			b.Fatalf("unexpected timeout: %s", client.Timeout)
		}
	}
}

func BenchmarkContextOptionsBuild(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		opts := &ContextOptions{
			Timeout: 2 * time.Second,
			ValueKeys: map[contextValueKey]any{
				"trace_id": "bench-trace",
			},
		}
		ctx, cancel := opts.Build()
		if cancel != nil {
			cancel()
		}
		if ctx == nil {
			b.Fatal("context should not be nil")
		}
	}
}
