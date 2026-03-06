package authx

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCasbinAuthorizer(t *testing.T) {
	authorizer, err := NewCasbinAuthorizer()
	assert.NoError(t, err)
	assert.NotNil(t, authorizer)
}

func TestCasbinAuthorizerAuthorize(t *testing.T) {
	authorizer, err := NewCasbinAuthorizer()
	assert.NoError(t, err)

	err = authorizer.LoadPermissions(context.Background(),
		AllowPermission("u-1", "order:1", "read"),
	)
	assert.NoError(t, err)

	identity := NewIdentity("u-1", "user", "Alice")
	decision, err := authorizer.Authorize(context.Background(), identity, NewRequest("read", "order:1", nil))
	assert.NoError(t, err)
	assert.True(t, decision.Allowed)

	decision, err = authorizer.Authorize(context.Background(), identity, NewRequest("write", "order:1", nil))
	assert.NoError(t, err)
	assert.False(t, decision.Allowed)
}

func TestCasbinAuthorizerRoleBindings(t *testing.T) {
	authorizer, err := NewCasbinAuthorizer()
	assert.NoError(t, err)

	err = authorizer.LoadPermissions(context.Background(),
		AllowPermission("role:admin", "order:1", "write"),
	)
	assert.NoError(t, err)

	err = authorizer.LoadRoleBindings(context.Background(),
		NewRoleBinding("u-1", "role:admin"),
	)
	assert.NoError(t, err)

	identity := NewIdentity("u-1", "user", "Alice")
	decision, err := authorizer.Authorize(context.Background(), identity, NewRequest("write", "order:1", nil))
	assert.NoError(t, err)
	assert.True(t, decision.Allowed)
}

func TestCasbinAuthorizerRejectsInvalidInput(t *testing.T) {
	authorizer, err := NewCasbinAuthorizer()
	assert.NoError(t, err)

	err = authorizer.LoadPermissions(context.Background(), PermissionRule{
		Subject:  " ",
		Resource: "order:1",
		Action:   "read",
		Allowed:  true,
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidPolicy))

	_, err = authorizer.Authorize(context.Background(), nil, NewRequest("read", "order:1", nil))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnauthorized))

	_, err = authorizer.Authorize(context.Background(), NewIdentity("u-1", "user", "Alice"), Request{
		Action:   " ",
		Resource: "order:1",
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidRequest))
}
