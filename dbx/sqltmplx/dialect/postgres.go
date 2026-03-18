package dialect

import "strconv"

type Postgres struct{}

func (Postgres) BindVar(n int) string {
	return "$" + strconv.Itoa(n)
}
