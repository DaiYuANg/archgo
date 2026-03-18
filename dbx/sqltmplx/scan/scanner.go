package scan

import (
	"fmt"
	"strings"
)

func Scan(input string) ([]Token, error) {
	var (
		tokens  []Token
		textBuf strings.Builder
	)

	flushText := func() {
		if textBuf.Len() == 0 {
			return
		}
		tokens = append(tokens, Token{Kind: Text, Value: textBuf.String()})
		textBuf.Reset()
	}

	for len(input) > 0 {
		start := strings.Index(input, "/*")
		if start < 0 {
			textBuf.WriteString(input)
			break
		}

		textBuf.WriteString(input[:start])
		input = input[start+2:]

		end := strings.Index(input, "*/")
		if end < 0 {
			return nil, fmt.Errorf("sqltmplx: unterminated directive comment")
		}

		rawBody := input[:end]
		raw := strings.TrimSpace(rawBody)
		fullComment := "/*" + rawBody + "*/"
		input = input[end+2:]

		if isTemplateDirective(raw) {
			flushText()
			tokens = append(tokens, Token{Kind: Directive, Value: raw})
			continue
		}

		textBuf.WriteString(fullComment)
	}

	flushText()
	return tokens, nil
}

func isTemplateDirective(s string) bool {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "%") {
		return false
	}
	s = strings.TrimSpace(strings.TrimPrefix(s, "%"))
	return s == "where" || s == "set" || s == "end" || strings.HasPrefix(s, "if ")
}
