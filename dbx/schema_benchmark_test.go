package dbx

import (
	"context"
	"strings"
	"testing"
)

func BenchmarkCompileAtlasSchema(b *testing.B) {
	roles := MustSchema("roles", advancedRoleSchema{})
	users := MustSchema("users", advancedUserSchema{})
	schemas := []SchemaResource{roles, users}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := compileAtlasSchema("sqlite", nil, "main", schemas); err != nil {
			b.Fatalf("compileAtlasSchema returned error: %v", err)
		}
	}
}

func BenchmarkPlanSchemaChangesSQLiteAtlasEmpty(b *testing.B) {
	ctx := context.Background()
	roles := MustSchema("roles", RoleSchema{})
	users := MustSchema("users", UserSchema{})
	core := New(benchmarkOpenSQLiteDB(b, "plan-empty"), testSQLiteDialect{})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := core.PlanSchemaChanges(ctx, roles, users); err != nil {
			b.Fatalf("PlanSchemaChanges returned error: %v", err)
		}
	}
}

func BenchmarkValidateSchemasSQLiteAtlasMatched(b *testing.B) {
	ctx := context.Background()
	roles := MustSchema("roles", RoleSchema{})
	users := MustSchema("users", UserSchema{})
	core := New(benchmarkOpenSQLiteDB(b, "validate-matched"), testSQLiteDialect{})
	if _, err := core.AutoMigrate(ctx, roles, users); err != nil {
		b.Fatalf("AutoMigrate returned error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := core.ValidateSchemas(ctx, roles, users); err != nil {
			b.Fatalf("ValidateSchemas returned error: %v", err)
		}
	}
}

func BenchmarkMigrationPlanSQLPreview(b *testing.B) {
	ctx := context.Background()
	roles := MustSchema("roles", RoleSchema{})
	users := MustSchema("users", UserSchema{})
	core := New(benchmarkOpenSQLiteDB(b, "preview"), testSQLiteDialect{})
	plan, err := core.PlanSchemaChanges(ctx, roles, users)
	if err != nil {
		b.Fatalf("PlanSchemaChanges returned error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		preview := plan.SQLPreview()
		if len(preview) == 0 || !strings.Contains(strings.ToLower(preview[0]), "create table") {
			b.Fatalf("unexpected preview: %+v", preview)
		}
	}
}
