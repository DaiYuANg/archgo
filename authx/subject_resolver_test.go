package authx

import (
	"context"
	"errors"
	"testing"
)

func TestDefaultSubjectResolver(t *testing.T) {
	resolver := NewDefaultSubjectResolver()
	ctx := context.Background()

	t.Run("resolves from ID", func(t *testing.T) {
		identity := NewIdentity("user-123", "user", "Alice")
		subject, err := resolver.ResolveSubject(ctx, identity)
		if err != nil {
			t.Fatalf("ResolveSubject() error = %v", err)
		}
		if subject != "user-123" {
			t.Errorf("subject = %q, want %q", subject, "user-123")
		}
	})

	t.Run("falls back to Name when ID is empty", func(t *testing.T) {
		// Use WithAuthenticated(true) to ensure identity is authenticated even with empty ID
		identity := NewIdentity("", "user", "Alice", WithAuthenticated(true))
		subject, err := resolver.ResolveSubject(ctx, identity)
		if err != nil {
			t.Fatalf("ResolveSubject() error = %v", err)
		}
		if subject != "Alice" {
			t.Errorf("subject = %q, want %q", subject, "Alice")
		}
	})

	t.Run("rejects unauthenticated identity", func(t *testing.T) {
		identity := AnonymousIdentity()
		_, err := resolver.ResolveSubject(ctx, identity)
		if !IsUnauthorized(err) {
			t.Errorf("expected unauthorized error, got %v", err)
		}
	})

	t.Run("rejects nil identity", func(t *testing.T) {
		_, err := resolver.ResolveSubject(ctx, nil)
		if !IsUnauthorized(err) {
			t.Errorf("expected unauthorized error, got %v", err)
		}
	})
}

func TestTenantSubjectResolver(t *testing.T) {
	ctx := context.Background()

	t.Run("resolves with tenant from attributes", func(t *testing.T) {
		cfg := TenantSubjectResolverConfig{}
		resolver := NewTenantSubjectResolver(cfg)
		identity := NewIdentity("user-123", "user", "Alice",
			WithAttributes(map[string]string{"tenant": "tenant-a"}),
		)

		subject, err := resolver.ResolveSubject(ctx, identity)
		if err != nil {
			t.Fatalf("ResolveSubject() error = %v", err)
		}

		expected := "tenant/tenant-a/subject/user-123"
		if subject != expected {
			t.Errorf("subject = %q, want %q", subject, expected)
		}
	})

	t.Run("uses custom subject key", func(t *testing.T) {
		cfg := TenantSubjectResolverConfig{
			SubjectKey: "user",
		}
		resolver := NewTenantSubjectResolver(cfg)
		identity := NewIdentity("user-123", "user", "Alice",
			WithAttributes(map[string]string{"tenant": "tenant-a"}),
		)

		subject, err := resolver.ResolveSubject(ctx, identity)
		if err != nil {
			t.Fatalf("ResolveSubject() error = %v", err)
		}

		expected := "tenant/tenant-a/user/user-123"
		if subject != expected {
			t.Errorf("subject = %q, want %q", subject, expected)
		}
	})

	t.Run("uses custom tenant extractor", func(t *testing.T) {
		cfg := TenantSubjectResolverConfig{
			TenantExtractor: func(identity Identity) string {
				return identity.Attributes()["org"]
			},
		}
		resolver := NewTenantSubjectResolver(cfg)
		identity := NewIdentity("user-123", "user", "Alice",
			WithAttributes(map[string]string{"org": "org-abc"}),
		)

		subject, err := resolver.ResolveSubject(ctx, identity)
		if err != nil {
			t.Fatalf("ResolveSubject() error = %v", err)
		}

		expected := "tenant/org-abc/subject/user-123"
		if subject != expected {
			t.Errorf("subject = %q, want %q", subject, expected)
		}
	})

	t.Run("rejects missing tenant", func(t *testing.T) {
		cfg := TenantSubjectResolverConfig{}
		resolver := NewTenantSubjectResolver(cfg)
		identity := NewIdentity("user-123", "user", "Alice")

		_, err := resolver.ResolveSubject(ctx, identity)
		if !IsCode(err, CodeInvalidRequest) {
			t.Errorf("expected invalid request error, got %v", err)
		}
	})

	t.Run("rejects unauthenticated identity", func(t *testing.T) {
		cfg := TenantSubjectResolverConfig{}
		resolver := NewTenantSubjectResolver(cfg)

		_, err := resolver.ResolveSubject(ctx, AnonymousIdentity())
		if !IsUnauthorized(err) {
			t.Errorf("expected unauthorized error, got %v", err)
		}
	})
}

func TestPrefixSubjectResolver(t *testing.T) {
	ctx := context.Background()

	t.Run("prepends prefix", func(t *testing.T) {
		resolver := NewPrefixSubjectResolver("api")
		identity := NewIdentity("user-123", "user", "Alice")

		subject, err := resolver.ResolveSubject(ctx, identity)
		if err != nil {
			t.Fatalf("ResolveSubject() error = %v", err)
		}

		expected := "api/user-123"
		if subject != expected {
			t.Errorf("subject = %q, want %q", subject, expected)
		}
	})

	t.Run("returns identity ID when prefix is empty", func(t *testing.T) {
		resolver := NewPrefixSubjectResolver("")
		identity := NewIdentity("user-123", "user", "Alice")

		subject, err := resolver.ResolveSubject(ctx, identity)
		if err != nil {
			t.Fatalf("ResolveSubject() error = %v", err)
		}

		if subject != "user-123" {
			t.Errorf("subject = %q, want %q", subject, "user-123")
		}
	})

	t.Run("rejects unauthenticated identity", func(t *testing.T) {
		resolver := NewPrefixSubjectResolver("api")

		_, err := resolver.ResolveSubject(ctx, AnonymousIdentity())
		if !IsUnauthorized(err) {
			t.Errorf("expected unauthorized error, got %v", err)
		}
	})
}

