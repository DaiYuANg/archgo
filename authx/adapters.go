package authx

import (
	"context"
	"reflect"
)

// TypedAuthenticationProvider keeps credential strongly typed while exposing a non-generic provider surface.
type TypedAuthenticationProvider[C any] interface {
	Authenticate(ctx context.Context, credential C) (AuthenticationResult, error)
}

// TypedAuthenticationProviderFunc is a lightweight typed provider helper.
type TypedAuthenticationProviderFunc[C any] func(ctx context.Context, credential C) (AuthenticationResult, error)

func (fn TypedAuthenticationProviderFunc[C]) Authenticate(
	ctx context.Context,
	credential C,
) (AuthenticationResult, error) {
	if fn == nil {
		return AuthenticationResult{}, ErrUnauthenticated
	}
	return fn(ctx, credential)
}

// NewAuthenticationProvider wraps a typed provider into a manager-compatible provider.
func NewAuthenticationProvider[C any](provider TypedAuthenticationProvider[C]) AuthenticationProvider {
	return &typedProviderAdapter[C]{
		provider:       provider,
		credentialType: reflect.TypeOf((*C)(nil)).Elem(),
	}
}

// NewAuthenticationProviderFunc wraps a typed function into a manager-compatible provider.
func NewAuthenticationProviderFunc[C any](
	fn func(ctx context.Context, credential C) (AuthenticationResult, error),
) AuthenticationProvider {
	return NewAuthenticationProvider[C](TypedAuthenticationProviderFunc[C](fn))
}

type typedProviderAdapter[C any] struct {
	provider       TypedAuthenticationProvider[C]
	credentialType reflect.Type
}

func (adapter *typedProviderAdapter[C]) CredentialType() reflect.Type {
	return adapter.credentialType
}

func (adapter *typedProviderAdapter[C]) AuthenticateAny(
	ctx context.Context,
	credential any,
) (AuthenticationResult, error) {
	typedCredential, ok := credential.(C)
	if !ok {
		return AuthenticationResult{}, ErrInvalidAuthenticationCredential
	}
	return adapter.provider.Authenticate(ctx, typedCredential)
}

// AuthenticationManagerFunc is a lightweight manager helper.
type AuthenticationManagerFunc func(ctx context.Context, credential any) (AuthenticationResult, error)

func (fn AuthenticationManagerFunc) Authenticate(
	ctx context.Context,
	credential any,
) (AuthenticationResult, error) {
	if fn == nil {
		return AuthenticationResult{}, ErrAuthenticationManagerNotConfigured
	}
	return fn(ctx, credential)
}

// AuthorizerFunc is a lightweight authorizer helper.
type AuthorizerFunc func(ctx context.Context, input AuthorizationModel) (Decision, error)

func (fn AuthorizerFunc) Authorize(ctx context.Context, input AuthorizationModel) (Decision, error) {
	if fn == nil {
		return Decision{}, ErrAuthorizerNotConfigured
	}
	return fn(ctx, input)
}
