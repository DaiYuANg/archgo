package authx

import (
	"context"
	"errors"
	"testing"
)

func TestDefaultPolicyMerger(t *testing.T) {
	merger := NewDefaultPolicyMerger()
	ctx := context.Background()

	t.Run("merges empty snapshots", func(t *testing.T) {
		result, err := merger.Merge(ctx)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}
		if len(result.Permissions) != 0 {
			t.Errorf("expected 0 permissions, got %d", len(result.Permissions))
		}
		if len(result.RoleBindings) != 0 {
			t.Errorf("expected 0 role bindings, got %d", len(result.RoleBindings))
		}
	})

	t.Run("merges single snapshot", func(t *testing.T) {
		snapshot := PolicySnapshot{
			Permissions: []PermissionRule{
				AllowPermission("user-1", "order:1", "read"),
			},
			RoleBindings: []RoleBinding{
				NewRoleBinding("user-1", "admin"),
			},
		}

		result, err := merger.Merge(ctx, snapshot)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}
		if len(result.Permissions) != 1 {
			t.Errorf("expected 1 permission, got %d", len(result.Permissions))
		}
		if len(result.RoleBindings) != 1 {
			t.Errorf("expected 1 role binding, got %d", len(result.RoleBindings))
		}
	})

	t.Run("deduplicates identical rules", func(t *testing.T) {
		perm := AllowPermission("user-1", "order:1", "read")
		snapshot1 := PolicySnapshot{
			Permissions:  []PermissionRule{perm},
			RoleBindings: []RoleBinding{},
		}
		snapshot2 := PolicySnapshot{
			Permissions:  []PermissionRule{perm},
			RoleBindings: []RoleBinding{},
		}

		result, err := merger.Merge(ctx, snapshot1, snapshot2)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}
		if len(result.Permissions) != 1 {
			t.Errorf("expected 1 permission after dedup, got %d", len(result.Permissions))
		}
	})

	t.Run("merges different rules", func(t *testing.T) {
		snapshot1 := PolicySnapshot{
			Permissions: []PermissionRule{
				AllowPermission("user-1", "order:1", "read"),
			},
			RoleBindings: []RoleBinding{},
		}
		snapshot2 := PolicySnapshot{
			Permissions: []PermissionRule{
				AllowPermission("user-1", "order:1", "write"),
			},
			RoleBindings: []RoleBinding{},
		}

		result, err := merger.Merge(ctx, snapshot1, snapshot2)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}
		if len(result.Permissions) != 2 {
			t.Errorf("expected 2 permissions, got %d", len(result.Permissions))
		}
	})
}

func TestPriorityPolicyMerger(t *testing.T) {
	ctx := context.Background()

	t.Run("higher priority overrides lower priority", func(t *testing.T) {
		merger := NewPriorityPolicyMerger(PriorityPolicyMergerConfig{})

		// Lower priority (processed first, should be overridden)
		lowPriority := PolicySnapshot{
			Permissions: []PermissionRule{
				AllowPermission("user-1", "order:1", "read"),
			},
		}
		// Higher priority (processed last, should win)
		highPriority := PolicySnapshot{
			Permissions: []PermissionRule{
				DenyPermission("user-1", "order:1", "read"),
			},
		}

		result, err := merger.Merge(ctx, lowPriority, highPriority)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}

		if len(result.Permissions) != 1 {
			t.Errorf("expected 1 permission, got %d", len(result.Permissions))
		}
		if result.Permissions[0].Allowed {
			t.Error("expected deny rule from higher priority source")
		}
	})

	t.Run("accumulates role bindings", func(t *testing.T) {
		merger := NewPriorityPolicyMerger(PriorityPolicyMergerConfig{})

		snapshot1 := PolicySnapshot{
			Permissions:  []PermissionRule{},
			RoleBindings: []RoleBinding{NewRoleBinding("user-1", "viewer")},
		}
		snapshot2 := PolicySnapshot{
			Permissions:  []PermissionRule{},
			RoleBindings: []RoleBinding{NewRoleBinding("user-1", "editor")},
		}

		result, err := merger.Merge(ctx, snapshot1, snapshot2)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}

		if len(result.RoleBindings) != 2 {
			t.Errorf("expected 2 role bindings, got %d", len(result.RoleBindings))
		}
	})
}

