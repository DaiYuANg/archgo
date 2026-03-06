package logx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/samber/oops"
	"go.opentelemetry.io/otel/trace"
)

type failingCloser struct {
	err error
}

func (c *failingCloser) Close() error {
	return c.err
}

func TestLevel(t *testing.T) {
	// Note.
	tests := []struct {
		level    Level
		expected string
	}{
		{TraceLevel, "trace"},
		{DebugLevel, "debug"},
		{InfoLevel, "info"},
		{WarnLevel, "warn"},
		{ErrorLevel, "error"},
		{FatalLevel, "fatal"},
		{PanicLevel, "panic"},
		{DisabledLevel, "disabled"},
		{NoLevel, "none"},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			if tt.level.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.level.String())
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
		hasError bool
	}{
		{"trace", TraceLevel, false},
		{"debug", DebugLevel, false},
		{"info", InfoLevel, false},
		{"warn", WarnLevel, false},
		{"warning", WarnLevel, false},
		{"error", ErrorLevel, false},
		{"fatal", FatalLevel, false},
		{"panic", PanicLevel, false},
		{"disabled", DisabledLevel, false},
		{"none", NoLevel, false},
		{"TRACE", TraceLevel, false},
		{"DEBUG", DebugLevel, false},
		{"invalid", NoLevel, true},
		{"", NoLevel, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level, err := ParseLevel(tt.input)
			if (err != nil) != tt.hasError {
				t.Errorf("ParseLevel(%q) error = %v, hasError = %v", tt.input, err, tt.hasError)
			}
			if level != tt.expected {
				t.Errorf("ParseLevel(%q) = %v, expected %v", tt.input, level, tt.expected)
			}
		})
	}
}

func TestMustParseLevel(t *testing.T) {
	// Note.
	level := MustParseLevel("debug")
	if level != DebugLevel {
		t.Errorf("expected DebugLevel, got %v", level)
	}

	// Note.
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseLevel should panic with invalid level")
		}
	}()
	_ = MustParseLevel("invalid")
}

func TestLevelConversion(t *testing.T) {
	levels := []Level{
		TraceLevel, DebugLevel, InfoLevel, WarnLevel,
		ErrorLevel, FatalLevel, PanicLevel, DisabledLevel,
	}

	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			zlLevel := level.ToZerologLevel()
			// Note.
			str := level.String()
			parsed, _ := ParseLevel(str)
			if parsed != level {
				t.Errorf("round-trip failed for %v", level)
			}
			_ = zlLevel
		})
	}
}

func TestLevelEnabled(t *testing.T) {
	tests := []struct {
		level   Level
		current Level
		want    bool
	}{
		{ErrorLevel, InfoLevel, true},
		{InfoLevel, InfoLevel, true},
		{DebugLevel, InfoLevel, false},
		{TraceLevel, ErrorLevel, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.level, tt.current), func(t *testing.T) {
			if got := tt.level.Enabled(tt.current); got != tt.want {
				t.Errorf("Level.Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogger(t *testing.T) {
	logger, err := New(WithConsole(true))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger.Close() }()

	// Note.
	logger.Info("test info message")
	logger.Debug("test debug message")
	logger.Warn("test warn message")
	logger.Error("test error message")
}

func TestLoggerWithFields(t *testing.T) {
	logger, err := New(WithConsole(true), WithLevel(DebugLevel))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger.Close() }()

	// Note.
	logger.WithField("user_id", "123").Info("user action")
	logger.WithFields(map[string]any{
		"user_id":   "456",
		"action":    "login",
		"timestamp": time.Now(),
	}).Info("user login")
}

func TestLoggerWithError(t *testing.T) {
	logger, err := New(WithConsole(true))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger.Close() }()

	// Note.
	err = os.ErrNotExist
	logger.WithError(err).Error("file not found")
}

func TestLoggerWithContext(t *testing.T) {
	logger, err := New(WithConsole(true), WithLevel(DebugLevel))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger.Close() }()

	ctx := context.Background()
	l := logger.WithContext(ctx).Logger()
	l.Info().Msg("context log")
}

func TestSlogIntegration(t *testing.T) {
	logger, err := New(WithConsole(true), WithLevel(DebugLevel))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger.Close() }()

	// Note.
	slogLogger := NewSlog(logger)
	slogLogger.Info("slog info message")
	slogLogger.Debug("slog debug message")
	slogLogger.Error("slog error message", "error", "test error")
}

func TestDevelopmentConfig(t *testing.T) {
	logger, err := New(DevelopmentConfig()...)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger.Close() }()

	logger.Debug("development mode debug")
	logger.Info("development mode info")
}

