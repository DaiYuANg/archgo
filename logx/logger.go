package logx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/samber/lo"
	oopszerolog "github.com/samber/oops/loggers/zerolog"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger logs related events.
type Logger struct {
	logger  zerolog.Logger
	closers []io.Closer
	config  *config

	slogOnce   sync.Once
	slogLogger *slog.Logger
}

// Close closes related resources.
func (l *Logger) Close() error {
	errs := lo.FilterMap(l.closers, func(closer io.Closer, _ int) (error, bool) {
		if closer == nil {
			return nil, false
		}
		err := closer.Close()
		return err, err != nil
	})
	return errors.Join(errs...)
}

// Config returns related data.
func (l *Logger) Config() *config {
	return l.config
}

// New creates related functionality.
// Note.
// Note.
func New(opts ...Option) (*Logger, error) {
	cfg := defaultConfig()

	lo.ForEach(opts, func(opt Option, _ int) {
		if opt != nil {
			opt(&cfg)
		}
	})

	// Note.
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var writers []io.Writer
	var closers []io.Closer

	// console
	if cfg.console {
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: cfg.timeFormat,
			NoColor:    cfg.noColor,
		})
	}

	// file
	if cfg.filePath != "" {
		// Note.
		if err := os.MkdirAll(filepath.Dir(cfg.filePath), 0o755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		lj := &lumberjack.Logger{
			Filename:   cfg.filePath,
			MaxSize:    cfg.maxSize,
			MaxAge:     cfg.maxAge,
			MaxBackups: cfg.maxBackups,
			LocalTime:  cfg.localTime,
			Compress:   cfg.compress,
		}
		writers = append(writers, lj)
		closers = append(closers, lj)
	}

	// Note.
	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	level := cfg.level.ToZerologLevel()

	mw := io.MultiWriter(writers...)

	// Note.
	builder := zerolog.New(mw).
		Level(level).
		With().
		Timestamp()

	if cfg.addCaller {
		builder = builder.Caller()
	}

	z := builder.Logger()

	// Note.
	zerolog.ErrorStackMarshaler = oopszerolog.OopsStackMarshaller
	zerolog.ErrorMarshalFunc = oopszerolog.OopsMarshalFunc

	// Note.
	if cfg.setGlobal {
		zlog.Logger = z
	}

	return &Logger{
		logger:  z,
		closers: closers,
		config:  &cfg,
	}, nil
}

// MustNew creates related functionality.
func MustNew(opts ...Option) *Logger {
	logger, err := New(opts...)
	if err != nil {
		panic(err)
	}
	return logger
}

// NewDevelopment creates related functionality.
// Note.
func NewDevelopment() (*Logger, error) {
	return New(
		WithConsole(true),
		WithLevel(DebugLevel),
		WithCaller(true),
	)
}

// NewProduction creates related functionality.
// Note.
func NewProduction() (*Logger, error) {
	return New(
		WithConsole(false),
		WithLevel(InfoLevel),
	)
}

// SetGlobalLogger configures related behavior.
func (l *Logger) SetGlobalLogger() {
	zlog.Logger = l.logger
}

// WithContext documents related behavior.
func (l *Logger) WithContext(ctx context.Context) zerolog.Context {
	return l.logger.With().Ctx(ctx)
}

// WithTraceContext enriches logger with trace/span IDs from context when available.
func (l *Logger) WithTraceContext(ctx context.Context) *Logger {
	if l == nil || ctx == nil {
		return l
	}

	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() {
		return l
	}

	return l.WithFields(map[string]interface{}{
		"trace_id": spanContext.TraceID().String(),
		"span_id":  spanContext.SpanID().String(),
	})
}

// Logger returns related data.
func (l *Logger) Logger() zerolog.Logger {
	return l.logger
}

// Note.

// Debug logs related events.
func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.logger.Debug().Fields(fields).Msg(msg)
}

// Info logs related events.
func (l *Logger) Info(msg string, fields ...interface{}) {
	l.logger.Info().Fields(fields).Msg(msg)
}

// Warn logs related events.
func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.logger.Warn().Fields(fields).Msg(msg)
}

// Error logs related events.
func (l *Logger) Error(msg string, fields ...interface{}) {
	l.logger.Error().Fields(fields).Msg(msg)
}

// Fatal logs related events.
func (l *Logger) Fatal(msg string, fields ...interface{}) {
	l.logger.Fatal().Fields(fields).Msg(msg)
}

// Panic logs related events.
func (l *Logger) Panic(msg string, fields ...interface{}) {
	l.logger.Panic().Fields(fields).Msg(msg)
}

// WithField documents related behavior.
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return l.derive(l.logger.With().Interface(key, value).Logger())
}

// WithFields documents related behavior.
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	logger := l.logger
	for k, v := range fields {
		logger = logger.With().Interface(k, v).Logger()
	}
	return l.derive(logger)
}

// WithError documents related behavior.
func (l *Logger) WithError(err error) *Logger {
	return l.derive(l.logger.With().Err(err).Logger())
}

// WithCaller enables related functionality.
func (l *Logger) WithCaller(enabled bool) *Logger {
	if enabled {
		return l.derive(l.logger.With().Caller().Logger())
	}
	return l.derive(l.logger)
}

// Helper documents related behavior.
func Helper() {
	_, file, line, _ := runtime.Caller(1)
	fmt.Printf("Called from %s:%d\n", file, line)
}

// Sync synchronizes related state.
func (l *Logger) Sync() error {
	// zerolog synchronizes related state.
	return nil
}

// GetLevel retrieves related data.
func (l *Logger) GetLevel() Level {
	return l.config.level
}

// GetLevelString retrieves related data.
func (l *Logger) GetLevelString() string {
	return l.config.level.String()
}

// IsDebug checks related state.
func (l *Logger) IsDebug() bool {
	return l.config.level <= DebugLevel
}

// IsTrace checks related state.
func (l *Logger) IsTrace() bool {
	return l.config.level <= TraceLevel
}

// IsInfo checks related state.
func (l *Logger) IsInfo() bool {
	return l.config.level <= InfoLevel
}

// IsWarn checks related state.
func (l *Logger) IsWarn() bool {
	return l.config.level <= WarnLevel
}

// IsError checks related state.
func (l *Logger) IsError() bool {
	return l.config.level <= ErrorLevel
}

func (l *Logger) derive(z zerolog.Logger) *Logger {
	return &Logger{
		logger:  z,
		closers: l.closers,
		config:  l.config,
	}
}
