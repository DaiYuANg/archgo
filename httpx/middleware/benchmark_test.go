package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkPrometheusMiddleware(b *testing.B) {
	handler := PrometheusMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r
		w.WriteHeader(http.StatusNoContent)
	}))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/metrics-demo", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusNoContent {
			b.Fatalf("unexpected status code: %d", w.Code)
		}
	}
}

func BenchmarkMetricsHandlerServeHTTP(b *testing.B) {
	handler := MetricsHandler()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("unexpected status code: %d", w.Code)
		}
	}
}
