package dbx

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"testing"

	"github.com/DaiYuANg/arcgo/dbx/internal/testsql"
)

type accountLabel string

func (l *accountLabel) Scan(src any) error {
	switch value := src.(type) {
	case string:
		*l = accountLabel(strings.ToUpper(value))
		return nil
	case []byte:
		*l = accountLabel(strings.ToUpper(string(value)))
		return nil
	default:
		return fmt.Errorf("unsupported scan type %T", src)
	}
}

func (l accountLabel) Value() (driver.Value, error) {
	return strings.ToLower(string(l)), nil
}

type AccountProfile struct {
	Nickname *string        `dbx:"nickname"`
	Bio      sql.NullString `dbx:"bio"`
}

type accountRecord struct {
	ID int64 `dbx:"id"`
	*AccountProfile
	Label accountLabel `dbx:"label"`
}

type auditFields struct {
	CreatedBy string `dbx:"created_by"`
	UpdatedBy string `dbx:"updated_by"`
}

type auditedUser struct {
	ID    int64       `dbx:"id"`
	Audit auditFields `dbx:",inline"`
}

type accountSchema struct {
	Schema[accountRecord]
	ID       Column[accountRecord, int64]          `dbx:"id,pk,auto"`
	Nickname Column[accountRecord, *string]        `dbx:"nickname,nullable"`
	Bio      Column[accountRecord, sql.NullString] `dbx:"bio,nullable"`
	Label    Column[accountRecord, accountLabel]   `dbx:"label"`
}

func TestStructMapperScansEmbeddedPointerNullableAndScanner(t *testing.T) {
	sqlDB, _, cleanup, err := testsql.Open(testsql.Plan{
		Queries: []testsql.QueryPlan{
			{
				SQL:     `SELECT "accounts"."id", "accounts"."nickname", "accounts"."bio", "accounts"."label" FROM "accounts"`,
				Columns: []string{"id", "nickname", "bio", "label"},
				Rows: [][]driver.Value{
					{int64(1), "ally", "hello", "admin"},
					{int64(2), nil, nil, "reader"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("testsql.Open returned error: %v", err)
	}
	defer cleanup()

	accounts := MustSchema("accounts", accountSchema{})
	mapper := MustStructMapper[accountRecord]()

	items, err := QueryAll(context.Background(), New(sqlDB, testSQLiteDialect{}), Select(accounts.AllColumns()...).From(accounts), mapper)
	if err != nil {
		t.Fatalf("QueryAll returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("unexpected item count: %d", len(items))
	}
	if items[0].AccountProfile == nil {
		t.Fatal("expected embedded profile to be allocated")
	}
	if items[0].Nickname == nil || *items[0].Nickname != "ally" {
		t.Fatalf("unexpected nickname: %+v", items[0].Nickname)
	}
	if !items[0].Bio.Valid || items[0].Bio.String != "hello" {
		t.Fatalf("unexpected bio: %+v", items[0].Bio)
	}
	if items[0].Label != "ADMIN" {
		t.Fatalf("unexpected custom scanner label: %q", items[0].Label)
	}
	if items[1].Nickname != nil {
		t.Fatalf("expected nil nickname, got: %+v", items[1].Nickname)
	}
	if items[1].Bio.Valid {
		t.Fatalf("expected invalid bio, got: %+v", items[1].Bio)
	}
	if items[1].Label != "READER" {
		t.Fatalf("unexpected second label: %q", items[1].Label)
	}
}

func TestMapperInsertAssignmentsWithNilEmbeddedPointerAndValuer(t *testing.T) {
	sqlDB, recorder, cleanup, err := testsql.Open(testsql.Plan{
		Execs: []testsql.ExecPlan{
			{
				SQL:          `INSERT INTO "accounts" ("nickname", "bio", "label") VALUES (?, ?, ?)`,
				Args:         []driver.Value{nil, nil, "admin"},
				RowsAffected: 1,
			},
		},
	})
	if err != nil {
		t.Fatalf("testsql.Open returned error: %v", err)
	}
	defer cleanup()

	accounts := MustSchema("accounts", accountSchema{})
	mapper := MustMapper[accountRecord](accounts)
	entity := &accountRecord{
		Label: "ADMIN",
	}

	assignments, err := mapper.InsertAssignments(accounts, entity)
	if err != nil {
		t.Fatalf("InsertAssignments returned error: %v", err)
	}
	if len(assignments) != 3 {
		t.Fatalf("unexpected assignment count: %d", len(assignments))
	}

	if _, err := Exec(context.Background(), New(sqlDB, testSQLiteDialect{}), InsertInto(accounts).Values(assignments...)); err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}
	if len(recorder.Execs) != 1 {
		t.Fatalf("unexpected exec count: %d", len(recorder.Execs))
	}
}

func TestStructMapperSupportsNamedInlineFields(t *testing.T) {
	mapper := MustStructMapper[auditedUser]()

	fields := mapper.Fields()
	if len(fields) != 3 {
		t.Fatalf("unexpected mapped field count: %d", len(fields))
	}
	createdBy, ok := mapper.FieldByColumn("created_by")
	if !ok {
		t.Fatal("expected created_by mapping")
	}
	if len(createdBy.Path) != 2 {
		t.Fatalf("expected inline field path depth=2, got: %+v", createdBy.Path)
	}
	if createdBy.Path[0] != 1 {
		t.Fatalf("unexpected inline field root path: %+v", createdBy.Path)
	}
}

func TestStructMapperScanPlanMatchesQualifiedAndCaseInsensitiveColumns(t *testing.T) {
	sqlDB, _, cleanup, err := testsql.Open(testsql.Plan{
		Queries: []testsql.QueryPlan{
			{
				SQL:     `SELECT "users"."id", COUNT(*) AS "user_count" FROM "users"`,
				Columns: []string{`"users"."id"`, `"USER_COUNT"`},
				Rows: [][]driver.Value{
					{int64(1), int64(2)},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("testsql.Open returned error: %v", err)
	}
	defer cleanup()

	type aggregateRow struct {
		ID        int64 `dbx:"id"`
		UserCount int64 `dbx:"user_count"`
	}

	users := MustSchema("users", UserSchema{})
	items, err := QueryAll(
		context.Background(),
		New(sqlDB, testSQLiteDialect{}),
		Select(users.ID, CountAll().As("user_count")).From(users),
		MustStructMapper[aggregateRow](),
	)
	if err != nil {
		t.Fatalf("QueryAll returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("unexpected item count: %d", len(items))
	}
	if items[0].ID != 1 || items[0].UserCount != 2 {
		t.Fatalf("unexpected aggregate row: %+v", items[0])
	}
}
