package authx

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolicySourceChain_Basic(t *testing.T) {
	src1 := NewMemoryPolicySource(MemoryPolicySourceConfig{
		Name: "src1",
		InitialPermissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
		},
	})

	src2 := NewMemoryPolicySource(MemoryPolicySourceConfig{
		Name: "src2",
		InitialPermissions: []PermissionRule{
			AllowPermission("bob", "/api/admin", "write"),
		},
		InitialRoleBindings: []RoleBinding{
			NewRoleBinding("charlie", "admin"),
		},
	})

	chain := NewPolicySourceChain(PolicySourceChainConfig{
		Sources: []PolicySource{src1, src2},
		Name:    "test-chain",
	})

	assert.Equal(t, "test-chain", chain.Name())
	assert.Equal(t, 2, chain.SourceCount())

	ctx := context.Background()
	snapshot, err := chain.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot.Permissions, 2)
	assert.Len(t, snapshot.RoleBindings, 1)
}

func TestPolicySourceChain_Empty(t *testing.T) {
	chain := NewPolicySourceChain(PolicySourceChainConfig{})

	ctx := context.Background()
	snapshot, err := chain.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Empty(t, snapshot.Permissions)
	assert.Empty(t, snapshot.RoleBindings)
}

func TestPolicySourceChain_AddSource(t *testing.T) {
	chain := NewPolicySourceChain(PolicySourceChainConfig{})

	src1 := NewMemoryPolicySource(MemoryPolicySourceConfig{Name: "src1"})
	src2 := NewMemoryPolicySource(MemoryPolicySourceConfig{Name: "src2"})

	count1 := chain.AddSource(src1)
	assert.Equal(t, 1, count1)

	count2 := chain.AddSource(src2)
	assert.Equal(t, 2, count2)

	// Nil source should not be added
	count3 := chain.AddSource(nil)
	assert.Equal(t, 2, count3)
}

func TestPolicySourceChain_PrependSource(t *testing.T) {
	chain := NewPolicySourceChain(PolicySourceChainConfig{})

	src1 := NewMemoryPolicySource(MemoryPolicySourceConfig{Name: "src1"})
	src2 := NewMemoryPolicySource(MemoryPolicySourceConfig{Name: "src2"})

	chain.AddSource(src1)
	chain.PrependSource(src2)

	sources := chain.Sources()
	assert.Len(t, sources, 2)
	assert.Equal(t, "src2", sources[0].Name())
	assert.Equal(t, "src1", sources[1].Name())
}

func TestPolicySourceChain_RemoveSource(t *testing.T) {
	src1 := NewMemoryPolicySource(MemoryPolicySourceConfig{Name: "src1"})
	src2 := NewMemoryPolicySource(MemoryPolicySourceConfig{Name: "src2"})
	src3 := NewMemoryPolicySource(MemoryPolicySourceConfig{Name: "src3"})

	chain := NewPolicySourceChain(PolicySourceChainConfig{
		Sources: []PolicySource{src1, src2, src3},
	})

	// Remove middle source
	removed := chain.RemoveSource("src2")
	assert.True(t, removed)
	assert.Equal(t, 2, chain.SourceCount())

	sources := chain.Sources()
	assert.Equal(t, "src1", sources[0].Name())
	assert.Equal(t, "src3", sources[1].Name())

	// Remove non-existent source
	removed2 := chain.RemoveSource("nonexistent")
	assert.False(t, removed2)
}

func TestPolicySourceChain_SourceLoadError(t *testing.T) {
	errSource := &errorPolicySource{err: errors.New("source error")}
	src2 := NewMemoryPolicySource(MemoryPolicySourceConfig{Name: "src2"})

	chain := NewPolicySourceChain(PolicySourceChainConfig{
		Sources: []PolicySource{errSource, src2},
	})

	ctx := context.Background()
	_, err := chain.LoadPolicies(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source error")
}

func TestPolicySourceChain_WithMerger(t *testing.T) {
	src1 := NewMemoryPolicySource(MemoryPolicySourceConfig{
		InitialPermissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
			AllowPermission("alice", "/api/users", "read"), // Duplicate
		},
	})

	src2 := NewMemoryPolicySource(MemoryPolicySourceConfig{
		InitialPermissions: []PermissionRule{
			AllowPermission("bob", "/api/admin", "write"),
		},
	})

	// Use default merger which deduplicates
	chain := NewPolicySourceChain(PolicySourceChainConfig{
		Sources: []PolicySource{src1, src2},
	})

	ctx := context.Background()
	snapshot, err := chain.LoadPolicies(ctx)
	assert.NoError(t, err)
	// Default merger deduplicates, so should have 2 unique permissions
	assert.Len(t, snapshot.Permissions, 2)
}

