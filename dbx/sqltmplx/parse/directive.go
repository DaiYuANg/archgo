package parse

type Directive struct {
	If    *IfDirective    `  @@`
	Where *WhereDirective `| @@`
	Set   *SetDirective   `| @@`
	End   *EndDirective   `| @@`
}

type IfDirective struct {
	Keyword string `@"if"`
	Expr    string `@Expr`
}

type WhereDirective struct {
	Keyword string `@"where"`
}

type SetDirective struct {
	Keyword string `@"set"`
}

type EndDirective struct {
	Keyword string `@"end"`
}
