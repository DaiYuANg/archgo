package authx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type hookRecorder struct {
	beforeCheck int
	afterCheck  int
	beforeCan   int
	afterCan    int
}

func (hook *hookRecorder) BeforeCheck(_ context.Context, _ any) error {
	hook.beforeCheck++
	return nil
}

func (hook *hookRecorder) AfterCheck(_ context.Context, _ any, _ AuthenticationResult, _ error) {
	hook.afterCheck++
}

func (hook *hookRecorder) BeforeCan(_ context.Context, _ AuthorizationModel) error {
	hook.beforeCan++
	return nil
}

func (hook *hookRecorder) AfterCan(_ context.Context, _ AuthorizationModel, _ Decision, _ error) {
	hook.afterCan++
}

type credentialA struct {
	ID string
}

func TestEngineCheckAndCan(t *testing.T) {
	provider := NewAuthenticationProviderFunc[credentialA](func(_ context.Context, credential credentialA) (AuthenticationResult, error) {
		return AuthenticationResult{Principal: Principal{ID: credential.ID}}, nil
	})
	manager := NewProviderManager(provider)

	engine := NewEngine(
		WithAuthenticationManager(manager),
		WithAuthorizer(AuthorizerFunc(func(_ context.Context, input AuthorizationModel) (Decision, error) {
			principal, ok := input.Principal.(Principal)
			if ok && principal.ID == "u1" && input.Action == "read" {
				return Decision{Allowed: true, PolicyID: "p1"}, nil
			}
			return Decision{Allowed: false, Reason: "deny"}, nil
		})),
	)

	authn, err := engine.Check(context.Background(), credentialA{ID: "u1"})
	require.NoError(t, err)
	principal, ok := authn.Principal.(Principal)
	require.True(t, ok)
	assert.Equal(t, "u1", principal.ID)

	decision, err := engine.Can(context.Background(), AuthorizationModel{
		Principal: authn.Principal,
		Action:    "read",
		Resource:  "/orders/1",
	})
	require.NoError(t, err)
	assert.True(t, decision.Allowed)
	assert.Equal(t, "p1", decision.PolicyID)
}

func TestEngineCheckManagerMissing(t *testing.T) {
	engine := NewEngine()
	_, err := engine.Check(context.Background(), credentialA{ID: "x"})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrAuthenticationManagerNotConfigured)
}

func TestEngineCanAuthorizerMissing(t *testing.T) {
	provider := NewAuthenticationProviderFunc[credentialA](func(_ context.Context, credential credentialA) (AuthenticationResult, error) {
		return AuthenticationResult{Principal: Principal{ID: credential.ID}}, nil
	})
	engine := NewEngine(WithAuthenticationManager(NewProviderManager(provider)))

	_, err := engine.Can(context.Background(), AuthorizationModel{
		Principal: Principal{ID: "u1"},
		Action:    "read",
		Resource:  "orders",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrAuthorizerNotConfigured)
}

func TestEngineHooks(t *testing.T) {
	hook := &hookRecorder{}
	provider := NewAuthenticationProviderFunc[credentialA](func(_ context.Context, credential credentialA) (AuthenticationResult, error) {
		return AuthenticationResult{Principal: Principal{ID: credential.ID}}, nil
	})
	engine := NewEngine(
		WithAuthenticationManager(NewProviderManager(provider)),
		WithAuthorizer(AuthorizerFunc(func(_ context.Context, _ AuthorizationModel) (Decision, error) {
			return Decision{Allowed: true}, nil
		})),
		WithHook(hook),
	)

	authn, err := engine.Check(context.Background(), credentialA{ID: "u1"})
	require.NoError(t, err)
	_, err = engine.Can(context.Background(), AuthorizationModel{
		Principal: authn.Principal,
		Action:    "read",
		Resource:  "orders",
	})
	require.NoError(t, err)

	assert.Equal(t, 1, hook.beforeCheck)
	assert.Equal(t, 1, hook.afterCheck)
	assert.Equal(t, 1, hook.beforeCan)
	assert.Equal(t, 1, hook.afterCan)
}

func TestEngineValidation(t *testing.T) {
	engine := NewEngine()

	_, err := engine.Check(context.Background(), nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAuthenticationCredential)

	_, err = engine.Can(context.Background(), AuthorizationModel{Action: "", Resource: "orders"})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAuthorizationModel)
}
