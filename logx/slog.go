package logx

import (
	"context"
	"log/slog"

	slogzerolog "github.com/samber/slog-zerolog/v2"
)

// NewSlog 从 Logger 创建 slog 实例
// 自动设置 source code 位置
func NewSlog(l *Logger) *slog.Logger {
	handler := slogzerolog.Option{
		Logger:    &l.logger,
		AddSource: true,
		Level:     slog.LevelDebug, // 允许所有级别通过，由 zerolog 控制过滤
	}.NewZerologHandler()

	logger := slog.New(handler)

	// 可选：设置为全局默认 slog
	// slog.SetDefault(logger)

	return logger
}

// NewSlogWithLevel 创建带级别控制的 slog 实例
func NewSlogWithLevel(l *Logger, level slog.Level) *slog.Logger {
	handler := slogzerolog.Option{
		Logger:    &l.logger,
		AddSource: true,
		Level:     level,
	}.NewZerologHandler()

	return slog.New(handler)
}

// NewSlogWithContext 创建带 context 的 slog 实例
func NewSlogWithContext(ctx context.Context, l *Logger) *slog.Logger {
	handler := slogzerolog.Option{
		Logger:    &l.logger,
		AddSource: true,
	}.NewZerologHandler()

	logger := slog.New(handler)
	return logger
}

// SetDefaultSlog 设置全局默认 slog
func SetDefaultSlog(l *Logger) *slog.Logger {
	logger := NewSlog(l)
	slog.SetDefault(logger)
	return logger
}

// SlogLogger 便捷方法：直接从 Logger 获取 slog 接口
func (l *Logger) SlogLogger() *slog.Logger {
	return NewSlog(l)
}

// SlogDebug 记录 debug 级别日志（通过 slog）
func (l *Logger) SlogDebug(msg string, args ...any) {
	NewSlog(l).Debug(msg, args...)
}

// SlogInfo 记录 info 级别日志（通过 slog）
func (l *Logger) SlogInfo(msg string, args ...any) {
	NewSlog(l).Info(msg, args...)
}

// SlogWarn 记录 warn 级别日志（通过 slog）
func (l *Logger) SlogWarn(msg string, args ...any) {
	NewSlog(l).Warn(msg, args...)
}

// SlogError 记录 error 级别日志（通过 slog）
func (l *Logger) SlogError(msg string, args ...any) {
	NewSlog(l).Error(msg, args...)
}

// SlogLogAttrs 记录带属性的日志（通过 slog）
func (l *Logger) SlogLogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	NewSlog(l).LogAttrs(ctx, level, msg, attrs...)
}
