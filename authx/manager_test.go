package authx

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type staticPolicySource struct {
	snapshot PolicySnapshot
}

func (s staticPolicySource) LoadPolicies(ctx context.Context) (PolicySnapshot, error) {
	return s.snapshot, nil
}

func (s staticPolicySource) Name() string {
	return "static-test"
}

type unauthorizedProvider struct{}

func (unauthorizedProvider) LoadByPrincipal(ctx context.Context, principal string) (UserDetails, error) {
	_ = ctx
	_ = principal
	return UserDetails{}, ErrUnauthorized
}

type testDBUser struct {
	id       string
	password string
	name     string
}

type testDBMappedProvider struct {
	hashedPassword string
}

func (p testDBMappedProvider) LoadByPrincipal(ctx context.Context, principal string) (testDBUser, error) {
	_ = ctx
	if principal != "alice" {
		return testDBUser{}, ErrUnauthorized
	}
	return testDBUser{
		id:       "u-1",
		password: p.hashedPassword,
		name:     "Alice",
	}, nil
}

func (p testDBMappedProvider) MapToUserDetails(ctx context.Context, principal string, payload testDBUser) (UserDetails, error) {
	_ = ctx
	_ = p
	return UserDetails{
		ID:           payload.id,
		Principal:    principal,
		PasswordHash: payload.password,
		Name:         payload.name,
	}, nil
}

func newTestProvider(t *testing.T) IdentityProvider {
	t.Helper()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	provider := NewInMemoryIdentityProvider()
	err = provider.UpsertUser(UserDetails{
		ID:           "u-1",
		Principal:    "alice",
		PasswordHash: string(hashedPassword),
		Name:         "Alice",
	})
	assert.NoError(t, err)

	return provider
}

func newTestProviderWithPassword(t *testing.T, password string) IdentityProvider {
	t.Helper()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	provider := NewInMemoryIdentityProvider()
	err = provider.UpsertUser(UserDetails{
		ID:           "u-1",
		Principal:    "alice",
		PasswordHash: string(hashedPassword),
		Name:         "Alice",
	})
	assert.NoError(t, err)

	return provider
}

func TestNewManagerWithDefaults(t *testing.T) {
	manager, err := NewManager()
	assert.NoError(t, err)
	assert.NotNil(t, manager)
}

func TestManagerSetIdentityProviders(t *testing.T) {
	manager, err := NewManager(
		WithProvider(newTestProviderWithPassword(t, "secret")),
	)
	assert.NoError(t, err)

	_, _, err = manager.AuthenticatePassword(context.Background(), "alice", "secret")
	assert.NoError(t, err)

	err = manager.SetIdentityProviders(newTestProviderWithPassword(t, "new-secret"))
	assert.NoError(t, err)

	_, _, err = manager.AuthenticatePassword(context.Background(), "alice", "secret")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnauthorized))

	_, _, err = manager.AuthenticatePassword(context.Background(), "alice", "new-secret")
	assert.NoError(t, err)
}

func TestManagerProviderChainFallback(t *testing.T) {
	manager, err := NewManager(
		WithProvider(unauthorizedProvider{}),
		WithProvider(newTestProvider(t)),
	)
	assert.NoError(t, err)

	_, authentication, err := manager.AuthenticatePassword(context.Background(), "alice", "secret")
	assert.NoError(t, err)
	assert.Equal(t, "u-1", authentication.Identity().ID())
}

func TestManagerSetIdentityProvidersRejectsNil(t *testing.T) {
	manager, err := NewManager(WithProvider(newTestProvider(t)))
	assert.NoError(t, err)

	err = manager.SetIdentityProviders(nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidAuthenticator))
}

func TestManagerCanWithoutAuthentication(t *testing.T) {
	manager, err := NewManager(WithProvider(newTestProvider(t)))
	assert.NoError(t, err)

	_, err = manager.Can(context.Background(), "read", "order:1")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoIdentity))
}

func TestManagerAuthenticateAndCan(t *testing.T) {
	provider := newTestProvider(t)

	manager, err := NewManager(
		WithSource(staticPolicySource{
			snapshot: NewPolicySnapshot(
				[]PermissionRule{AllowPermission("u-1", "order:1", "read")},
				nil,
			),
		}),
		WithProvider(provider),
	)
	assert.NoError(t, err)

	version, err := manager.LoadPolicies(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), version)

	ctx, authentication, err := manager.AuthenticatePassword(context.Background(), "alice", "secret")
	assert.NoError(t, err)
	assert.True(t, authentication.IsAuthenticated())
	assert.Equal(t, "u-1", authentication.Identity().ID())

	allowed, err := manager.Can(ctx, "read", "order:1")
	assert.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = manager.Can(ctx, "write", "order:1")
	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestManagerPolicyHotReload(t *testing.T) {
	provider := newTestProvider(t)

	manager, err := NewManager(WithProvider(provider))
	assert.NoError(t, err)

	_, _, err = manager.AuthenticatePassword(context.Background(), "alice", "secret")
	assert.NoError(t, err)

	version, err := manager.ReplacePolicies(context.Background(), NewPolicySnapshot(
		[]PermissionRule{AllowPermission("u-1", "order:1", "read")},
		nil,
	))
	assert.NoError(t, err)
	assert.Equal(t, int64(1), version)

	ctx, _, err := manager.AuthenticatePassword(context.Background(), "alice", "secret")
	assert.NoError(t, err)

	allowed, err := manager.Can(ctx, "read", "order:1")
	assert.NoError(t, err)
	assert.True(t, allowed)

	version, err = manager.ReplacePolicies(context.Background(), NewPolicySnapshot(
		[]PermissionRule{DenyPermission("u-1", "order:1", "read")},
		nil,
	))
	assert.NoError(t, err)
	assert.Equal(t, int64(2), version)

	allowed, err = manager.Can(ctx, "read", "order:1")
	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestManagerLoadsFromMultipleSources(t *testing.T) {
	manager, err := NewManager(
		WithProvider(newTestProvider(t)),
		WithSource(staticPolicySource{
			snapshot: NewPolicySnapshot(
				[]PermissionRule{AllowPermission("u-1", "order:1", "read")},
				nil,
			),
		}),
		WithSource(staticPolicySource{
			snapshot: NewPolicySnapshot(
				[]PermissionRule{AllowPermission("u-1", "order:1", "write")},
				nil,
			),
		}),
	)
	assert.NoError(t, err)

	_, err = manager.LoadPolicies(context.Background())
	assert.NoError(t, err)

	ctx, _, err := manager.AuthenticatePassword(context.Background(), "alice", "secret")
	assert.NoError(t, err)

	allowedRead, err := manager.Can(ctx, "read", "order:1")
	assert.NoError(t, err)
	assert.True(t, allowedRead)

	allowedWrite, err := manager.Can(ctx, "write", "order:1")
	assert.NoError(t, err)
	assert.True(t, allowedWrite)
}

func TestManagerWithMappedProvider(t *testing.T) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	manager, err := NewManager(
		WithMappedProvider(testDBMappedProvider{hashedPassword: string(hashedPassword)}),
	)
	assert.NoError(t, err)

	ctx, auth, err := manager.AuthenticatePassword(context.Background(), "alice", "secret")
	assert.NoError(t, err)
	assert.Equal(t, "u-1", auth.Identity().ID())
	principal, ok := CurrentPrincipalAs[testDBUser](ctx)
	assert.True(t, ok)
	assert.Equal(t, "u-1", principal.id)
}
