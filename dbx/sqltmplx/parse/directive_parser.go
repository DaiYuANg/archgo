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
	// keep expression text normalized in one place.
	return nilRegex.ReplaceAllString(in, "nil")
}
