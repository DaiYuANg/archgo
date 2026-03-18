package sqltmplx

import "github.com/DaiYuANg/arcgo/dbx/sqltmplx/validate"

type Option func(*config)

type config struct {
	validator validate.Validator
}

func WithValidator(v validate.Validator) Option {
	return func(c *config) {
		c.validator = v
	}
}
