package authx

import (
	"context"
	"testing"
)

func TestCasbinAuthorizerModelTypes(t *testing.T) {
	ctx := context.Background()

	t.Run("exact match", func(t *testing.T) {
		authorizer, err := NewCasbinAuthorizer(WithCasbinModelType(CasbinModelExact))
		if err != nil {
			t.Fatalf("NewCasbinAuthorizer() error = %v", err)
		}

		perms := []PermissionRule{
			AllowPermission("user-1", "/api/orders/1", "read"),
		}
		if err := authorizer.LoadPermissions(ctx, perms...); err != nil {
			t.Fatalf("LoadPermissions() error = %v", err)
		}

		identity := NewIdentity("user-1", "user", "User1", WithAuthenticated(true))

		// Exact match should work
		decision, err := authorizer.Authorize(ctx, identity, Request{Action: "read", Resource: "/api/orders/1"})
		if err != nil {
			t.Fatalf("Authorize() error = %v", err)
		}
		if !decision.Allowed {
			t.Error("expected allow for exact match")
		}

		// Different resource should fail
		decision, err = authorizer.Authorize(ctx, identity, Request{Action: "read", Resource: "/api/orders/2"})
		if err != nil {
			t.Fatalf("Authorize() error = %v", err)
		}
		if decision.Allowed {
			t.Error("expected deny for different resource")
		}
	})

	t.Run("prefix match with keyMatch", func(t *testing.T) {
		authorizer, err := NewCasbinAuthorizer(WithCasbinModelType(CasbinModelPrefix))
		if err != nil {
			t.Fatalf("NewCasbinAuthorizer() error = %v", err)
		}

		perms := []PermissionRule{
			AllowPermission("user-1", "/api/admin/*", "read"),
		}
		if err := authorizer.LoadPermissions(ctx, perms...); err != nil {
			t.Fatalf("LoadPermissions() error = %v", err)
		}

		identity := NewIdentity("user-1", "user", "User1", WithAuthenticated(true))

		// Should match prefix
		decision, err := authorizer.Authorize(ctx, identity, Request{Action: "read", Resource: "/api/admin/users"})
		if err != nil {
			t.Fatalf("Authorize() error = %v", err)
		}
		if !decision.Allowed {
			t.Error("expected allow for prefix match")
		}

		// Should match deeper path
		decision, err = authorizer.Authorize(ctx, identity, Request{Action: "read", Resource: "/api/admin/users/123/details"})
		if err != nil {
			t.Fatalf("Authorize() error = %v", err)
		}
		if !decision.Allowed {
			t.Error("expected allow for deeper prefix match")
		}

		// Should not match non-prefix
		decision, err = authorizer.Authorize(ctx, identity, Request{Action: "read", Resource: "/api/orders/1"})
		if err != nil {
			t.Fatalf("Authorize() error = %v", err)
		}
		if decision.Allowed {
			t.Error("expected deny for non-prefix resource")
		}
	})

	t.Run("keyMatch2 with path parameters", func(t *testing.T) {
		authorizer, err := NewCasbinAuthorizer(WithCasbinModelType(CasbinModelKeyMatch2))
		if err != nil {
			t.Fatalf("NewCasbinAuthorizer() error = %v", err)
		}

		perms := []PermissionRule{
			AllowPermission("user-1", "/api/orders/:id", "read"),
		}
		if err := authorizer.LoadPermissions(ctx, perms...); err != nil {
			t.Fatalf("LoadPermissions() error = %v", err)
		}

		identity := NewIdentity("user-1", "user", "User1", WithAuthenticated(true))

		// Should match with any ID
		decision, err := authorizer.Authorize(ctx, identity, Request{Action: "read", Resource: "/api/orders/123"})
		if err != nil {
			t.Fatalf("Authorize() error = %v", err)
		}
		if !decision.Allowed {
			t.Error("expected allow for keyMatch2 with parameter")
		}

		// Should match with different ID
		decision, err = authorizer.Authorize(ctx, identity, Request{Action: "read", Resource: "/api/orders/456"})
		if err != nil {
			t.Fatalf("Authorize() error = %v", err)
		}
		if !decision.Allowed {
			t.Error("expected allow for keyMatch2 with different parameter")
		}

		// Should not match different path
		decision, err = authorizer.Authorize(ctx, identity, Request{Action: "read", Resource: "/api/users/123"})
		if err != nil {
			t.Fatalf("Authorize() error = %v", err)
		}
		if decision.Allowed {
			t.Error("expected deny for different path")
		}
	})

	t.Run("glob match", func(t *testing.T) {
		authorizer, err := NewCasbinAuthorizer(WithCasbinModelType(CasbinModelGlob))
		if err != nil {
			t.Fatalf("NewCasbinAuthorizer() error = %v", err)
		}

		perms := []PermissionRule{
			AllowPermission("user-1", "/api/*/orders/*", "read"),
		}
		if err := authorizer.LoadPermissions(ctx, perms...); err != nil {
			t.Fatalf("LoadPermissions() error = %v", err)
		}

		identity := NewIdentity("user-1", "user", "User1", WithAuthenticated(true))

		// Should match glob pattern
		decision, err := authorizer.Authorize(ctx, identity, Request{Action: "read", Resource: "/api/tenant1/orders/123"})
		if err != nil {
			t.Fatalf("Authorize() error = %v", err)
		}
		if !decision.Allowed {
			t.Error("expected allow for glob match")
		}

		// Should match different values
		decision, err = authorizer.Authorize(ctx, identity, Request{Action: "read", Resource: "/api/tenant2/orders/456"})
		if err != nil {
			t.Fatalf("Authorize() error = %v", err)
		}
		if !decision.Allowed {
			t.Error("expected allow for glob match with different values")
		}

		// Should not match wrong pattern
		decision, err = authorizer.Authorize(ctx, identity, Request{Action: "read", Resource: "/api/tenant1/users/123"})
		if err != nil {
			t.Fatalf("Authorize() error = %v", err)
		}
		if decision.Allowed {
			t.Error("expected deny for non-matching glob")
		}
	})
}

