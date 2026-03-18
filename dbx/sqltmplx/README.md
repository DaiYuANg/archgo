# sqltmplx

A SQL-first conditional template renderer for Go.

Current prototype supports:

- `/* if expr */ ... /* end */`
- `/* where */ ... /* end */`
- `/* set */ ... /* end */`
- `#{name}`
- `#{ids*}` slice expansion
- MySQL and PostgreSQL bind variables
- optional render-after validation hook

## Example

```go
package main

import (
    "fmt"

    "github.com/DaiYuANg/sqltmplx"
    "github.com/DaiYuANg/sqltmplx/dialect"
    "github.com/DaiYuANg/sqltmplx/validate"
)

func main() {
    engine := sqltmplx.New(
        dialect.Postgres{},
        sqltmplx.WithValidator(validate.Func(func(sql string) error {
            // Plug in a real SQL parser here.
            return nil
        })),
    )

    tpl := `
SELECT id, name, status
FROM users
/* where */
/* if name != nil */
  AND name = #{name}
/* end */
/* if ids != nil */
  AND id IN (#{ids*})
/* end */
/* end */
ORDER BY id DESC
`

    bound, err := engine.Render(tpl, map[string]any{
        "name": "alice",
        "ids":  []int{1, 2, 3},
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

This is a first-pass scaffold, not a finished production library.

The design intentionally separates:

- template scanning
- directive parsing
- expression evaluation
- rendering and bind collection
- render-after SQL validation hook

That keeps it easy to evolve toward a real v1.
