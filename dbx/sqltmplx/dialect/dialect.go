package dialect

type Dialect interface {
	BindVar(n int) string
	Name() string
}
