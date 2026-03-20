package dbx

import (
	"context"
	"database/sql/driver"
	"testing"

	"github.com/DaiYuANg/arcgo/dbx/internal/testsql"
	"github.com/samber/mo"
)

func BenchmarkLoadBelongsTo(b *testing.B) {
	users := MustSchema("users", relationUserSchema{})
	roles := MustSchema("roles", relationRoleSchema{})
	items := []relationUser{{ID: 1, Name: "alice", RoleID: 2}, {ID: 2, Name: "bob", RoleID: 4}}

	sqlDB, _, cleanup, err := testsql.Open(testsql.Plan{
		Queries: repeatedQueryPlans(b.N, `SELECT "roles"."id", "roles"."name" FROM "roles" WHERE "roles"."id" IN (?, ?)`, []driver.Value{int64(2), int64(4)}, []string{"id", "name"}, [][]driver.Value{{int64(2), "admin"}}),
	})
	if err != nil {
		b.Fatalf("testsql.Open returned error: %v", err)
	}
	defer cleanup()

	core := New(sqlDB, testSQLiteDialect{})
	sourceMapper := MustMapper[relationUser](users)
	targetMapper := MustMapper[relationRole](roles)
	loaded := make([]mo.Option[relationRole], len(items))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := LoadBelongsTo(context.Background(), core, items, users, sourceMapper, users.Role, roles, targetMapper, func(index int, _ *relationUser, value mo.Option[relationRole]) {
			loaded[index] = value
		}); err != nil {
			b.Fatalf("LoadBelongsTo returned error: %v", err)
		}
	}
}

func BenchmarkLoadHasMany(b *testing.B) {
	users := MustSchema("users", relationUserSchema{})
	posts := MustSchema("posts", relationPostSchema{})
	items := []relationUser{{ID: 1, Name: "alice"}, {ID: 2, Name: "bob"}}

	sqlDB, _, cleanup, err := testsql.Open(testsql.Plan{
		Queries: repeatedQueryPlans(b.N, `SELECT "posts"."id", "posts"."user_id", "posts"."title" FROM "posts" WHERE "posts"."user_id" IN (?, ?)`, []driver.Value{int64(1), int64(2)}, []string{"id", "user_id", "title"}, [][]driver.Value{{int64(100), int64(1), "first"}, {int64(101), int64(1), "second"}, {int64(200), int64(2), "third"}}),
	})
	if err != nil {
		b.Fatalf("testsql.Open returned error: %v", err)
	}
	defer cleanup()

	core := New(sqlDB, testSQLiteDialect{})
	sourceMapper := MustMapper[relationUser](users)
	targetMapper := MustMapper[relationPost](posts)
	loaded := make([][]relationPost, len(items))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := LoadHasMany(context.Background(), core, items, users, sourceMapper, users.Posts, posts, targetMapper, func(index int, _ *relationUser, value []relationPost) {
			loaded[index] = value
		}); err != nil {
			b.Fatalf("LoadHasMany returned error: %v", err)
		}
	}
}

func BenchmarkLoadManyToMany(b *testing.B) {
	users := MustSchema("users", relationUserSchema{})
	tags := MustSchema("tags", relationTagSchema{})
	items := []relationUser{{ID: 1, Name: "alice"}, {ID: 2, Name: "bob"}}

	plans := make([]testsql.QueryPlan, 0, b.N*2)
	for i := 0; i < b.N; i++ {
		plans = append(plans,
			testsql.QueryPlan{
				SQL:     `SELECT "user_tags"."user_id", "user_tags"."tag_id" FROM "user_tags" WHERE "user_tags"."user_id" IN (?, ?)`,
				Args:    []driver.Value{int64(1), int64(2)},
				Columns: []string{"user_id", "tag_id"},
				Rows:    cloneDriverRows([][]driver.Value{{int64(1), int64(10)}, {int64(1), int64(11)}, {int64(2), int64(11)}}),
			},
			testsql.QueryPlan{
				SQL:     `SELECT "tags"."id", "tags"."name" FROM "tags" WHERE "tags"."id" IN (?, ?)`,
				Args:    []driver.Value{int64(10), int64(11)},
				Columns: []string{"id", "name"},
				Rows:    cloneDriverRows([][]driver.Value{{int64(10), "red"}, {int64(11), "blue"}}),
			},
		)
	}

	sqlDB, _, cleanup, err := testsql.Open(testsql.Plan{Queries: plans})
	if err != nil {
		b.Fatalf("testsql.Open returned error: %v", err)
	}
	defer cleanup()

	core := New(sqlDB, testSQLiteDialect{})
	sourceMapper := MustMapper[relationUser](users)
	targetMapper := MustMapper[relationTag](tags)
	loaded := make([][]relationTag, len(items))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := LoadManyToMany(context.Background(), core, items, users, sourceMapper, users.Tags, tags, targetMapper, func(index int, _ *relationUser, value []relationTag) {
			loaded[index] = value
		}); err != nil {
			b.Fatalf("LoadManyToMany returned error: %v", err)
		}
	}
}
