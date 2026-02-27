package logx

import (
	"log/slog"

	slogzerolog "github.com/samber/slog-zerolog/v2"
)

func NewSlog(l *Logger) *slog.Logger {
	handler := slogzerolog.Option{
		Logger:    &l.Logger,
		AddSource: true,
	}.NewZerologHandler()

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
