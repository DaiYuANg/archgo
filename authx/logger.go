package authx

import (
	"log/slog"

	"github.com/DaiYuANg/arcgo/observability"
)

func normalizeLogger(logger *slog.Logger) *slog.Logger {
	return observability.NormalizeLogger(logger)
}
