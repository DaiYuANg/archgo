package authx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryPolicySource_Basic(t *testing.T) {
	src := NewMemoryPolicySource(MemoryPolicySourceConfig{
		Name: "test-memory",
		InitialPermissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
		},
		InitialRoleBindings: []RoleBinding{
			NewRoleBinding("bob", "admin"),
		},
	})

	ctx := context.Background()
	snapshot, err := src.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot.Permissions, 1)
	assert.Len(t, snapshot.RoleBindings, 1)
	assert.Equal(t, int64(1), src.Version())
	assert.Equal(t, "test-memory", src.Name())
}

func TestMemoryPolicySource_Update(t *testing.T) {
	src := NewMemoryPolicySource(MemoryPolicySourceConfig{})

	// Initial version should be 1
	assert.Equal(t, int64(1), src.Version())

	// Update should increment version
	newVersion := src.Update(
		[]PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
			AllowPermission("alice", "/api/users", "write"),
		},
		[]RoleBinding{
			NewRoleBinding("bob", "admin"),
		},
	)
	assert.Equal(t, int64(2), newVersion)
	assert.Equal(t, int64(2), src.Version())

	ctx := context.Background()
	snapshot, err := src.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot.Permissions, 2)
	assert.Len(t, snapshot.RoleBindings, 1)
}

func TestMemoryPolicySource_UpdateSnapshot(t *testing.T) {
	src := NewMemoryPolicySource(MemoryPolicySourceConfig{})

	snapshot := PolicySnapshot{
		Permissions: []PermissionRule{
			AllowPermission("charlie", "/api/orders", "read"),
		},
		RoleBindings: []RoleBinding{
			NewRoleBinding("diana", "viewer"),
		},
	}

	newVersion := src.UpdateSnapshot(snapshot)
	assert.Equal(t, int64(2), newVersion)

	ctx := context.Background()
	loaded, err := src.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, loaded.Permissions, 1)
	assert.Len(t, loaded.RoleBindings, 1)
	assert.Equal(t, "charlie", loaded.Permissions[0].Subject)
}

func TestMemoryPolicySource_Clear(t *testing.T) {
	src := NewMemoryPolicySource(MemoryPolicySourceConfig{
		InitialPermissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
		},
	})

	assert.Equal(t, int64(1), src.Version())

	newVersion := src.Clear()
	assert.Equal(t, int64(2), newVersion)

	ctx := context.Background()
	snapshot, err := src.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Empty(t, snapshot.Permissions)
	assert.Empty(t, snapshot.RoleBindings)
}

func TestMemoryPolicySource_AddPermission(t *testing.T) {
	src := NewMemoryPolicySource(MemoryPolicySourceConfig{})

	newVersion := src.AddPermission(AllowPermission("alice", "/api/users", "read"))
	assert.Equal(t, int64(2), newVersion)

	newVersion = src.AddPermission(AllowPermission("alice", "/api/users", "write"))
	assert.Equal(t, int64(3), newVersion)

	ctx := context.Background()
	snapshot, err := src.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot.Permissions, 2)
}

func TestMemoryPolicySource_RemovePermission(t *testing.T) {
	src := NewMemoryPolicySource(MemoryPolicySourceConfig{
		InitialPermissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
			AllowPermission("alice", "/api/users", "write"),
			DenyPermission("bob", "/api/admin", "read"),
		},
	})

	// Remove all deny rules
	newVersion := src.RemovePermission(func(p PermissionRule) bool {
		return !p.Allowed
	})
	assert.Equal(t, int64(2), newVersion)

	ctx := context.Background()
	snapshot, err := src.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot.Permissions, 2)
	for _, p := range snapshot.Permissions {
		assert.True(t, p.Allowed)
	}
}

