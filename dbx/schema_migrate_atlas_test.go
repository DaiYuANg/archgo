package dbx

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	atlasschema "ariga.io/atlas/sql/schema"
	_ "modernc.org/sqlite"
)

func TestCompileAtlasSchemaIncludesDerivedMetadata(t *testing.T) {
	users := MustSchema("users", advancedUserSchema{})
	compiled, err := compileAtlasSchema("sqlite", nil, "main", []SchemaResource{users})
	if err != nil {
		t.Fatalf("compileAtlasSchema returned error: %v", err)
	}
	if compiled == nil || compiled.schema == nil {
		t.Fatal("expected compiled atlas schema")
	}
	if len(compiled.schema.Tables) != 1 {
		t.Fatalf("unexpected table count: %d", len(compiled.schema.Tables))
	}
	table := compiled.schema.Tables[0]
	if table.PrimaryKey == nil || len(table.PrimaryKey.Parts) != 2 {
		t.Fatalf("expected composite primary key, got: %+v", table.PrimaryKey)
	}
	if len(table.Indexes) != 1 {
		t.Fatalf("expected one secondary index, got: %d", len(table.Indexes))
	}
	if len(table.ForeignKeys) != 1 {
		t.Fatalf("expected one derived foreign key, got: %d", len(table.ForeignKeys))
	}
	if len(table.Attrs) != 1 {
		t.Fatalf("expected one check constraint attr, got: %d", len(table.Attrs))
	}
}

func TestAtlasSplitChangesSeparatesExecutableAndManualChanges(t *testing.T) {
	users := MustSchema("users", advancedUserSchema{})
	compiled, err := compileAtlasSchema("sqlite", nil, "main", []SchemaResource{users})
	if err != nil {
		t.Fatalf("compileAtlasSchema returned error: %v", err)
	}
	compiledTable, ok := compiled.tables.Get("users")
	if !ok {
		t.Fatal("expected compiled users table")
	}
	changes := []atlasschema.Change{
		&atlasschema.ModifyTable{T: compiledTable.table, Changes: []atlasschema.Change{
			&atlasschema.AddColumn{C: atlasschema.NewStringColumn("nickname", "text")},
			&atlasschema.AddPrimaryKey{P: atlasschema.NewPrimaryKey(atlasschema.NewColumn("id"))},
		}},
	}
	current := atlasschema.New("main").AddTables(atlasschema.NewTable("users"))
	safe, manual := atlasSplitChanges(changes, compiled, current)
	if len(safe) != 1 {
		t.Fatalf("expected one executable atlas change, got: %d", len(safe))
	}
	if len(manual) != 1 {
		t.Fatalf("expected one manual action, got: %d", len(manual))
	}
	if manual[0].Kind != MigrationActionManual {
		t.Fatalf("unexpected manual action kind: %+v", manual[0])
	}
}

func TestPlanSchemaChangesWithAtlasIncludesSQLPreview(t *testing.T) {
	raw, err := sql.Open("sqlite", "file:dbx-atlas-preview?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("sql.Open returned error: %v", err)
	}
	defer raw.Close()

	core := New(raw, testSQLiteDialect{})
	users := MustSchema("users", UserSchema{})

	plan, err := core.PlanSchemaChanges(context.Background(), users)
	if err != nil {
		t.Fatalf("PlanSchemaChanges returned error: %v", err)
	}

	preview := plan.SQLPreview()
	if len(preview) == 0 {
		t.Fatal("expected atlas preview sql")
	}
	if !strings.Contains(strings.ToLower(preview[0]), "create table") {
		t.Fatalf("unexpected atlas preview sql: %+v", preview)
	}
	if !strings.Contains(strings.ToLower(strings.Join(preview, " ")), "integer not null primary key autoincrement") {
		t.Fatalf("expected sqlite autoincrement preview to use integer primary key: %+v", preview)
	}
}
