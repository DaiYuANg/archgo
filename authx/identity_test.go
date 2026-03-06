package authx

import "testing"

import "github.com/stretchr/testify/assert"

func TestIdentityIsImmutableByConstructionAndRead(t *testing.T) {
	attrs := map[string]string{
		"tenant": "acme",
	}

	identity := NewIdentity(
		"u-100",
		"user",
		"Alice",
		WithRoles("admin", "admin"),
		WithPermissions("order:read", "order:read"),
		WithAttributes(attrs),
	)

	attrs["tenant"] = "changed"

	assert.Equal(t, "u-100", identity.ID())
	assert.Equal(t, "user", identity.Type())
	assert.Equal(t, "Alice", identity.Name())
	assert.Equal(t, []string{"admin"}, identity.Roles())
	assert.Equal(t, []string{"order:read"}, identity.Permissions())
	assert.Equal(t, "acme", identity.Attributes()["tenant"])

	roles := identity.Roles()
	roles[0] = "hacked"
	assert.Equal(t, []string{"admin"}, identity.Roles())

	permissions := identity.Permissions()
	permissions[0] = "hacked"
	assert.Equal(t, []string{"order:read"}, identity.Permissions())

	gotAttrs := identity.Attributes()
	gotAttrs["tenant"] = "hacked"
	assert.Equal(t, "acme", identity.Attributes()["tenant"])
}

func TestAnonymousIdentity(t *testing.T) {
	identity := AnonymousIdentity()

	assert.Equal(t, "", identity.ID())
	assert.Equal(t, "anonymous", identity.Type())
	assert.False(t, identity.IsAuthenticated())
}
