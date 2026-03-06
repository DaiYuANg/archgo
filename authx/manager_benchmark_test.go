package authx

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type benchmarkUnauthorizedProvider struct{}

func (benchmarkUnauthorizedProvider) LoadByPrincipal(ctx context.Context, principal string) (UserDetails, error) {
	_ = ctx
	_ = principal
	return UserDetails{}, ErrUnauthorized
}

func benchmarkProviderWithPassword(b *testing.B, password string) IdentityProvider {
	b.Helper()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		b.Fatalf("generate password hash: %v", err)
	}

	provider := NewInMemoryIdentityProvider()
	if err := provider.UpsertUser(UserDetails{
		ID:           "u-1",
		Principal:    "alice",
		PasswordHash: string(hashedPassword),
		Name:         "Alice",
	}); err != nil {
		b.Fatalf("upsert user: %v", err)
	}
	return provider
}

func benchmarkManager(b *testing.B, opts ...ManagerOption) *Manager {
	b.Helper()

	silentLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	opts = append([]ManagerOption{WithLogger(silentLogger)}, opts...)

	manager, err := NewManager(opts...)
	if err != nil {
		b.Fatalf("new manager: %v", err)
	}
	return manager
}

func BenchmarkManagerAuthenticatePasswordSingleProvider(b *testing.B) {
	manager := benchmarkManager(b, WithProvider(benchmarkProviderWithPassword(b, "secret")))
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := manager.AuthenticatePassword(ctx, "alice", "secret"); err != nil {
			b.Fatalf("authenticate: %v", err)
		}
	}
}

func BenchmarkManagerAuthenticatePasswordChainFallback(b *testing.B) {
	manager := benchmarkManager(
		b,
		WithProvider(benchmarkUnauthorizedProvider{}),
		WithProvider(benchmarkUnauthorizedProvider{}),
		WithProvider(benchmarkProviderWithPassword(b, "secret")),
	)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := manager.AuthenticatePassword(ctx, "alice", "secret"); err != nil {
			b.Fatalf("authenticate with chain fallback: %v", err)
		}
	}
}

func BenchmarkManagerCanHotPath(b *testing.B) {
	manager := benchmarkManager(
		b,
		WithProvider(benchmarkProviderWithPassword(b, "secret")),
		WithSource(staticPolicySource{
			snapshot: NewPolicySnapshot(
				[]PermissionRule{AllowPermission("u-1", "order:1", "read")},
				nil,
			),
		}),
	)
	ctx := context.Background()

	if _, err := manager.LoadPolicies(ctx); err != nil {
		b.Fatalf("load policies: %v", err)
	}
	authCtx, _, err := manager.AuthenticatePassword(ctx, "alice", "secret")
	if err != nil {
		b.Fatalf("authenticate: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		allowed, canErr := manager.Can(authCtx, "read", "order:1")
		if canErr != nil {
			b.Fatalf("authorize: %v", canErr)
		}
		if !allowed {
			b.Fatal("expected allowed decision")
		}
	}
}

func BenchmarkManagerReplacePolicies(b *testing.B) {
	manager := benchmarkManager(b, WithProvider(benchmarkProviderWithPassword(b, "secret")))
	ctx := context.Background()
	snapshot := NewPolicySnapshot(
		[]PermissionRule{
			AllowPermission("u-1", "order:1", "read"),
			DenyPermission("u-1", "order:1", "write"),
		},
		[]RoleBinding{
			NewRoleBinding("u-1", "role:admin"),
		},
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := manager.ReplacePolicies(ctx, snapshot); err != nil {
			b.Fatalf("replace policies: %v", err)
		}
	}
}

func BenchmarkManagerLoadPoliciesMultiSource(b *testing.B) {
	manager := benchmarkManager(
		b,
		WithProvider(benchmarkProviderWithPassword(b, "secret")),
		WithSource(staticPolicySource{
			snapshot: NewPolicySnapshot(
				[]PermissionRule{AllowPermission("u-1", "order:1", "read")},
				nil,
			),
		}),
		WithSource(staticPolicySource{
			snapshot: NewPolicySnapshot(
				[]PermissionRule{AllowPermission("u-1", "order:1", "write")},
				[]RoleBinding{NewRoleBinding("u-1", "role:admin")},
			),
		}),
	)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := manager.LoadPolicies(ctx); err != nil {
			b.Fatalf("load policies: %v", err)
		}
	}
}