func TestPolicySourceChain_SetMerger(t *testing.T) {
	chain := NewPolicySourceChain(PolicySourceChainConfig{})

	// Set nil merger should be ignored
	chain.SetMerger(nil)

	// Set valid merger
	merger := NewDefaultPolicyMerger()
	chain.SetMerger(merger)

	// Verify merger is set (indirectly via LoadPolicies)
	src := NewMemoryPolicySource(MemoryPolicySourceConfig{
		InitialPermissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
		},
	})
	chain.AddSource(src)

	ctx := context.Background()
	_, err := chain.LoadPolicies(ctx)
	assert.NoError(t, err)
}

func TestFallbackPolicySource_Success(t *testing.T) {
	src1 := NewMemoryPolicySource(MemoryPolicySourceConfig{
		Name: "primary",
		InitialPermissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
		},
	})

	src2 := NewMemoryPolicySource(MemoryPolicySourceConfig{
		Name: "backup",
		InitialPermissions: []PermissionRule{
			AllowPermission("bob", "/api/admin", "write"),
		},
	})

	fallback := NewFallbackPolicySource(FallbackPolicySourceConfig{
		Sources: []PolicySource{src1, src2},
		Name:    "test-fallback",
	})

	assert.Equal(t, "test-fallback", fallback.Name())

	ctx := context.Background()
	snapshot, err := fallback.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot.Permissions, 1)
	assert.Equal(t, "alice", snapshot.Permissions[0].Subject) // Should use primary
}

func TestFallbackPolicySource_Fallback(t *testing.T) {
	errSource := &errorPolicySource{err: errors.New("primary failed")}
	src2 := NewMemoryPolicySource(MemoryPolicySourceConfig{
		Name: "backup",
		InitialPermissions: []PermissionRule{
			AllowPermission("bob", "/api/admin", "write"),
		},
	})

	fallback := NewFallbackPolicySource(FallbackPolicySourceConfig{
		Sources: []PolicySource{errSource, src2},
	})

	ctx := context.Background()
	snapshot, err := fallback.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot.Permissions, 1)
	assert.Equal(t, "bob", snapshot.Permissions[0].Subject) // Should use backup
}

func TestFallbackPolicySource_AllFail(t *testing.T) {
	errSource1 := &errorPolicySource{err: errors.New("source1 failed")}
	errSource2 := &errorPolicySource{err: errors.New("source2 failed")}

	fallback := NewFallbackPolicySource(FallbackPolicySourceConfig{
		Sources: []PolicySource{errSource1, errSource2},
	})

	ctx := context.Background()
	_, err := fallback.LoadPolicies(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all fallback sources failed")
}

func TestFallbackPolicySource_Empty(t *testing.T) {
	fallback := NewFallbackPolicySource(FallbackPolicySourceConfig{})

	ctx := context.Background()
	snapshot, err := fallback.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Empty(t, snapshot.Permissions)
	assert.Empty(t, snapshot.RoleBindings)
}

func TestFallbackPolicySource_AddSource(t *testing.T) {
	fallback := NewFallbackPolicySource(FallbackPolicySourceConfig{})

	src1 := NewMemoryPolicySource(MemoryPolicySourceConfig{Name: "src1"})
	src2 := NewMemoryPolicySource(MemoryPolicySourceConfig{Name: "src2"})

	fallback.AddSource(src1)
	fallback.AddSource(src2)
	fallback.AddSource(nil) // Should be ignored

	ctx := context.Background()
	_, err := fallback.LoadPolicies(ctx)
	assert.NoError(t, err)
}

func TestConditionalPolicySource(t *testing.T) {
	src1 := NewMemoryPolicySource(MemoryPolicySourceConfig{
		Name: "src1",
		InitialPermissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
		},
	})

	src2 := NewMemoryPolicySource(MemoryPolicySourceConfig{
		Name: "src2",
		InitialPermissions: []PermissionRule{
			AllowPermission("bob", "/api/admin", "write"),
		},
	})

	useSrc1 := true
	conditional := NewConditionalPolicySource(ConditionalPolicySourceConfig{
		Condition: func(ctx context.Context) PolicySource {
			if useSrc1 {
				return src1
			}
			return src2
		},
		Sources: []PolicySource{src1, src2},
		Name:    "test-conditional",
	})

	assert.Equal(t, "test-conditional", conditional.Name())

	ctx := context.Background()

	// Should use src1
	snapshot1, err := conditional.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "alice", snapshot1.Permissions[0].Subject)

	// Switch to src2
	useSrc1 = false
	snapshot2, err := conditional.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "bob", snapshot2.Permissions[0].Subject)
}

