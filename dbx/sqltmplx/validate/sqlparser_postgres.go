package validate

import (
	pgquery "github.com/wasilibs/go-pgquery"
)

type PostgresParser struct{}

func NewPostgresParser() *PostgresParser { return &PostgresParser{} }

func (p *PostgresParser) Validate(sql string) error {
	_, err := pgquery.Parse(sql)
	return err
}

func (p *PostgresParser) Analyze(sql string) (*Analysis, error) {
	result, err := pgquery.Parse(sql)
	if err != nil {
		return nil, err
	}
	normalized, err := pgquery.Normalize(sql)
	if err != nil {
		normalized = normalizeWhitespace(sql)
	}
	return &Analysis{
		Dialect:       "postgres",
		StatementType: detectStatementType(sql),
		NormalizedSQL: normalized,
		AST:           result,
	}, nil
}
