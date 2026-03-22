package dbx

type BoundQuery struct {
	Name          string
	SQL           string
	Args          []any
	CapacityHint  int // when >0, hint for pre-allocating result slice (e.g. from LIMIT)
}
