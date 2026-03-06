package std

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func benchmarkAdapterWithRoute(b *testing.B) *Adapter {
	b.Helper()

	a := New()
	a.Handle(http.MethodGet, "/ping", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		_ = ctx
		_ = r
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
		return nil
	})
	return a
}

func BenchmarkAdapterServeHTTP(b *testing.B) {
	adapter := benchmarkAdapterWithRoute(b)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()
		adapter.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("unexpected status code: %d", w.Code)
		}
	}
}

func BenchmarkJoinPath(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		value := joinPath("/api/v1/", "users")
		if value == "" {
			b.Fatal("unexpected empty path")
		}
	}
}