func TestProductionConfig(t *testing.T) {
	// Note.
	tmpFile := t.TempDir() + "/test.log"
	logger, err := New(ProductionConfig(tmpFile)...)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger.Close() }()

	logger.Info("production mode info")
	logger.Error("production mode error")

	// Note.
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("log file should exist")
	}
}

func TestMustNew(t *testing.T) {
	// Note.
	logger := MustNew(WithConsole(true))
	defer func() { _ = logger.Close() }()
	logger.Info("must new test")

	// Note.
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNew should panic with invalid config")
		}
	}()
	_ = MustNew(WithFileRotation(0, 7, 10)) // invalid maxSize
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name  string
		level Level
	}{
		{"trace", TraceLevel},
		{"debug", DebugLevel},
		{"info", InfoLevel},
		{"warn", WarnLevel},
		{"error", ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(WithConsole(true), WithLevel(tt.level))
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = logger.Close() }()

			if logger.GetLevel() != tt.level {
				t.Errorf("expected level %s, got %s", tt.level.String(), logger.GetLevel().String())
			}
		})
	}
}

func TestLoggerConvenienceMethods(t *testing.T) {
	logger, err := New(WithConsole(true), WithLevel(DebugLevel))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger.Close() }()

	// Note.
	logger.Debug("debug message", "key", "value")
	logger.Info("info message", "key", "value")
	logger.Warn("warn message", "key", "value")
	logger.Error("error message", "key", "value")
}

func TestLoggerConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "invalid maxSize",
			opts:    []Option{WithFileRotation(0, 7, 10)},
			wantErr: true,
		},
		{
			name:    "invalid maxAge",
			opts:    []Option{WithFileRotation(100, -1, 10)},
			wantErr: true,
		},
		{
			name:    "valid config with info level",
			opts:    []Option{WithLevel(InfoLevel), WithConsole(true)},
			wantErr: false,
		},
		{
			name:    "valid config with debug level",
			opts:    []Option{WithLevel(DebugLevel), WithConsole(true)},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOopsIntegration(t *testing.T) {
	logger, err := New(WithConsole(true), WithLevel(DebugLevel))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger.Close() }()

	// Note.
	err = oops.New("user.not_found")

	// Note.
	logger.LogOops(err)

	// Note.
	err = logger.Oops()
	if err == nil {
		t.Error("Expected error")
	}

	// Note.
	ctx := context.Background()
	err = logger.OopsWith(ctx)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestOopsf(t *testing.T) {
	logger, err := New(WithConsole(true), WithLevel(DebugLevel))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger.Close() }()

	oopsErr := logger.Oopsf("user.%s", "not_found")
	if oopsErr == nil {
		t.Fatal("expected non-nil error")
	}
	if !strings.Contains(oopsErr.Error(), "user.not_found") {
		t.Fatalf("unexpected error message: %s", oopsErr.Error())
	}
}

func TestWithCallerOption(t *testing.T) {
	logger1, err := New(WithConsole(true), WithCaller(false))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger1.Close() }()
	if logger1.Config().addCaller {
		t.Fatal("expected caller to be disabled")
	}

	logger2, err := New(WithConsole(true), WithCaller(true))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger2.Close() }()
	if !logger2.Config().addCaller {
		t.Fatal("expected caller to be enabled")
	}
}

func TestLoggerClose_ReturnsJoinedError(t *testing.T) {
	errA := errors.New("close-a")
	errB := errors.New("close-b")

	logger := &Logger{
		closers: []io.Closer{
			&failingCloser{err: errA},
			nil,
			&failingCloser{err: errB},
		},
	}

	err := logger.Close()
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, errA) {
		t.Fatal("expected joined error to include errA")
	}
	if !errors.Is(err, errB) {
		t.Fatal("expected joined error to include errB")
	}
}

func TestWithTraceContext(t *testing.T) {
	logger, err := New(WithConsole(true))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger.Close() }()

	invalid := logger.WithTraceContext(context.Background())
	if invalid != logger {
		t.Fatal("expected same logger for context without span")
	}

	traceID, traceIDErr := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	if traceIDErr != nil {
		t.Fatal(traceIDErr)
	}
	spanID, spanIDErr := trace.SpanIDFromHex("00f067aa0ba902b7")
	if spanIDErr != nil {
		t.Fatal(spanIDErr)
	}

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	withTrace := logger.WithTraceContext(ctx)
	if withTrace == logger {
		t.Fatal("expected derived logger when span context exists")
	}
}

func TestWithFieldTypedHelpers(t *testing.T) {
	logger, err := New(WithConsole(true), WithDebugLevel())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger.Close() }()

	typedField := WithFieldT(logger, "retry", 3)
	if typedField == nil {
		t.Fatal("expected non-nil logger")
	}

	typedFields := WithFieldsT(logger, map[string]int{
		"batch": 7,
	})
	if typedFields == nil {
		t.Fatal("expected non-nil logger")
	}
}
