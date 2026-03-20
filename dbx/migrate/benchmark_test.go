package migrate

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"testing/fstest"

	_ "modernc.org/sqlite"
)

func BenchmarkFileSourceList(b *testing.B) {
	source := FileSource{
		FS: fstest.MapFS{
			"sql/V1__create_roles.sql":        &fstest.MapFile{Data: []byte("CREATE TABLE roles (id INTEGER PRIMARY KEY);\n")},
			"sql/V2__create_users.sql":        &fstest.MapFile{Data: []byte("CREATE TABLE users (id INTEGER PRIMARY KEY);\n")},
			"sql/U2__drop_users.sql":          &fstest.MapFile{Data: []byte("DROP TABLE users;\n")},
			"sql/R__refresh_materialized.sql": &fstest.MapFile{Data: []byte("SELECT 1;\n")},
		},
		Dir: "sql",
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		items, err := source.List()
		if err != nil {
			b.Fatalf("List returned error: %v", err)
		}
		if len(items) != 4 {
			b.Fatalf("unexpected migration count: %d", len(items))
		}
	}
}

func BenchmarkRunnerPendingSQL(b *testing.B) {
	ctx := context.Background()
	db := benchmarkOpenRunnerSQLiteDB(b, "pending")
	runner := NewRunner(db, testDialect{}, RunnerOptions{ValidateHash: true})
	source := FileSource{
		FS: fstest.MapFS{
			"sql/V1__create_logs.sql":  &fstest.MapFile{Data: []byte("CREATE TABLE logs (id INTEGER PRIMARY KEY);\n")},
			"sql/R__refresh_cache.sql": &fstest.MapFile{Data: []byte("SELECT 1;\n")},
		},
		Dir: "sql",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		items, err := runner.PendingSQL(ctx, source)
		if err != nil {
			b.Fatalf("PendingSQL returned error: %v", err)
		}
		if len(items) != 2 {
			b.Fatalf("unexpected pending count: %d", len(items))
		}
	}
}

func BenchmarkRunnerApplied(b *testing.B) {
	ctx := context.Background()
	db := benchmarkOpenRunnerSQLiteDB(b, "applied")
	runner := NewRunner(db, testDialect{}, RunnerOptions{ValidateHash: true})

	_, err := runner.UpGo(ctx, NewGoMigration("1", "create sample", func(ctx context.Context, tx *sql.Tx) error {
		_, execErr := tx.ExecContext(ctx, `CREATE TABLE sample (id INTEGER PRIMARY KEY)`)
		return execErr
	}, nil))
	if err != nil {
		b.Fatalf("UpGo returned error: %v", err)
	}

	source := FileSource{
		FS: fstest.MapFS{
			"sql/V2__seed_sample.sql": &fstest.MapFile{Data: []byte("INSERT INTO sample (id) VALUES (1);\n")},
		},
		Dir: "sql",
	}
	if _, err := runner.UpSQL(ctx, source); err != nil {
		b.Fatalf("UpSQL returned error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		items, err := runner.Applied(ctx)
		if err != nil {
			b.Fatalf("Applied returned error: %v", err)
		}
		if len(items) != 2 {
			b.Fatalf("unexpected applied count: %d", len(items))
		}
	}
}

func benchmarkOpenRunnerSQLiteDB(b *testing.B, name string) *sql.DB {
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
