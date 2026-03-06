package logx

import (
	"io"
	"testing"

	"github.com/rs/zerolog"
)

func benchmarkLogger() *Logger {
	cfg := defaultConfig()
	cfg.console = false

	return &Logger{
		logger: zerolog.New(io.Discard).Level(zerolog.DebugLevel).With().Timestamp().Logger(),
		config: &cfg,
	}
}

func BenchmarkLoggerInfo(b *testing.B) {
	logger := benchmarkLogger()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "key", "value", "count", i)
	}
}

func BenchmarkLoggerWithFieldsInfo(b *testing.B) {
	logger := benchmarkLogger()
	fields := map[string]interface{}{
		"service": "arcgo",
		"env":     "bench",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.WithFields(fields).Info("with-fields")
	}
}

func BenchmarkSlogInfo(b *testing.B) {
	logger := benchmarkLogger()
	slogLogger := NewSlog(logger)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		slogLogger.Info("slog benchmark", "key", "value", "count", i)
	}
}
