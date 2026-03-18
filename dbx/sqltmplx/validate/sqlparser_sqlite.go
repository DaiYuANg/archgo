package validate

import (
	"database/sql"
	"fmt"
	"strings"

	rqlitesql "github.com/rqlite/sql"
	_ "modernc.org/sqlite"
)

type SQLiteParser struct{}

func NewSQLiteParser() *SQLiteParser { return &SQLiteParser{} }

func (p *SQLiteParser) Validate(sqlText string) error {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(sqlText)
	if stmt != nil {
		_ = stmt.Close()
	}
	return err
}

func (p *SQLiteParser) Analyze(sqlText string) (*Analysis, error) {
	parser := rqlitesql.NewParser(strings.NewReader(sqlText))
	astNode, err := parser.ParseStatement()
	if err != nil {
		return nil, err
	}
	if err := p.Validate(sqlText); err != nil {
		return nil, fmt.Errorf("sqlite engine validation failed after AST parse: %w", err)
	}
	return &Analysis{
		Dialect:       "sqlite",
		StatementType: detectStatementType(sqlText),
		NormalizedSQL: normalizeWhitespace(sqlText),
		AST:           astNode,
	}, nil
}
