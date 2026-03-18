package bunx

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/uptrace/bun"
)

type queryLogHook struct {
	logger        *slog.Logger
	slowThreshold time.Duration
	logQuery      bool
	logArgs       bool
}

func newQueryLogHook(
	logger *slog.Logger,
	slowThreshold time.Duration,
	logQuery bool,
	logArgs bool,
) bun.QueryHook {
	return &queryLogHook{
		logger:        logger,
		slowThreshold: slowThreshold,
		logQuery:      logQuery,
		logArgs:       logArgs,
	}
}

func (h *queryLogHook) BeforeQuery(ctx context.Context, _ *bun.QueryEvent) context.Context {
	return ctx
}

func (h *queryLogHook) AfterQuery(_ context.Context, event *bun.QueryEvent) {
	if h == nil || h.logger == nil || event == nil {
		return
	}

	duration := time.Since(event.StartTime)
	attrs := []slog.Attr{
		slog.String("operation", event.Operation()),
		slog.Duration("duration", duration),
	}

	if event.Result != nil {
		if affected, err := event.Result.RowsAffected(); err == nil {
			attrs = append(attrs, slog.Int64("rows_affected", affected))
		}
	}
	if h.logQuery {
		attrs = append(attrs, slog.String("query", compactQuery(event)))
	}
	if h.logArgs && len(event.QueryArgs) > 0 {
		attrs = append(attrs, slog.Any("query_args", event.QueryArgs))
	}

	switch {
	case event.Err != nil:
		attrs = append(attrs, slog.Any("error", event.Err))
		h.logger.Error("bun query failed", attrsToAny(attrs)...)
	case h.slowThreshold > 0 && duration >= h.slowThreshold:
		h.logger.Warn("bun slow query", attrsToAny(attrs)...)
	default:
		h.logger.Debug("bun query", attrsToAny(attrs)...)
	}
}

func compactQuery(event *bun.QueryEvent) string {
	query := strings.TrimSpace(event.QueryTemplate)
	if query == "" {
		query = strings.TrimSpace(event.Query)
	}
	return strings.Join(strings.Fields(query), " ")
}

func attrsToAny(attrs []slog.Attr) []any {
	return lo.Map(attrs, func(attr slog.Attr, _ int) any {
		return attr
	})
}
