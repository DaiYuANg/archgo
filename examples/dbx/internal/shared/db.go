package shared

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/DaiYuANg/arcgo/dbx"
	sqlitedialect "github.com/DaiYuANg/arcgo/dbx/dialect/sqlite"
	"github.com/DaiYuANg/arcgo/logx"
	_ "modernc.org/sqlite"
)

func NewLogger() *slog.Logger {
	development, err := logx.NewDevelopment()
	if err != nil {
		return nil
	}
	return development
}

func OpenSQLite(name string, opts ...dbx.Option) (*dbx.DB, func() error, error) {
	raw, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", name))
	if err != nil {
		return nil, nil, err
	}
	if err := raw.Ping(); err != nil {
		_ = raw.Close()
		return nil, nil, err
	}
	if _, err := raw.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		_ = raw.Close()
		return nil, nil, err
	}
	return dbx.NewWithOptions(raw, sqlitedialect.Dialect{}, opts...), raw.Close, nil
}
