# sqltmplx

A SQL-first conditional template renderer for Go.

Current prototype supports:

- `/*%if expr */ ... /*%end */`
- `/*%where */ ... /*%end */`
- `/*%set */ ... /*%end */`
- Doma-style bind placeholders like `/* name */'alice'`
- Doma-style slice expansion like `/* ids */(1, 2, 3)`
- expression helpers: `empty(x)`, `blank(x)`, `present(x)`
- struct binding with **field name first**, then tag aliases from `sqltmpl`, `db`, `json`
- MySQL, PostgreSQL, and SQLite bind variables
- optional render-after validation hook

## Example

```go
package main

import (
    "fmt"

    "github.com/DaiYuANg/arcgo/dbx/sqltmplx"
    "github.com/DaiYuANg/arcgo/dbx/sqltmplx/dialect"
    "github.com/DaiYuANg/arcgo/dbx/sqltmplx/validate"
)

type Query struct {
    Name string `db:"name"`
    IDs  []int  `json:"ids"`
}

func main() {
    engine := sqltmplx.New(
        dialect.Postgres{},
        sqltmplx.WithValidator(validate.NewSQLParser(dialect.Postgres{})),
    )

    tpl := `
SELECT id, name, status
FROM users
/*%where */
/*%if present(Name) */
  AND name = /* Name */'alice'
/*%end */
/*%if !empty(IDs) */
  AND id IN /* IDs */(1, 2, 3)
/*%end */
/*%end */
ORDER BY id DESC
`

    bound, err := engine.Render(tpl, Query{
        Name: "alice",
        IDs:  []int{1, 2, 3},
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(bound.Query)
    fmt.Println(bound.Args)
}
```

Expected output:

```text
SELECT id, name, status FROM users WHERE name = $1 AND id IN ($2, $3, $4) ORDER BY id DESC
[alice 1 2 3]
```

## Notes

- template control stays inside SQL comments
- placeholder samples keep the SQL executable in database tools
- rendering swaps sample literals for real bind variables
- legacy `#{...}` placeholders are intentionally removed in this iteration


## Parser-backed validation

`validate.NewSQLParser(...)` selects a no-cgo validator by dialect:

- MySQL: PingCAP parser
- PostgreSQL: wasilibs/go-pgquery (WASM + wazero by default)
- SQLite: `modernc.org/sqlite` prepare validation plus `github.com/rqlite/sql` AST parsing