func TestConflictDetectingPolicyMerger(t *testing.T) {
	ctx := context.Background()

	t.Run("detects conflicting rules", func(t *testing.T) {
		var conflicts []PolicyConflict
		handler := func(ctx context.Context, conflict PolicyConflict) error {
			conflicts = append(conflicts, conflict)
			return nil
		}

		merger := NewConflictDetectingPolicyMerger(ConflictDetectingPolicyMergerConfig{
			OnConflict: handler,
		})

		snapshot1 := PolicySnapshot{
			Permissions: []PermissionRule{
				AllowPermission("user-1", "order:1", "read"),
			},
		}
		snapshot2 := PolicySnapshot{
			Permissions: []PermissionRule{
				DenyPermission("user-1", "order:1", "read"),
			},
		}

		_, err := merger.Merge(ctx, snapshot1, snapshot2)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}

		if len(conflicts) != 1 {
			t.Errorf("expected 1 conflict, got %d", len(conflicts))
		}
	})

	t.Run("stops on conflict handler error", func(t *testing.T) {
		handler := func(ctx context.Context, conflict PolicyConflict) error {
			return errors.New("conflict not allowed")
		}

		merger := NewConflictDetectingPolicyMerger(ConflictDetectingPolicyMergerConfig{
			OnConflict: handler,
		})

		snapshot1 := PolicySnapshot{
			Permissions: []PermissionRule{
				AllowPermission("user-1", "order:1", "read"),
			},
		}
		snapshot2 := PolicySnapshot{
			Permissions: []PermissionRule{
				DenyPermission("user-1", "order:1", "read"),
			},
		}

		_, err := merger.Merge(ctx, snapshot1, snapshot2)
		if err == nil {
			t.Error("expected error from conflict handler")
		}
		if !IsCode(err, CodePolicyMergeConflict) {
			t.Errorf("expected policy merge conflict error, got %v", err)
		}
	})

	t.Run("no conflicts for same effect", func(t *testing.T) {
		var conflicts []PolicyConflict
		handler := func(ctx context.Context, conflict PolicyConflict) error {
			conflicts = append(conflicts, conflict)
			return nil
		}

		merger := NewConflictDetectingPolicyMerger(ConflictDetectingPolicyMergerConfig{
			OnConflict: handler,
		})

		snapshot1 := PolicySnapshot{
			Permissions: []PermissionRule{
				AllowPermission("user-1", "order:1", "read"),
			},
		}
		snapshot2 := PolicySnapshot{
			Permissions: []PermissionRule{
				AllowPermission("user-1", "order:1", "read"),
			},
		}

		_, err := merger.Merge(ctx, snapshot1, snapshot2)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}

		if len(conflicts) != 0 {
			t.Errorf("expected 0 conflicts for same effect, got %d", len(conflicts))
		}
	})
}

func TestStrictPolicyMerger(t *testing.T) {
	merger := NewStrictPolicyMerger()
	ctx := context.Background()

	t.Run("rejects conflicting rules", func(t *testing.T) {
		snapshot1 := PolicySnapshot{
			Permissions: []PermissionRule{
				AllowPermission("user-1", "order:1", "read"),
			},
		}
		snapshot2 := PolicySnapshot{
			Permissions: []PermissionRule{
				DenyPermission("user-1", "order:1", "read"),
			},
		}

		_, err := merger.Merge(ctx, snapshot1, snapshot2)
		if err == nil {
			t.Error("expected error for conflicting rules")
		}
		if !IsCode(err, CodePolicyMergeConflict) {
			t.Errorf("expected policy merge conflict error, got %v", err)
		}
	})

	t.Run("allows non-conflicting rules", func(t *testing.T) {
		snapshot1 := PolicySnapshot{
			Permissions: []PermissionRule{
				AllowPermission("user-1", "order:1", "read"),
			},
		}
		snapshot2 := PolicySnapshot{
			Permissions: []PermissionRule{
				AllowPermission("user-1", "order:1", "write"),
			},
		}

		result, err := merger.Merge(ctx, snapshot1, snapshot2)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}
		if len(result.Permissions) != 2 {
			t.Errorf("expected 2 permissions, got %d", len(result.Permissions))
		}
	})
}

