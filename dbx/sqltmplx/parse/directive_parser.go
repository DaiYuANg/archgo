package parse

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var directiveLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Keyword", Pattern: `\b(if|where|set|end)\b`},
	{Name: "Expr", Pattern: `[^\r\n]+`},
	{Name: "Whitespace", Pattern: `[ \t]+`},
})

var directiveParser = participle.MustBuild[Directive](
	participle.Lexer(directiveLexer),
	participle.Elide("Whitespace"),
)

var nilRegex = regexp.MustCompile(`\bnil\b`)

func parseDirective(input string) (*Directive, error) {
	raw := strings.TrimSpace(input)
	if !strings.HasPrefix(raw, "%") {
		return nil, fmt.Errorf("sqltmplx: directive %q must start with %%", raw)
	}
	input = strings.TrimSpace(strings.TrimPrefix(raw, "%"))
	d, err := directiveParser.ParseString("directive", input)
	if err != nil {
		return nil, fmt.Errorf("sqltmplx: parse directive %q: %w", input, err)
	}
	if d.If != nil {
		d.If.Expr = normalizeExpr(d.If.Expr)
	}
	return d, nil
}

func normalizeExpr(in string) string {
	in = strings.TrimSpace(in)
	return nilRegex.ReplaceAllString(in, "nil")
}
