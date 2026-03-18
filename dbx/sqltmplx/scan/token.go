package scan

type Kind int

const (
	Text Kind = iota + 1
	Directive
)

type Token struct {
	Kind  Kind
	Value string
}