func TestMappedSubjectResolver(t *testing.T) {
	ctx := context.Background()

	t.Run("applies custom mapper", func(t *testing.T) {
		mapper := func(ctx context.Context, identity Identity) (string, error) {
			return "custom:" + identity.ID(), nil
		}
		resolver := NewMappedSubjectResolver(mapper)
		identity := NewIdentity("user-123", "user", "Alice")

		subject, err := resolver.ResolveSubject(ctx, identity)
		if err != nil {
			t.Fatalf("ResolveSubject() error = %v", err)
		}

		expected := "custom:user-123"
		if subject != expected {
			t.Errorf("subject = %q, want %q", subject, expected)
		}
	})

	t.Run("rejects nil mapper", func(t *testing.T) {
		resolver := NewMappedSubjectResolver(nil)
		identity := NewIdentity("user-123", "user", "Alice")

		_, err := resolver.ResolveSubject(ctx, identity)
		if !IsCode(err, CodeInvalidRequest) {
			t.Errorf("expected invalid request error, got %v", err)
		}
	})
}

func TestComposedSubjectResolver(t *testing.T) {
	ctx := context.Background()

	t.Run("returns first successful resolver", func(t *testing.T) {
		resolver1 := SubjectResolverFunc(func(ctx context.Context, identity Identity) (string, error) {
			return "", errors.New("resolver1 failed")
		})
		resolver2 := SubjectResolverFunc(func(ctx context.Context, identity Identity) (string, error) {
			return "from-resolver2", nil
		})
		resolver3 := SubjectResolverFunc(func(ctx context.Context, identity Identity) (string, error) {
			return "from-resolver3", nil
		})

		composed := NewComposedSubjectResolver(resolver1, resolver2, resolver3)
		identity := NewIdentity("user-123", "user", "Alice")

		subject, err := composed.ResolveSubject(ctx, identity)
		if err != nil {
			t.Fatalf("ResolveSubject() error = %v", err)
		}

		expected := "from-resolver2"
		if subject != expected {
			t.Errorf("subject = %q, want %q", subject, expected)
		}
	})

	t.Run("skips nil resolvers", func(t *testing.T) {
		resolver := SubjectResolverFunc(func(ctx context.Context, identity Identity) (string, error) {
			return "success", nil
		})

		composed := NewComposedSubjectResolver(nil, resolver, nil)
		identity := NewIdentity("user-123", "user", "Alice")

		subject, err := composed.ResolveSubject(ctx, identity)
		if err != nil {
			t.Fatalf("ResolveSubject() error = %v", err)
		}

		if subject != "success" {
			t.Errorf("subject = %q, want %q", subject, "success")
		}
	})

	t.Run("returns error when all resolvers fail", func(t *testing.T) {
		resolver1 := SubjectResolverFunc(func(ctx context.Context, identity Identity) (string, error) {
			return "", errors.New("resolver1 failed")
		})
		resolver2 := SubjectResolverFunc(func(ctx context.Context, identity Identity) (string, error) {
			return "", errors.New("resolver2 failed")
		})

		composed := NewComposedSubjectResolver(resolver1, resolver2)
		identity := NewIdentity("user-123", "user", "Alice")

		_, err := composed.ResolveSubject(ctx, identity)
		if err == nil {
			t.Error("expected error when all resolvers fail")
		}
	})

	t.Run("rejects empty resolvers", func(t *testing.T) {
		composed := NewComposedSubjectResolver()
		identity := NewIdentity("user-123", "user", "Alice")

		_, err := composed.ResolveSubject(ctx, identity)
		if !IsCode(err, CodeInvalidRequest) {
			t.Errorf("expected invalid request error, got %v", err)
		}
	})
}

func TestSubjectResolverFunc(t *testing.T) {
	ctx := context.Background()

	t.Run("executes function", func(t *testing.T) {
		fn := SubjectResolverFunc(func(ctx context.Context, identity Identity) (string, error) {
			return "func:" + identity.ID(), nil
		})
		identity := NewIdentity("user-123", "user", "Alice")

		subject, err := fn.ResolveSubject(ctx, identity)
		if err != nil {
			t.Fatalf("ResolveSubject() error = %v", err)
		}

		expected := "func:user-123"
		if subject != expected {
			t.Errorf("subject = %q, want %q", subject, expected)
		}
	})

	t.Run("rejects nil function", func(t *testing.T) {
		var fn SubjectResolverFunc
		identity := NewIdentity("user-123", "user", "Alice")

		_, err := fn.ResolveSubject(ctx, identity)
		if !IsCode(err, CodeInvalidRequest) {
			t.Errorf("expected invalid request error, got %v", err)
		}
	})
}
