package authx

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorCode(t *testing.T) {
	tests := []struct {
		name string
		code ErrorCode
	}{
		{"InvalidCredential", CodeInvalidCredential},
		{"PrincipalNotFound", CodePrincipalNotFound},
		{"BadPassword", CodeBadPassword},
		{"Unauthenticated", CodeUnauthenticated},
		{"Forbidden", CodeForbidden},
		{"NoIdentity", CodeNoIdentity},
		{"InvalidPolicy", CodeInvalidPolicy},
		{"PolicyMergeConflict", CodePolicyMergeConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code == "" {
				t.Errorf("expected non-empty code for %s", tt.name)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "without wrapped error",
			err:      NewError(CodeUnauthenticated, "unauthorized"),
			expected: "authx: [unauthenticated] unauthorized",
		},
		{
			name:     "with wrapped error",
			err:      WrapError(CodeInvalidCredential, "invalid input", errors.New("missing username")),
			expected: "authx: [invalid_credential] invalid input: missing username",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	wrapped := errors.New("underlying error")
	err := WrapError(CodeBadPassword, "password mismatch", wrapped)

	if got := err.Unwrap(); got != wrapped {
		t.Errorf("Unwrap() = %v, want %v", got, wrapped)
	}
}

func TestError_Is(t *testing.T) {
	err1 := NewError(CodeForbidden, "access denied")
	err2 := NewError(CodeForbidden, "different message")
	err3 := NewError(CodeUnauthenticated, "unauthorized")

	if !errors.Is(err1, err2) {
		t.Error("errors with same code should be equal")
	}

	if errors.Is(err1, err3) {
		t.Error("errors with different code should not be equal")
	}
}

func TestNewError(t *testing.T) {
	err := NewError(CodeForbidden, "custom message")

	if err.Code != CodeForbidden {
		t.Errorf("Code = %v, want %v", err.Code, CodeForbidden)
	}
	if err.Message != "custom message" {
		t.Errorf("Message = %q, want %q", err.Message, "custom message")
	}
	if err.Err != nil {
		t.Errorf("Err = %v, want nil", err.Err)
	}
}

func TestWrapError(t *testing.T) {
	underlying := errors.New("database connection failed")
	err := WrapError(CodeProviderUnavailable, "identity provider error", underlying)

	if err.Code != CodeProviderUnavailable {
		t.Errorf("Code = %v, want %v", err.Code, CodeProviderUnavailable)
	}
	if err.Message != "identity provider error" {
		t.Errorf("Message = %q, want %q", err.Message, "identity provider error")
	}
	if !errors.Is(err.Err, underlying) {
		t.Error("wrapped error should match underlying error")
	}
}

func TestGetCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorCode
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "authx error",
			err:      NewError(CodeForbidden, "denied"),
			expected: CodeForbidden,
		},
		{
			name:     "wrapped authx error",
			err:      fmt.Errorf("wrapped: %w", NewError(CodeUnauthenticated, "unauthorized")),
			expected: CodeUnauthenticated,
		},
		{
			name:     "non-authx error",
			err:      errors.New("some other error"),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCode(tt.err); got != tt.expected {
				t.Errorf("GetCode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsCode(t *testing.T) {
	err := NewError(CodeForbidden, "access denied")

	if !IsCode(err, CodeForbidden) {
		t.Error("IsCode should return true for matching code")
	}
	if IsCode(err, CodeUnauthenticated) {
		t.Error("IsCode should return false for different code")
	}
	if IsCode(nil, CodeForbidden) {
		t.Error("IsCode should return false for nil error")
	}
}

func TestIsUnauthorized(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "unauthenticated error",
			err:      NewError(CodeUnauthenticated, "unauthorized"),
			expected: true,
		},
		{
			name:     "invalid credential error",
			err:      NewError(CodeInvalidCredential, "bad credential"),
			expected: true,
		},
		{
			name:     "principal not found error",
			err:      NewError(CodePrincipalNotFound, "user not found"),
			expected: true,
		},
		{
			name:     "bad password error",
			err:      NewError(CodeBadPassword, "wrong password"),
			expected: true,
		},
		{
			name:     "forbidden error",
			err:      NewError(CodeForbidden, "denied"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnauthorized(tt.err); got != tt.expected {
				t.Errorf("IsUnauthorized() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsForbidden(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "forbidden error",
			err:      NewError(CodeForbidden, "access denied"),
			expected: true,
		},
		{
			name:     "unauthenticated error",
			err:      NewError(CodeUnauthenticated, "unauthorized"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsForbidden(tt.err); got != tt.expected {
				t.Errorf("IsForbidden() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "principal not found error",
			err:      NewError(CodePrincipalNotFound, "user not found"),
			expected: true,
		},
		{
			name:     "authenticator not found error",
			err:      NewError(CodeAuthenticatorNotFound, "no authenticator"),
			expected: true,
		},
		{
			name:     "forbidden error",
			err:      NewError(CodeForbidden, "denied"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.expected {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLegacyErrorVariables(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorCode
	}{
		{"ErrInvalidCredential", ErrInvalidCredential, CodeInvalidCredential},
		{"ErrInvalidAuthenticator", ErrInvalidAuthenticator, CodeInvalidAuthenticator},
		{"ErrInvalidAuthorizer", ErrInvalidAuthorizer, CodeInvalidAuthorizer},
		{"ErrInvalidPolicy", ErrInvalidPolicy, CodeInvalidPolicy},
		{"ErrInvalidRequest", ErrInvalidRequest, CodeInvalidRequest},
		{"ErrAuthenticatorNotFound", ErrAuthenticatorNotFound, CodeAuthenticatorNotFound},
		{"ErrDuplicateAuthenticator", ErrDuplicateAuthenticator, CodeDuplicateAuthenticator},
		{"ErrUnauthorized", ErrUnauthorized, CodeUnauthenticated},
		{"ErrForbidden", ErrForbidden, CodeForbidden},
		{"ErrNoIdentity", ErrNoIdentity, CodeNoIdentity},
		{"ErrPrincipalNotFound", ErrPrincipalNotFound, CodePrincipalNotFound},
		{"ErrBadPassword", ErrBadPassword, CodeBadPassword},
		{"ErrProviderUnavailable", ErrProviderUnavailable, CodeProviderUnavailable},
		{"ErrPolicyMergeConflict", ErrPolicyMergeConflict, CodePolicyMergeConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCode(tt.err); got != tt.expected {
				t.Errorf("GetCode(%T) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}

func TestErrorErrorsCompatibility(t *testing.T) {
	// Test that errors work with errors.Is and errors.As
	baseErr := NewError(CodeForbidden, "base error")
	wrappedErr := fmt.Errorf("context: %w", baseErr)

	if !errors.Is(wrappedErr, baseErr) {
		t.Error("errors.Is should work with wrapped authx errors")
	}

	var authxErr *Error
	if !errors.As(wrappedErr, &authxErr) {
		t.Error("errors.As should extract authx Error from wrapped error")
	}
	if authxErr.Code != CodeForbidden {
		t.Errorf("extracted error code = %v, want %v", authxErr.Code, CodeForbidden)
	}
}
