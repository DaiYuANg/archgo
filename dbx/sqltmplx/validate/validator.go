package validate

type Validator interface {
	Validate(sql string) error
}

type Func func(string) error

func (f Func) Validate(sql string) error { return f(sql) }
