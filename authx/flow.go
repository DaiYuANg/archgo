package authx

import (
	"context"
	"fmt"
	"slices"

	collectionmapping "github.com/DaiYuANg/arcgo/collectionx/mapping"
)

// AuthFlow dispatches credential authentication by credential kind.
type AuthFlow struct {
	authenticators *collectionmapping.ConcurrentMap[string, Authenticator]
}

// NewAuthFlow creates an authentication flow with fixed authenticators.
func NewAuthFlow(authenticators ...Authenticator) (*AuthFlow, error) {
	flow := &AuthFlow{
		authenticators: collectionmapping.NewConcurrentMap[string, Authenticator](),
	}

	for _, authenticator := range authenticators {
		if authenticator == nil {
			return nil, fmt.Errorf("%w: authenticator is nil", ErrInvalidAuthenticator)
		}

		kind := normalizeKind(authenticator.Kind())
		if kind == "" {
			return nil, fmt.Errorf("%w: authenticator kind is empty", ErrInvalidAuthenticator)
		}

		if _, exists := flow.authenticators.Get(kind); exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateAuthenticator, kind)
		}

		flow.authenticators.Set(kind, authenticator)
	}

	return flow, nil
}

// Authenticate authenticates a credential by its kind.
func (f *AuthFlow) Authenticate(ctx context.Context, credential Credential) (Identity, error) {
	if f == nil || f.authenticators == nil {
		return nil, fmt.Errorf("%w: flow is nil", ErrInvalidAuthenticator)
	}
	if credential == nil {
		return nil, fmt.Errorf("%w: credential is nil", ErrInvalidCredential)
	}

	kind := normalizeKind(credential.Kind())
	if kind == "" {
		return nil, fmt.Errorf("%w: credential kind is empty", ErrInvalidCredential)
	}

	authenticator := f.authenticators.GetOption(kind).OrElse(nil)
	if authenticator == nil {
		return nil, fmt.Errorf("%w: %s", ErrAuthenticatorNotFound, kind)
	}

	identity, err := authenticator.Authenticate(ctx, credential)
	if err != nil {
		return nil, err
	}
	if identity == nil {
		return nil, ErrUnauthorized
	}

	return identity, nil
}

// HasAuthenticator reports whether kind is configured in flow.
func (f *AuthFlow) HasAuthenticator(kind string) bool {
	if f == nil || f.authenticators == nil {
		return false
	}
	_, exists := f.authenticators.Get(normalizeKind(kind))
	return exists
}

// RegisteredKinds returns a sorted list of configured credential kinds.
func (f *AuthFlow) RegisteredKinds() []string {
	if f == nil || f.authenticators == nil {
		return nil
	}
	kinds := f.authenticators.Keys()
	slices.Sort(kinds)
	return kinds
}
