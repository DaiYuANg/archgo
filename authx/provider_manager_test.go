package authx

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type usernamePasswordCredential struct {
	Username string
	Password string
}

type phoneOTPCredential struct {
	Phone string
	Code  string
}

func TestProviderManagerRoutesMultipleTypedProviders(t *testing.T) {
	passwordProvider := NewAuthenticationProviderFunc[usernamePasswordCredential](
		func(_ context.Context, credential usernamePasswordCredential) (AuthenticationResult, error) {
			if credential.Username == "alice" && credential.Password == "secret" {
				return AuthenticationResult{Principal: Principal{ID: "alice"}}, nil
			}
			return AuthenticationResult{}, fmt.Errorf("bad credentials")
		},
	)

	phoneProvider := NewAuthenticationProviderFunc[phoneOTPCredential](
		func(_ context.Context, credential phoneOTPCredential) (AuthenticationResult, error) {
			if credential.Phone == "13800000000" && credential.Code == "123456" {
				return AuthenticationResult{Principal: Principal{ID: "phone-user"}}, nil
			}
			return AuthenticationResult{}, fmt.Errorf("bad otp")
		},
	)

	manager := NewProviderManager(passwordProvider, phoneProvider)

	res1, err := manager.Authenticate(context.Background(), usernamePasswordCredential{Username: "alice", Password: "secret"})
	require.NoError(t, err)
	principal1, ok := res1.Principal.(Principal)
	require.True(t, ok)
	assert.Equal(t, "alice", principal1.ID)

	res2, err := manager.Authenticate(context.Background(), phoneOTPCredential{Phone: "13800000000", Code: "123456"})
	require.NoError(t, err)
	principal2, ok := res2.Principal.(Principal)
	require.True(t, ok)
	assert.Equal(t, "phone-user", principal2.ID)
}

func TestProviderManagerProviderNotFound(t *testing.T) {
	manager := NewProviderManager()
	_, err := manager.Authenticate(context.Background(), struct{ Value string }{Value: "x"})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrAuthenticationProviderNotFound)
}

func TestProviderManagerVariadicRegister(t *testing.T) {
	providerA := NewAuthenticationProviderFunc[string](
		func(_ context.Context, credential string) (AuthenticationResult, error) {
			return AuthenticationResult{Principal: "A:" + credential}, nil
		},
	)
	providerB := NewAuthenticationProviderFunc[int](
		func(_ context.Context, credential int) (AuthenticationResult, error) {
			return AuthenticationResult{Principal: credential + 1}, nil
		},
	)

	manager := NewProviderManager()
	manager.Register(providerA, providerB)

	res, err := manager.Authenticate(context.Background(), 41)
	require.NoError(t, err)
	assert.Equal(t, 42, res.Principal)
}

func TestProviderManagerRejectsNilCredential(t *testing.T) {
	manager := NewProviderManager()
	_, err := manager.Authenticate(context.Background(), nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAuthenticationCredential)
}
