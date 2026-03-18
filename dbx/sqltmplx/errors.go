package sqltmplx

import "errors"

var (
	ErrSpreadParamEmpty = errors.New("sqltmplx: spread parameter is empty")
	ErrSpreadParamType  = errors.New("sqltmplx: spread parameter must be slice or array")
)
