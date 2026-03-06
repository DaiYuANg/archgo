package authx

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type fakeIdentityProvider struct {
	users map[string]UserDetails
}

func (s *fakeIdentityProvider) LoadByPrincipal(ctx context.Context, principal string) (UserDetails, error) {
	_ = ctx
	user, ok := s.users[principal]
	if !ok {
		return UserDetails{}, ErrUnauthorized
	}
	return user, nil
}

func TestNewAuthbossPasswordAuthenticator(t *testing.T) {
	_, err := NewAuthbossPasswordAuthenticator(nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidAuthenticator))
}

func TestAuthbossPasswordAuthenticatorAuthenticate(t *testing.T) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	service := &fakeIdentityProvider{
		users: map[string]UserDetails{
			"alice": {
				ID:           "u-1",
				Principal:    "alice",
				PasswordHash: string(hashedPassword),
				Name:         "Alice",
				Roles:        []string{"admin"},
				Permissions:  []string{"order:read"},
				Attributes: map[string]string{
					"tenant": "acme",
				},
			},
		},
	}

	authenticator, err := NewAuthbossPasswordAuthenticator(service)
	assert.NoError(t, err)
	assert.Equal(t, CredentialKindPassword, authenticator.Kind())

	identity, err := authenticator.Authenticate(context.Background(), PasswordCredential{
		Username: "alice",
		Password: "secret",
	})
	assert.NoError(t, err)
	assert.Equal(t, "u-1", identity.ID())
	assert.Equal(t, "Alice", identity.Name())
	assert.Equal(t, []string{"admin"}, identity.Roles())
	assert.Equal(t, []string{"order:read"}, identity.Permissions())
	assert.Equal(t, "acme", identity.Attributes()["tenant"])

	_, err = authenticator.Authenticate(context.Background(), PasswordCredential{
		Username: "alice",
		Password: "wrong",
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnauthorized))

	_, err = authenticator.Authenticate(context.Background(), APIKeyCredential{Key: "abc"})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidCredential))
}

func TestAuthbossPasswordAuthenticatorRejectsInvalidUserDetails(t *testing.T) {
	service := &fakeIdentityProvider{
		users: map[string]UserDetails{
			"alice": {
				ID:        "u-1",
				Principal: "alice",
			},
		},
	}

	authenticator, err := NewAuthbossPasswordAuthenticator(service)
	assert.NoError(t, err)

	_, err = authenticator.Authenticate(context.Background(), PasswordCredential{
		Username: "alice",
		Password: "secret",
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidAuthenticator))
}
