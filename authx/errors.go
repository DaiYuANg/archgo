package authx

import (
	"errors"
	"fmt"
)

// ErrorCode represents a unique error code for authx errors.
type ErrorCode string

// Authentication error codes.
const (
	// CodeInvalidCredential indicates the credential input is invalid.
	CodeInvalidCredential ErrorCode = "invalid_credential"
	// CodePrincipalNotFound indicates the principal cannot be loaded.
	CodePrincipalNotFound ErrorCode = "principal_not_found"
	// CodeBadPassword indicates the password does not match.
	CodeBadPassword ErrorCode = "bad_password"
	// CodeUnauthenticated indicates authentication failure.
	CodeUnauthenticated ErrorCode = "unauthenticated"
	// CodeAuthenticatorNotFound indicates no matching authenticator found.
	CodeAuthenticatorNotFound ErrorCode = "authenticator_not_found"
	// CodeDuplicateAuthenticator indicates duplicate authenticator registration.
	CodeDuplicateAuthenticator ErrorCode = "duplicate_authenticator"
)

// Authorization error codes.
const (
	// CodeForbidden indicates authorization deny result.
	CodeForbidden ErrorCode = "forbidden"
	// CodeNoIdentity indicates context has no attached identity.
	CodeNoIdentity ErrorCode = "no_identity"
	// CodeInvalidRequest indicates authorization request is invalid.
	CodeInvalidRequest ErrorCode = "invalid_request"
)

// Policy error codes.
const (
	// CodeInvalidPolicy indicates policy input is invalid.
	CodeInvalidPolicy ErrorCode = "policy_invalid"
	// CodePolicyMergeConflict indicates policy merge conflict.
	CodePolicyMergeConflict ErrorCode = "policy_merge_conflict"
)

// Component error codes.
const (
	// CodeInvalidAuthenticator indicates authenticator definition is invalid.
	CodeInvalidAuthenticator ErrorCode = "invalid_authenticator"
	// CodeInvalidAuthorizer indicates authorizer definition is invalid.
	CodeInvalidAuthorizer ErrorCode = "invalid_authorizer"
	// CodeProviderUnavailable indicates identity provider is unavailable.
	CodeProviderUnavailable ErrorCode = "provider_unavailable"
)

// Error is a typed error with code and optional wrapped error.
type Error struct {
	Code    ErrorCode
	Message string
	Err     error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("authx: [%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("authx: [%s] %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error.
func (e *Error) Unwrap() error {
	return e.Err
}

// Is supports errors.Is compatibility.
func (e *Error) Is(target error) bool {
	if t, ok := target.(*Error); ok {
		return e.Code == t.Code
	}
	return false
}

// NewError creates a new typed error with code and message.
func NewError(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// WrapError wraps an existing error with code and message.
func WrapError(code ErrorCode, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// GetCode extracts ErrorCode from an error.
// Returns empty code if error is not an authx Error.
func GetCode(err error) ErrorCode {
	if err == nil {
		return ""
	}

	var authxErr *Error
	if errors.As(err, &authxErr) {
		return authxErr.Code
	}
	return ""
}

// IsCode reports whether the error has the specified code.
func IsCode(err error, code ErrorCode) bool {
	return GetCode(err) == code
}

// IsUnauthorized reports whether the error is an authentication failure.
func IsUnauthorized(err error) bool {
	return IsCode(err, CodeUnauthenticated) ||
		IsCode(err, CodeInvalidCredential) ||
		IsCode(err, CodePrincipalNotFound) ||
		IsCode(err, CodeBadPassword)
}

// IsForbidden reports whether the error is an authorization deny.
func IsForbidden(err error) bool {
	return IsCode(err, CodeForbidden)
}

// IsNotFound reports whether the error indicates a resource/principal not found.
func IsNotFound(err error) bool {
	return IsCode(err, CodePrincipalNotFound) ||
		IsCode(err, CodeAuthenticatorNotFound)
}

// Legacy error variables for backward compatibility.
// New code should use NewError() and error codes directly.
var (
	// ErrInvalidCredential is returned when credential input is invalid.
	ErrInvalidCredential = NewError(CodeInvalidCredential, "invalid credential")
	// ErrInvalidAuthenticator is returned when authenticator definition is invalid.
	ErrInvalidAuthenticator = NewError(CodeInvalidAuthenticator, "invalid authenticator")
	// ErrInvalidAuthorizer is returned when authorizer definition is invalid.
	ErrInvalidAuthorizer = NewError(CodeInvalidAuthorizer, "invalid authorizer")
	// ErrInvalidPolicy is returned when policy input is invalid.
	ErrInvalidPolicy = NewError(CodeInvalidPolicy, "invalid policy")
	// ErrInvalidRequest is returned when authorization request is invalid.
	ErrInvalidRequest = NewError(CodeInvalidRequest, "invalid authorization request")
	// ErrAuthenticatorNotFound is returned when no authenticator matches credential kind.
	ErrAuthenticatorNotFound = NewError(CodeAuthenticatorNotFound, "authenticator not found")
	// ErrDuplicateAuthenticator is returned when registering duplicate credential kind.
	ErrDuplicateAuthenticator = NewError(CodeDuplicateAuthenticator, "duplicate authenticator kind")
	// ErrUnauthorized indicates authentication failure.
	ErrUnauthorized = NewError(CodeUnauthenticated, "unauthorized")
	// ErrForbidden indicates authorization deny result.
	ErrForbidden = NewError(CodeForbidden, "forbidden")
	// ErrNoIdentity indicates context has no attached identity.
	ErrNoIdentity = NewError(CodeNoIdentity, "no identity in context")
	// ErrPrincipalNotFound indicates principal cannot be loaded.
	ErrPrincipalNotFound = NewError(CodePrincipalNotFound, "principal not found")
	// ErrBadPassword indicates password does not match.
	ErrBadPassword = NewError(CodeBadPassword, "bad password")
	// ErrProviderUnavailable indicates identity provider is unavailable.
	ErrProviderUnavailable = NewError(CodeProviderUnavailable, "provider unavailable")
	// ErrPolicyMergeConflict indicates policy merge conflict.
	ErrPolicyMergeConflict = NewError(CodePolicyMergeConflict, "policy merge conflict")
)