func TestCasbinAuthorizerModelType(t *testing.T) {
	t.Run("returns configured model type", func(t *testing.T) {
		authorizer, err := NewCasbinAuthorizer(WithCasbinModelType(CasbinModelKeyMatch2))
		if err != nil {
			t.Fatalf("NewCasbinAuthorizer() error = %v", err)
		}

		if authorizer.ModelType() != CasbinModelKeyMatch2 {
			t.Errorf("ModelType() = %v, want %v", authorizer.ModelType(), CasbinModelKeyMatch2)
		}
	})

	t.Run("returns exact for nil authorizer", func(t *testing.T) {
		var authorizer *CasbinAuthorizer
		if authorizer.ModelType() != CasbinModelExact {
			t.Errorf("ModelType() = %v, want %v", authorizer.ModelType(), CasbinModelExact)
		}
	})
}

func TestCasbinAuthorizerWithRoleInheritance(t *testing.T) {
	ctx := context.Background()

	t.Run("prefix match with roles", func(t *testing.T) {
		authorizer, err := NewCasbinAuthorizer(WithCasbinModelType(CasbinModelPrefix))
		if err != nil {
			t.Fatalf("NewCasbinAuthorizer() error = %v", err)
		}

		perms := []PermissionRule{
			AllowPermission("role:admin", "/api/admin/*", "write"),
		}
		if err := authorizer.LoadPermissions(ctx, perms...); err != nil {
			t.Fatalf("LoadPermissions() error = %v", err)
		}

		bindings := []RoleBinding{
			NewRoleBinding("user-1", "role:admin"),
		}
		if err := authorizer.LoadRoleBindings(ctx, bindings...); err != nil {
			t.Fatalf("LoadRoleBindings() error = %v", err)
		}

		identity := NewIdentity("user-1", "user", "User1", WithAuthenticated(true))

		// Should allow via role inheritance
		decision, err := authorizer.Authorize(ctx, identity, Request{Action: "write", Resource: "/api/admin/settings"})
		if err != nil {
			t.Fatalf("Authorize() error = %v", err)
		}
		if !decision.Allowed {
			t.Error("expected allow for role-based prefix match")
		}
	})
}