func TestConditionalPolicySource_NilCondition(t *testing.T) {
	conditional := NewConditionalPolicySource(ConditionalPolicySourceConfig{
		Condition: nil,
	})

	ctx := context.Background()
	_, err := conditional.LoadPolicies(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no condition function configured")
}

func TestConditionalPolicySource_NilSource(t *testing.T) {
	conditional := NewConditionalPolicySource(ConditionalPolicySourceConfig{
		Condition: func(ctx context.Context) PolicySource {
			return nil
		},
	})

	ctx := context.Background()
	_, err := conditional.LoadPolicies(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "condition returned nil source")
}

func TestCachedPolicySource(t *testing.T) {
	wrapped := NewMemoryPolicySource(MemoryPolicySourceConfig{
		InitialPermissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
		},
	})

	cached := NewCachedPolicySource(CachedPolicySourceConfig{
		Wrapped: wrapped,
		Name:    "test-cached",
	})

	assert.Equal(t, "test-cached", cached.Name())
	assert.Equal(t, wrapped, cached.Wrapped())

	ctx := context.Background()

	// First load - cache miss
	snapshot1, err := cached.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot1.Permissions, 1)

	// Modify wrapped source
	wrapped.AddPermission(AllowPermission("bob", "/api/admin", "write"))

	// Second load - cache hit, should still have old data
	snapshot2, err := cached.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot2.Permissions, 1) // Still 1, not 2

	// Refresh cache
	_, err = cached.Refresh(ctx)
	assert.NoError(t, err)

	// Third load - should have new data
	snapshot3, err := cached.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot3.Permissions, 2)
}

func TestCachedPolicySource_Invalidate(t *testing.T) {
	wrapped := NewMemoryPolicySource(MemoryPolicySourceConfig{
		InitialPermissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
		},
	})

	cached := NewCachedPolicySource(CachedPolicySourceConfig{
		Wrapped: wrapped,
	})

	ctx := context.Background()

	// Load to populate cache
	_, err := cached.LoadPolicies(ctx)
	assert.NoError(t, err)

	// Invalidate cache
	cached.Invalidate()

	// Modify wrapped source
	wrapped.AddPermission(AllowPermission("bob", "/api/admin", "write"))

	// Load after invalidate - should refresh from wrapped
	snapshot, err := cached.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot.Permissions, 2)
}

func TestCachedPolicySource_NilWrapped(t *testing.T) {
	assert.Panics(t, func() {
		NewCachedPolicySource(CachedPolicySourceConfig{
			Wrapped: nil,
		})
	})
}

// errorPolicySource is a test helper that always returns an error.
type errorPolicySource struct {
	err error
}

func (s *errorPolicySource) LoadPolicies(ctx context.Context) (PolicySnapshot, error) {
	return PolicySnapshot{}, s.err
}

func (s *errorPolicySource) Name() string {
	return "error-source"
}
