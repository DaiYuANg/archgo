package validate

import (
	"testing"

	"github.com/DaiYuANg/arcgo/dbx/sqltmplx/dialect"
)

func TestNewSQLParser(t *testing.T) {
	cases := []struct {
		name string
		d    dialect.Dialect
	}{
		{name: "mysql", d: dialect.MySQL{}},
		{name: "postgres", d: dialect.Postgres{}},
		{name: "sqlite", d: dialect.SQLite{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewSQLParser(tc.d)
			if p == nil {
				t.Fatal("parser should not be nil")
			}
		})
	}
}
