package dbx

import (
	"context"
	"database/sql"
	"testing"
)

func BenchmarkSQLList(b *testing.B) {
	statement := NewSQLStatement("user.list", func(_ any) (BoundQuery, error) {
		return BoundQuery{SQL: `SELECT "id", "username" FROM "users" WHERE "status" = ?`, Args: []any{int64(1)}}, nil
	})
	dataSQL := []string{
		`INSERT INTO "roles" ("id","name") VALUES (1,'r')`,
		`INSERT INTO "users" ("username","email_address","status","role_id") VALUES ('alice','a@x.com',1,1),('bob','b@x.com',1,1)`,
	}

	run := func(b *testing.B, sqlDB *sql.DB) {
		db := New(sqlDB, testSQLiteDialect{})
		mapper := MustStructMapper[UserSummary]()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := SQLList(context.Background(), db, statement, nil, mapper); err != nil {
				b.Fatalf("SQLList returned error: %v", err)
			}
		}
	}

	b.Run("Memory", func(b *testing.B) {
		sqlDB, cleanup := OpenBenchmarkSQLiteMemoryWithSchema(b, dataSQL...)
		defer cleanup()
		run(b, sqlDB)
	})
	b.Run("IO", func(b *testing.B) {
		sqlDB, cleanup := OpenBenchmarkSQLiteWithSchema(b, dataSQL...)
		defer cleanup()
		run(b, sqlDB)
	})
}

func BenchmarkSQLGet(b *testing.B) {
	statement := NewSQLStatement("user.get", func(_ any) (BoundQuery, error) {
		return BoundQuery{SQL: `SELECT "id", "username" FROM "users" WHERE "id" = ?`, Args: []any{int64(1)}}, nil
	})
	dataSQL := []string{
		`INSERT INTO "roles" ("id","name") VALUES (1,'r')`,
		`INSERT INTO "users" ("id","username","email_address","status","role_id") VALUES (1,'alice','a@x.com',1,1)`,
	}

	run := func(b *testing.B, sqlDB *sql.DB) {
		db := New(sqlDB, testSQLiteDialect{})
		mapper := MustStructMapper[UserSummary]()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := SQLGet(context.Background(), db, statement, nil, mapper); err != nil {
				b.Fatalf("SQLGet returned error: %v", err)
			}
		}
	}

	b.Run("Memory", func(b *testing.B) {
		sqlDB, cleanup := OpenBenchmarkSQLiteMemoryWithSchema(b, dataSQL...)
		defer cleanup()
		run(b, sqlDB)
	})
	b.Run("IO", func(b *testing.B) {
		sqlDB, cleanup := OpenBenchmarkSQLiteWithSchema(b, dataSQL...)
		defer cleanup()
		run(b, sqlDB)
	})
}

func BenchmarkSQLFind(b *testing.B) {
	statement := NewSQLStatement("user.find", func(_ any) (BoundQuery, error) {
		return BoundQuery{SQL: `SELECT "id", "username" FROM "users" WHERE "id" = ?`, Args: []any{int64(1)}}, nil
	})
	dataSQL := []string{
		`INSERT INTO "roles" ("id","name") VALUES (1,'r')`,
		`INSERT INTO "users" ("id","username","email_address","status","role_id") VALUES (1,'alice','a@x.com',1,1)`,
	}

	run := func(b *testing.B, sqlDB *sql.DB) {
		db := New(sqlDB, testSQLiteDialect{})
		mapper := MustStructMapper[UserSummary]()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result, err := SQLFind(context.Background(), db, statement, nil, mapper)
			if err != nil {
				b.Fatalf("SQLFind returned error: %v", err)
			}
			if result.IsAbsent() {
				b.Fatal("expected result to be present")
			}
		}
	}

	b.Run("Memory", func(b *testing.B) {
		sqlDB, cleanup := OpenBenchmarkSQLiteMemoryWithSchema(b, dataSQL...)
		defer cleanup()
		run(b, sqlDB)
	})
	b.Run("IO", func(b *testing.B) {
		sqlDB, cleanup := OpenBenchmarkSQLiteWithSchema(b, dataSQL...)
		defer cleanup()
		run(b, sqlDB)
	})
}
