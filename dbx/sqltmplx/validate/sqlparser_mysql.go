package validate

import (
	"strings"

	mysqlparser "github.com/pingcap/parser"
	mysqlast "github.com/pingcap/parser/ast"
)

type MySQLParser struct {
	parser *mysqlparser.Parser
}

func NewMySQLParser() *MySQLParser {
	return &MySQLParser{parser: mysqlparser.New()}
}

func (p *MySQLParser) Validate(sql string) error {
	_, err := p.parser.ParseOneStmt(sql, "", "")
	return err
}

func (p *MySQLParser) Analyze(sql string) (*Analysis, error) {
	stmt, err := p.parser.ParseOneStmt(sql, "", "")
	if err != nil {
		return nil, err
	}
	return &Analysis{
		Dialect:       "mysql",
		StatementType: mysqlStatementType(stmt),
		NormalizedSQL: normalizeWhitespace(sql),
		AST:           stmt,
	}, nil
}

func mysqlStatementType(stmt mysqlast.StmtNode) string {
	switch stmt.(type) {
	case *mysqlast.SelectStmt:
		return "SELECT"
	case *mysqlast.InsertStmt:
		return "INSERT"
	case *mysqlast.UpdateStmt:
		return "UPDATE"
	case *mysqlast.DeleteStmt:
		return "DELETE"
	case *mysqlast.SetStmt:
		return "SET"
	case *mysqlast.CreateTableStmt:
		return "CREATE_TABLE"
	case *mysqlast.AlterTableStmt:
		return "ALTER_TABLE"
	case *mysqlast.DropTableStmt:
		return "DROP_TABLE"
	default:
		return detectStatementType(strings.TrimSpace(stmt.Text()))
	}
}
