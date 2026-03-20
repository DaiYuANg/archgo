package dbx

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func benchmarkOpenSQLiteDB(b *testing.B, name string) *sql.DB {
	b.Helper()
	path := filepath.Join(b.TempDir(), name+".db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		b.Fatalf("sql.Open returned error: %v", err)
	}
	db.SetMaxOpenConns(1)
	b.Cleanup(func() { _ = db.Close() })
	return db
}
