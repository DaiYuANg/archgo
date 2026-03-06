package authx

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithAndCurrentSecurityContext(t *testing.T) {
	identity := NewIdentity("u-1", "user", "Alice")
	authentication := NewAuthentication(identity, 3)
	securityContext := NewSecurityContext(authentication)

	ctx := WithSecurityContext(context.Background(), securityContext)

	currentSecurityContext, ok := CurrentSecurityContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "u-1", currentSecurityContext.Identity().ID())

	currentAuthentication, ok := CurrentAuthentication(ctx)
	assert.True(t, ok)
	assert.Equal(t, int64(3), currentAuthentication.PolicyVersion())
	assert.False(t, currentAuthentication.AuthenticatedAt().IsZero())
}

func TestCurrentSecurityContextFallbackToIdentity(t *testing.T) {
	identity := NewIdentity("u-1", "user", "Alice")
	ctx := context.WithValue(context.Background(), identityContextKey{}, identity)

	securityContext, ok := CurrentSecurityContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "u-1", securityContext.Identity().ID())
}

func TestRequireSecurityContextAndAuthentication(t *testing.T) {
	_, err := RequireSecurityContext(context.Background())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoIdentity))

	_, err = RequireAuthentication(context.Background())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoIdentity))
}
