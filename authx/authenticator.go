package authx

import (
	"context"
	"fmt"
	"strings"
)

// Authenticator converts a credential into an identity.
type Authenticator interface {
	Name() string
	Kind() string
	Authenticate(ctx context.Context, credential Credential) (Identity, error)
}

// AuthenticatorFunc is a function-based authenticator.
type AuthenticatorFunc struct {
	name string
	kind string
	fn   func(ctx context.Context, credential Credential) (Identity, error)
}

// NewAuthenticator creates a function-based authenticator.
func NewAuthenticator(
	name string,
	kind string,
	fn func(ctx context.Context, credential Credential) (Identity, error),
) (*AuthenticatorFunc, error) {
	normalizedName := strings.TrimSpace(name)
	normalizedKind := normalizeKind(kind)

	switch {
	case normalizedName == "":
		return nil, fmt.Errorf("%w: name is required", ErrInvalidAuthenticator)
	case normalizedKind == "":
		return nil, fmt.Errorf("%w: kind is required", ErrInvalidAuthenticator)
	case fn == nil:
		return nil, fmt.Errorf("%w: authenticate function is nil", ErrInvalidAuthenticator)
	}

	return &AuthenticatorFunc{
		name: normalizedName,
		kind: normalizedKind,
		fn:   fn,
	}, nil
}

// Name returns authenticator name.
func (a *AuthenticatorFunc) Name() string {
	return a.name
}

// Kind returns supported credential kind.
func (a *AuthenticatorFunc) Kind() string {
	return a.kind
}

// Authenticate executes authenticator function.
func (a *AuthenticatorFunc) Authenticate(ctx context.Context, credential Credential) (Identity, error) {
	if a == nil || a.fn == nil {
		return nil, fmt.Errorf("%w: authenticator function is nil", ErrInvalidAuthenticator)
	}
	return a.fn(ctx, credential)
}

func normalizeKind(kind string) string {
	return strings.ToLower(strings.TrimSpace(kind))
}
