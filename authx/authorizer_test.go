package authx

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestValidate(t *testing.T) {
	err := Request{Action: "read", Resource: "order:1"}.Validate()
	assert.NoError(t, err)

	err = Request{Action: " ", Resource: "order:1"}.Validate()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidRequest))

	err = Request{Action: "read", Resource: " "}.Validate()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidRequest))
}

func TestRequireSuccess(t *testing.T) {
	authorizer := AuthorizerFunc(func(ctx context.Context, identity Identity, request Request) (Decision, error) {
		return Allow("ok"), nil
	})

	identity := NewIdentity("u-1", "user", "Alice")
	request := NewRequest("read", "order:1", map[string]string{"tenant": "acme"})

	err := Require(authorizer, context.Background(), identity, request)
	assert.NoError(t, err)
}

func TestRequireDeny(t *testing.T) {
	authorizer := AuthorizerFunc(func(ctx context.Context, identity Identity, request Request) (Decision, error) {
		return Deny("forbidden"), nil
	})

	identity := NewIdentity("u-1", "user", "Alice")
	request := NewRequest("write", "order:1", nil)

	err := Require(authorizer, context.Background(), identity, request)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrForbidden))
}

func TestRequireUnauthorizedAndInvalidAuthorizer(t *testing.T) {
	request := NewRequest("read", "order:1", nil)

	err := Require(nil, context.Background(), NewIdentity("u-1", "user", "Alice"), request)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidAuthorizer))

	err = Require(AuthorizerFunc(func(ctx context.Context, identity Identity, request Request) (Decision, error) {
		return Allow("ok"), nil
	}), context.Background(), AnonymousIdentity(), request)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnauthorized))
}

func TestAuthorizerFuncNil(t *testing.T) {
	var authorizer AuthorizerFunc
	_, err := authorizer.Authorize(context.Background(), NewIdentity("u-1", "user", "Alice"), Request{
		Action:   "read",
		Resource: "order:1",
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidAuthorizer))
}