func TestMemoryPolicySource_AddRoleBinding(t *testing.T) {
	src := NewMemoryPolicySource(MemoryPolicySourceConfig{})

	newVersion := src.AddRoleBinding(NewRoleBinding("alice", "admin"))
	assert.Equal(t, int64(2), newVersion)

	newVersion = src.AddRoleBinding(NewRoleBinding("bob", "viewer"))
	assert.Equal(t, int64(3), newVersion)

	ctx := context.Background()
	snapshot, err := src.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot.RoleBindings, 2)
}

func TestMemoryPolicySource_RemoveRoleBinding(t *testing.T) {
	src := NewMemoryPolicySource(MemoryPolicySourceConfig{
		InitialRoleBindings: []RoleBinding{
			NewRoleBinding("alice", "admin"),
			NewRoleBinding("bob", "viewer"),
			NewRoleBinding("charlie", "admin"),
		},
	})

	// Remove all admin bindings
	newVersion := src.RemoveRoleBinding(func(b RoleBinding) bool {
		return b.Role == "admin"
	})
	assert.Equal(t, int64(2), newVersion)

	ctx := context.Background()
	snapshot, err := src.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot.RoleBindings, 1)
	assert.Equal(t, "viewer", snapshot.RoleBindings[0].Role)
}

func TestMemoryPolicySource_Snapshot(t *testing.T) {
	src := NewMemoryPolicySource(MemoryPolicySourceConfig{
		InitialPermissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
		},
	})

	snapshot1 := src.Snapshot()
	_ = src.Snapshot() // snapshot2 - verify multiple calls work

	// Modify snapshot1 should not affect source
	snapshot1.Permissions = append(snapshot1.Permissions, AllowPermission("bob", "/api/admin", "read"))

	snapshot3 := src.Snapshot()
	assert.Len(t, snapshot3.Permissions, 1) // Should still be 1
}

func TestStaticPolicySource(t *testing.T) {
	src := NewStaticPolicySource(StaticPolicySourceConfig{
		Name: "test-static",
		Permissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
		},
		RoleBindings: []RoleBinding{
			NewRoleBinding("bob", "admin"),
		},
	})

	ctx := context.Background()
	snapshot1, err := src.LoadPolicies(ctx)
	assert.NoError(t, err)

	snapshot2, err := src.LoadPolicies(ctx)
	assert.NoError(t, err)

	// Both snapshots should be identical
	assert.Equal(t, snapshot1.Permissions, snapshot2.Permissions)
	assert.Equal(t, snapshot1.RoleBindings, snapshot2.RoleBindings)
	assert.Equal(t, "test-static", src.Name())
}

func TestMutablePolicySource_WithoutMutator(t *testing.T) {
	src := NewMutablePolicySource(MutablePolicySourceConfig{
		Name: "test-mutable",
		InitialSnapshot: PolicySnapshot{
			Permissions: []PermissionRule{
				AllowPermission("alice", "/api/users", "read"),
			},
		},
	})

	ctx := context.Background()
	snapshot, err := src.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot.Permissions, 1)

	// Update should change the base snapshot
	src.Update(
		[]PermissionRule{
			AllowPermission("bob", "/api/admin", "write"),
		},
		nil,
	)

	snapshot2, err := src.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot2.Permissions, 1)
	assert.Equal(t, "bob", snapshot2.Permissions[0].Subject)
}

func TestMutablePolicySource_WithMutator(t *testing.T) {
	mutatorCalled := false
	src := NewMutablePolicySource(MutablePolicySourceConfig{
		InitialSnapshot: PolicySnapshot{
			Permissions: []PermissionRule{
				AllowPermission("alice", "/api/users", "read"),
			},
		},
		Mutator: func(ctx context.Context, snapshot PolicySnapshot) (PolicySnapshot, error) {
			mutatorCalled = true
			// Add an extra permission
			snapshot.Permissions = append(snapshot.Permissions, AllowPermission("system", "/api/health", "read"))
			return snapshot, nil
		},
	})

	ctx := context.Background()
	snapshot, err := src.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.True(t, mutatorCalled)
	assert.Len(t, snapshot.Permissions, 2)
	assert.Equal(t, "system", snapshot.Permissions[1].Subject)
}
