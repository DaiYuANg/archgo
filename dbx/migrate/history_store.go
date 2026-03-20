package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/DaiYuANg/arcgo/dbx/dialect"
	goosedatabase "github.com/pressly/goose/v3/database"
)

type historyStore struct {
	tableName     string
	dialect       dialect.Dialect
	metaByVersion collectionx.Map[int64, AppliedRecord]
}

func newHistoryStore(d dialect.Dialect, tableName string, metaByVersion collectionx.Map[int64, AppliedRecord]) *historyStore {
	return &historyStore{
		tableName:     tableName,
		dialect:       d,
		metaByVersion: metaByVersion,
	}
}

func (s *historyStore) Tablename() string {
	return s.tableName
}

func (s *historyStore) CreateVersionTable(ctx context.Context, db goosedatabase.DBTxConn) error {
	_, err := db.ExecContext(ctx, historyTableDDL(s.dialect, s.tableName))
	return err
}

func (s *historyStore) TableExists(ctx context.Context, db goosedatabase.DBTxConn) (bool, error) {
	query, err := historyTableExistsSQL(s.dialect)
	if err != nil {
		return false, err
	}
	var exists bool
	if err := db.QueryRowContext(ctx, query, s.tableName).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *historyStore) Insert(ctx context.Context, db goosedatabase.DBTxConn, req goosedatabase.InsertRequest) error {
	if req.Version == 0 {
		return nil
	}
	record, ok := s.metaByVersion.Get(req.Version)
	if !ok {
		return fmt.Errorf("dbx/migrate: migration metadata not found for version %d", req.Version)
	}
	record.AppliedAt = time.Now().UTC()
	record.Success = true
	return replaceAppliedRecordOnConn(ctx, db, s.dialect, s.tableName, record)
}

func (s *historyStore) Delete(ctx context.Context, db goosedatabase.DBTxConn, version int64) error {
	if version == 0 {
		return nil
	}
	record, ok := s.metaByVersion.Get(version)
	if !ok {
		return fmt.Errorf("dbx/migrate: migration metadata not found for version %d", version)
	}
	q := s.dialect.QuoteIdent
	deleteSQL := "DELETE FROM " + q(s.tableName) +
		" WHERE " + q("version") + " = " + s.dialect.BindVar(1) +
		" AND " + q("kind") + " = " + s.dialect.BindVar(2) +
		" AND " + q("description") + " = " + s.dialect.BindVar(3)
	_, err := db.ExecContext(ctx, deleteSQL, record.Version, string(record.Kind), record.Description)
	return err
}

func (s *historyStore) GetMigration(ctx context.Context, db goosedatabase.DBTxConn, version int64) (*goosedatabase.GetMigrationResult, error) {
	if version == 0 {
		return &goosedatabase.GetMigrationResult{IsApplied: true}, nil
	}
	record, ok := s.metaByVersion.Get(version)
	if !ok {
		return nil, goosedatabase.ErrVersionNotFound
	}

	query := specificAppliedMigrationSQL(s.dialect, s.tableName)
	var (
		appliedAt string
		success   bool
	)
	if err := db.QueryRowContext(ctx, query, record.Version, string(record.Kind), record.Description).Scan(&appliedAt, &success); err != nil {
		if err == sql.ErrNoRows {
			return nil, goosedatabase.ErrVersionNotFound
		}
		return nil, err
	}
	timestamp, err := time.Parse(timeLayout, appliedAt)
	if err != nil {
		return nil, fmt.Errorf("dbx/migrate: parse applied_at: %w", err)
	}
	return &goosedatabase.GetMigrationResult{Timestamp: timestamp, IsApplied: success}, nil
}

func (s *historyStore) GetLatestVersion(ctx context.Context, db goosedatabase.DBTxConn) (int64, error) {
	items, err := s.ListMigrations(ctx, db)
	if err != nil {
		return 0, err
	}
	var maxVersion int64
	for _, item := range items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	return maxVersion, nil
}

func (s *historyStore) ListMigrations(ctx context.Context, db goosedatabase.DBTxConn) ([]*goosedatabase.ListMigrationsResult, error) {
	rows, err := db.QueryContext(ctx, historyRowsForStatusSQL(s.dialect, s.tableName), string(KindRepeatable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := collectionx.NewList[*goosedatabase.ListMigrationsResult]()
	for rows.Next() {
		var (
			version     string
			description string
			kind        string
			appliedAt   string
			success     bool
		)
		if err := rows.Scan(&version, &description, &kind, &appliedAt, &success); err != nil {
			return nil, err
		}
		parsed, err := parseNumericVersion(version)
		if err != nil {
			return nil, err
		}
		record, ok := s.metaByVersion.Get(parsed)
		if !ok {
			continue
		}
		if record.Kind != Kind(kind) || record.Description != description {
			continue
		}
		items.Add(&goosedatabase.ListMigrationsResult{Version: parsed, IsApplied: true})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if items.Len() == 0 {
		items.Add(&goosedatabase.ListMigrationsResult{Version: 0, IsApplied: true})
	}
	return items.Values(), nil
}

func historyTableExistsSQL(d dialect.Dialect) (string, error) {
	switch d.Name() {
	case "sqlite":
		return "SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ?)", nil
	case "postgres":
		return "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = " + d.BindVar(1) + ")", nil
	case "mysql":
		return "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?)", nil
	default:
		return "", errors.ErrUnsupported
	}
}
