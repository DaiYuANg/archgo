package validate

import "github.com/DaiYuANg/arcgo/dbx/sqltmplx/dialect"

type Validator interface {
	Validate(sql string) error
}

type Analyzer interface {
	Analyze(sql string) (*Analysis, error)
}

type SQLParser interface {
	Validator
	Analyzer
}

type Analysis struct {
	Dialect       string
	StatementType string
	NormalizedSQL string
	AST           any
}

type Func func(string) error

func (f Func) Validate(sql string) error { return f(sql) }

func NewSQLParser(d dialect.Dialect) SQLParser {
	switch d.Name() {
	case "mysql":
		return NewMySQLParser()
	case "postgres":
		return NewPostgresParser()
	case "sqlite":
		return NewSQLiteParser()
	default:
		return &noopParser{dialect: d.Name()}
	}
}

type noopParser struct{ dialect string }

func (n *noopParser) Validate(_ string) error { return nil }

func (n *noopParser) Analyze(sql string) (*Analysis, error) {
	return &Analysis{
		Dialect:       n.dialect,
		StatementType: detectStatementType(sql),
		NormalizedSQL: sql,
		AST:           nil,
	}, nil
}
