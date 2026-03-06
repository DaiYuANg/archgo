package httpx

import (
	"errors"

	"github.com/samber/mo"
)

// Note.
var (
	ErrAdapterNotFound    = errors.New("httpx: adapter not found")
	ErrInvalidEndpoint    = errors.New("httpx: invalid endpoint struct")
	ErrInvalidHandlerName = errors.New("httpx: invalid handler function name")
	ErrInvalidHandlerSig  = errors.New("httpx: invalid handler signature")
	ErrRouteNotRegistered = errors.New("httpx: route not registered")
)

// Error documents related behavior.
type Error struct {
	Code    int
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

// NewError creates related functionality.
func NewError(code int, message string, err ...error) *Error {
	e := &Error{
		Code:    code,
		Message: message,
	}
	if len(err) > 0 {
		e.Err = err[0]
	}
	return e
}

// ToOption converts related values.
func (e *Error) ToOption() mo.Option[error] {
	if e == nil {
		return mo.None[error]()
	}
	return mo.Some[error](e)
}

// IsAdapterNotFound checks related state.
func IsAdapterNotFound(err error) bool {
	return errors.Is(err, ErrAdapterNotFound)
}

// IsInvalidEndpoint checks related state.
func IsInvalidEndpoint(err error) bool {
	return errors.Is(err, ErrInvalidEndpoint)
}

// IsInvalidHandlerSignature checks related state.
func IsInvalidHandlerSignature(err error) bool {
	return errors.Is(err, ErrInvalidHandlerSig)
}
