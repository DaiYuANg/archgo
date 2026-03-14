package bunx

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"strings"

	"github.com/samber/lo"
	"github.com/uptrace/bun"
)

// BaseRepository provides reusable CRUD helpers for Bun models.
type BaseRepository[T any] struct {
	db     bun.IDB
	logger *slog.Logger
}

// NewBaseRepository builds a generic repository over a Bun database handle.
func NewBaseRepository[T any](db bun.IDB, logger *slog.Logger) BaseRepository[T] {
	return BaseRepository[T]{
		db:     db,
		logger: logger,
	}
}

func (r BaseRepository[T]) List(ctx context.Context, orderExpr string) ([]T, error) {
	rows := make([]T, 0)
	q := r.db.NewSelect().Model(&rows)
	if strings.TrimSpace(orderExpr) != "" {
		q = q.OrderExpr(orderExpr)
	}
	if err := q.Scan(ctx); err != nil {
		r.logError("list", err)
		return nil, err
	}
	return rows, nil
}

func (r BaseRepository[T]) GetByID(ctx context.Context, id int64) (T, error) {
	var row T
	err := r.db.NewSelect().
		Model(&row).
		Where("id = ?", id).
		Limit(1).
		Scan(ctx)
	if err != nil {
		r.logError("get_by_id", err)
		var zero T
		return zero, err
	}
	return row, nil
}

func (r BaseRepository[T]) Create(ctx context.Context, row *T) error {
	if row == nil {
		err := errors.New("nil row")
		r.logError("create", err)
		return err
	}
	_, err := r.db.NewInsert().Model(row).Exec(ctx)
	if err != nil {
		r.logError("create", err)
	}
	return err
}

func (r BaseRepository[T]) UpdateByID(ctx context.Context, id int64, setters map[string]any) (bool, error) {
	q := r.db.NewUpdate().
		Model((*T)(nil)).
		Where("id = ?", id)

	lo.ForEach(lo.Entries(setters), func(entry lo.Entry[string, any], _ int) {
		q = q.Set(entry.Key+" = ?", entry.Value)
	})

	res, err := q.Exec(ctx)
	if err != nil {
		r.logError("update_by_id", err)
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		r.logError("update_by_id_rows_affected", err)
		return false, err
	}
	return affected > 0, nil
}

func (r BaseRepository[T]) DeleteByID(ctx context.Context, id int64) (bool, error) {
	res, err := r.db.NewDelete().
		Model((*T)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		r.logError("delete_by_id", err)
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		r.logError("delete_by_id_rows_affected", err)
		return false, err
	}
	return affected > 0, nil
}

func (r BaseRepository[T]) logError(op string, err error) {
	if r.logger == nil || err == nil {
		return
	}
	r.logger.Error(
		"bunx repository operation failed",
		slog.String("op", op),
		slog.String("model", modelName[T]()),
		slog.Any("error", err),
	)
}

func modelName[T any]() string {
	t := reflect.TypeOf((*T)(nil)).Elem()
	if t == nil {
		return "unknown"
	}
	return t.String()
}
