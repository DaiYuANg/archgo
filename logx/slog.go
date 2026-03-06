package logx

import (
	"context"
	"log/slog"

	slogzerolog "github.com/samber/slog-zerolog/v2"
)

// NewSlog creates related functionality.
// Note.
func NewSlog(l *Logger) *slog.Logger {
	if l == nil {
		return slog.Default()
	}

	l.slogOnce.Do(func() {
		l.slogLogger = buildSlog(l, slog.LevelDebug)
	})
	return l.slogLogger
}

// NewSlogWithLevel creates related functionality.
func NewSlogWithLevel(l *Logger, level slog.Level) *slog.Logger {
	return buildSlog(l, level)
}

// NewSlogWithContext creates related functionality.
func NewSlogWithContext(ctx context.Context, l *Logger) *slog.Logger {
	_ = ctx
	return NewSlog(l)
}

// SetDefaultSlog configures related behavior.
func SetDefaultSlog(l *Logger) *slog.Logger {
	logger := NewSlog(l)
	slog.SetDefault(logger)
	return logger
}

// SlogLogger retrieves related data.
func (l *Logger) SlogLogger() *slog.Logger {
	return NewSlog(l)
}

// SlogDebug logs related events.
func (l *Logger) SlogDebug(msg string, args ...any) {
	l.SlogLogger().Debug(msg, args...)
}

// SlogInfo logs related events.
func (l *Logger) SlogInfo(msg string, args ...any) {
	l.SlogLogger().Info(msg, args...)
}

// SlogWarn logs related events.
func (l *Logger) SlogWarn(msg string, args ...any) {
	l.SlogLogger().Warn(msg, args...)
}

// SlogError logs related events.
func (l *Logger) SlogError(msg string, args ...any) {
	l.SlogLogger().Error(msg, args...)
}

// SlogLogAttrs logs related events.
func (l *Logger) SlogLogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l.SlogLogger().LogAttrs(ctx, level, msg, attrs...)
}

func buildSlog(l *Logger, level slog.Level) *slog.Logger {
	if l == nil {
		return slog.Default()
	}

	handler := slogzerolog.Option{
		Logger:    &l.logger,
		AddSource: true,
		Level:     level,
	}.NewZerologHandler()

	return slog.New(handler)
}