func TestDenyOverridesPolicyMerger(t *testing.T) {
	merger := NewDenyOverridesPolicyMerger()
	ctx := context.Background()

	t.Run("deny overrides allow", func(t *testing.T) {
		snapshot1 := PolicySnapshot{
			Permissions: []PermissionRule{
				AllowPermission("user-1", "order:1", "read"),
			},
		}
		snapshot2 := PolicySnapshot{
			Permissions: []PermissionRule{
				DenyPermission("user-1", "order:1", "read"),
			},
		}

		result, err := merger.Merge(ctx, snapshot1, snapshot2)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}

		if len(result.Permissions) != 1 {
			t.Errorf("expected 1 permission, got %d", len(result.Permissions))
		}
		if result.Permissions[0].Allowed {
			t.Error("expected deny to override allow")
		}
	})

	t.Run("allow does not override deny", func(t *testing.T) {
		snapshot1 := PolicySnapshot{
			Permissions: []PermissionRule{
				DenyPermission("user-1", "order:1", "read"),
			},
		}
		snapshot2 := PolicySnapshot{
			Permissions: []PermissionRule{
				AllowPermission("user-1", "order:1", "read"),
			},
		}

		result, err := merger.Merge(ctx, snapshot1, snapshot2)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}

		if len(result.Permissions) != 1 {
			t.Errorf("expected 1 permission, got %d", len(result.Permissions))
		}
		if result.Permissions[0].Allowed {
			t.Error("expected deny to persist over allow")
		}
	})

	t.Run("merges role bindings", func(t *testing.T) {
		snapshot1 := PolicySnapshot{
			Permissions:  []PermissionRule{},
			RoleBindings: []RoleBinding{NewRoleBinding("user-1", "viewer")},
		}
		snapshot2 := PolicySnapshot{
			Permissions:  []PermissionRule{},
			RoleBindings: []RoleBinding{NewRoleBinding("user-1", "editor")},
		}

		result, err := merger.Merge(ctx, snapshot1, snapshot2)
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}

		if len(result.RoleBindings) != 2 {
			t.Errorf("expected 2 role bindings, got %d", len(result.RoleBindings))
		}
	})
}

func TestPolicyMergerFunc(t *testing.T) {
	ctx := context.Background()

	t.Run("executes function", func(t *testing.T) {
		fn := PolicyMergerFunc(func(ctx context.Context, snapshots ...PolicySnapshot) (PolicySnapshot, error) {
			return PolicySnapshot{
				Permissions:  []PermissionRule{AllowPermission("user-1", "order:1", "read")},
				RoleBindings: []RoleBinding{NewRoleBinding("user-1", "admin")},
			}, nil
		})

		result, err := fn.Merge(ctx, PolicySnapshot{})
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}
		if len(result.Permissions) != 1 {
			t.Errorf("expected 1 permission, got %d", len(result.Permissions))
		}
		if len(result.RoleBindings) != 1 {
			t.Errorf("expected 1 role binding, got %d", len(result.RoleBindings))
		}
	})

	t.Run("rejects nil function", func(t *testing.T) {
		var fn PolicyMergerFunc
		_, err := fn.Merge(ctx, PolicySnapshot{})
		if err == nil {
			t.Error("expected error for nil function")
		}
	})
}

func TestPermissionKey(t *testing.T) {
	tests := []struct {
		name     string
		perm     PermissionRule
		expected string
	}{
		{
			name:     "allow rule",
			perm:     AllowPermission("user-1", "order:1", "read"),
			expected: "user-1|order:1|read",
		},
		{
			name:     "deny rule",
			perm:     DenyPermission("user-1", "order:1", "read"),
			expected: "user-1|order:1|read",
		},
		{
			name:     "trims whitespace",
			perm:     AllowPermission("  user-1  ", "  order:1  ", "  read  "),
			expected: "user-1|order:1|read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := permissionKey(tt.perm)
			if got != tt.expected {
				t.Errorf("permissionKey() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestRoleBindingKey(t *testing.T) {
	tests := []struct {
		name     string
		binding  RoleBinding
		expected string
	}{
		{
			name:     "basic binding",
			binding:  NewRoleBinding("user-1", "admin"),
			expected: "user-1|admin",
		},
		{
			name:     "trims whitespace",
			binding:  NewRoleBinding("  user-1  ", "  admin  "),
			expected: "user-1|admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := roleBindingKey(tt.binding)
			if got != tt.expected {
				t.Errorf("roleBindingKey() = %q, want %q", got, tt.expected)
			}
		})
	}
}
