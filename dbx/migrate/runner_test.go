package migrate

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/DaiYuANg/arcgo/dbx/internal/testsql"
	_ "modernc.org/sqlite"
)

type testDialect struct{}

func (testDialect) Name() string                                         { return "sqlite" }
func (testDialect) BindVar(_ int) string                                 { return "?" }
func (testDialect) QuoteIdent(ident string) string                       { return `"` + ident + `"` }
func (testDialect) RenderLimitOffset(limit, offset *int) (string, error) { return "", nil }

func TestRunnerUpGoCreatesHistoryAndAppliesMigration(t *testing.T) {
	ctx := context.Background()
	db := openSQLiteRunnerDB(t)
	runner := NewRunner(db, testDialect{}, RunnerOptions{ValidateHash: true})

	report, err := runner.UpGo(ctx, NewGoMigration("1", "create sample", func(ctx context.Context, tx *sql.Tx) error {
		_, execErr := tx.ExecContext(ctx, `CREATE TABLE sample (id INTEGER PRIMARY KEY)`)
		return execErr
	}, nil))
	if err != nil {
		t.Fatalf("UpGo returned error: %v", err)
	}
	if len(report.Applied) != 1 || report.Applied[0].Version != "1" || report.Applied[0].Kind != KindGo {
		t.Fatalf("unexpected go migration report: %+v", report)
	}

	applied, err := runner.Applied(ctx)
	if err != nil {
		t.Fatalf("Applied returned error: %v", err)
	}
	if len(applied) != 1 || applied[0].Version != "1" || applied[0].Kind != KindGo || !applied[0].Success {
		t.Fatalf("unexpected applied records: %+v", applied)
	}

	if !sqliteTableExists(t, db, "sample") {
		t.Fatalf("expected sample table to exist")
	}
	if !sqliteTableExists(t, db, "schema_history") {
		t.Fatalf("expected schema_history table to exist")
	}
}

func TestRunnerPendingSQLTracksRepeatableChecksum(t *testing.T) {
	sqlDB, _, cleanup, err := testsql.Open(testsql.Plan{
		Execs: []testsql.ExecPlan{{SQL: `CREATE TABLE IF NOT EXISTS "schema_history" ("version" VARCHAR(255) NOT NULL, "description" VARCHAR(255) NOT NULL, "kind" VARCHAR(32) NOT NULL, "checksum" VARCHAR(128) NOT NULL, "success" BOOLEAN NOT NULL, "applied_at" VARCHAR(64) NOT NULL, PRIMARY KEY ("version", "kind", "description"))`}},
		Queries: []testsql.QueryPlan{{
			SQL:     `SELECT "version", "description", "kind", "applied_at", "checksum", "success" FROM "schema_history" ORDER BY "applied_at", "version", "description"`,
			Columns: []string{"version", "description", "kind", "applied_at", "checksum", "success"},
			Rows:    [][]driver.Value{{"", "refresh cache", "repeatable", "2026-03-19T22:00:00Z", checksumString(strings.Join([]string{"repeatable", "", "refresh cache", "SELECT 2;\n", ""}, "\n--dbx-migrate--\n")), true}},
		}},
	})
	if err != nil {
		t.Fatalf("testsql.Open returned error: %v", err)
	}
	defer cleanup()

	source := FileSource{
		FS: fstest.MapFS{
			"sql/R__refresh_cache.sql": &fstest.MapFile{Data: []byte("SELECT 1;\n")},
		},
		Dir: "sql",
	}
	runner := NewRunner(sqlDB, testDialect{}, RunnerOptions{ValidateHash: true})
	pending, err := runner.PendingSQL(context.Background(), source)
	if err != nil {
		t.Fatalf("PendingSQL returned error: %v", err)
	}
	if len(pending) != 1 || !pending[0].Repeatable {
		t.Fatalf("unexpected pending repeatable migrations: %+v", pending)
	}
}

func TestRunnerUpSQLAppliesVersionedFiles(t *testing.T) {
	ctx := context.Background()
	db := openSQLiteRunnerDB(t)
	runner := NewRunner(db, testDialect{}, RunnerOptions{ValidateHash: true})

	source := FileSource{
		FS: fstest.MapFS{
			"sql/V1__create_logs.sql": &fstest.MapFile{Data: []byte("CREATE TABLE logs (id INTEGER PRIMARY KEY);\n")},
		},
		Dir: "sql",
	}

	report, err := runner.UpSQL(ctx, source)
	if err != nil {
		t.Fatalf("UpSQL returned error: %v", err)
	}
	if len(report.Applied) != 1 || report.Applied[0].Version != "1" || report.Applied[0].Kind != KindSQL {
		t.Fatalf("unexpected sql migration report: %+v", report)
	}

	applied, err := runner.Applied(ctx)
	if err != nil {
		t.Fatalf("Applied returned error: %v", err)
	}
	if len(applied) != 1 || applied[0].Version != "1" || applied[0].Kind != KindSQL || !applied[0].Success {
		t.Fatalf("unexpected applied records: %+v", applied)
	}

	if !sqliteTableExists(t, db, "logs") {
		t.Fatalf("expected logs table to exist")
	}
	if !sqliteTableExists(t, db, "schema_history") {
		t.Fatalf("expected schema_history table to exist")
	}
}

func openSQLiteRunnerDB(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "runner.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open returned error: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func sqliteTableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ?)`, name).Scan(&exists)
	if err != nil {
		t.Fatalf("table exists query returned error: %v", err)
	}
	return exists
}
