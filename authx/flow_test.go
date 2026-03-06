package authx

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCredential struct {
	kind string
}

func (c testCredential) Kind() string {
	return c.kind
}

func TestFlowAuthenticateSuccess(t *testing.T) {
	authenticator, err := NewAuthenticator(
		"password",
		CredentialKindPassword,
		func(ctx context.Context, credential Credential) (Identity, error) {
			c, ok := credential.(PasswordCredential)
			if !ok {
				return nil, ErrInvalidCredential
			}
			if c.Username != "alice" || c.Password != "secret" {
				return nil, ErrUnauthorized
			}
			return NewIdentity("u-1", "user", "Alice", WithRoles("admin")), nil
		},
	)
	assert.NoError(t, err)

	flow, err := NewAuthFlow(authenticator)
	assert.NoError(t, err)

	identity, err := flow.Authenticate(context.Background(), PasswordCredential{
		Username: "alice",
		Password: "secret",
	})
	assert.NoError(t, err)
	assert.Equal(t, "u-1", identity.ID())
	assert.Equal(t, []string{"admin"}, identity.Roles())
}

func TestFlowDuplicateAuthenticatorKind(t *testing.T) {
	a1, err := NewAuthenticator("password-1", CredentialKindPassword, func(ctx context.Context, credential Credential) (Identity, error) {
		return AnonymousIdentity(), nil
	})
	assert.NoError(t, err)

	a2, err := NewAuthenticator("password-2", CredentialKindPassword, func(ctx context.Context, credential Credential) (Identity, error) {
		return AnonymousIdentity(), nil
	})
	assert.NoError(t, err)

	_, err = NewAuthFlow(a1, a2)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrDuplicateAuthenticator))
}

func TestFlowAuthenticatorNotFound(t *testing.T) {
	flow, err := NewAuthFlow()
	assert.NoError(t, err)

	_, err = flow.Authenticate(context.Background(), AnonymousCredential{})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrAuthenticatorNotFound))
}

func TestFlowInvalidCredential(t *testing.T) {
	flow, err := NewAuthFlow()
	assert.NoError(t, err)

	_, err = flow.Authenticate(context.Background(), nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidCredential))

	_, err = flow.Authenticate(context.Background(), testCredential{kind: " "})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidCredential))
}

func TestFlowRegisteredKindsSorted(t *testing.T) {
	a1, err := NewAuthenticator("z", "z-kind", func(ctx context.Context, credential Credential) (Identity, error) {
		return AnonymousIdentity(), nil
	})
	assert.NoError(t, err)

	a2, err := NewAuthenticator("a", "a-kind", func(ctx context.Context, credential Credential) (Identity, error) {
		return AnonymousIdentity(), nil
	})
	assert.NoError(t, err)

	flow, err := NewAuthFlow(a1, a2)
	assert.NoError(t, err)

	assert.Equal(t, []string{"a-kind", "z-kind"}, flow.RegisteredKinds())
	assert.True(t, flow.HasAuthenticator("A-KIND"))
	assert.False(t, flow.HasAuthenticator("missing"))
}
