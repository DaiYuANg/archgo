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

func benchmarkOpenSQLiteDBMemory(b *testing.B) *sql.DB {
	b.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		b.Fatalf("sql.Open returned error: %v", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		b.Fatalf("PRAGMA foreign_keys: %v", err)
	}
	b.Cleanup(func() { _ = db.Close() })
	return db
}
