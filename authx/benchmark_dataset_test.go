package authx

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

type benchmarkDatasetCredential struct {
	UserID string
}

type benchmarkDatasetQuery struct {
	userID   string
	action   string
	resource string
	allowed  bool
}

type benchmarkDataset struct {
	userPermissions map[string]map[string]struct{}
	queries         []benchmarkDatasetQuery
}

func BenchmarkEngineCheckThenCan10kUsers10kPermissions(b *testing.B) {
	ctx := context.Background()
	dataset := newBenchmarkDataset(10_000, 10_000, 16, 4_096)

	for _, benchCase := range []struct {
		name     string
		withHook bool
	}{
		{name: "NoHook", withHook: false},
		{name: "WithHook", withHook: true},
	} {
		b.Run(benchCase.name, func(b *testing.B) {
			engine := newBenchmarkDatasetEngine(dataset, benchCase.withHook)
			queries := dataset.queries

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				query := queries[i%len(queries)]

				result, err := engine.Check(ctx, benchmarkDatasetCredential{UserID: query.userID})
				if err != nil {
					b.Fatalf("check failed: %v", err)
				}

				decision, err := engine.Can(ctx, AuthorizationModel{
					Principal: result.Principal,
					Action:    query.action,
					Resource:  query.resource,
				})
				if err != nil {
					b.Fatalf("can failed: %v", err)
				}
				if decision.Allowed != query.allowed {
					b.Fatalf("decision mismatch: allowed=%v expected=%v", decision.Allowed, query.allowed)
				}
			}
		})
	}
}

func BenchmarkEngineCheckThenCan10kUsers10kPermissionsParallel(b *testing.B) {
	ctx := context.Background()
	dataset := newBenchmarkDataset(10_000, 10_000, 16, 4_096)

	for _, benchCase := range []struct {
		name     string
		withHook bool
	}{
		{name: "NoHook", withHook: false},
		{name: "WithHook", withHook: true},
	} {
		b.Run(benchCase.name, func(b *testing.B) {
			engine := newBenchmarkDatasetEngine(dataset, benchCase.withHook)
			queries := dataset.queries

			b.ReportAllocs()
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				queryIndex := 0
				for pb.Next() {
					query := queries[queryIndex%len(queries)]
					queryIndex++

					result, err := engine.Check(ctx, benchmarkDatasetCredential{UserID: query.userID})
					if err != nil {
						b.Fatalf("check failed: %v", err)
					}

					decision, err := engine.Can(ctx, AuthorizationModel{
						Principal: result.Principal,
						Action:    query.action,
						Resource:  query.resource,
					})
					if err != nil {
						b.Fatalf("can failed: %v", err)
					}
					if decision.Allowed != query.allowed {
						b.Fatalf("decision mismatch: allowed=%v expected=%v", decision.Allowed, query.allowed)
					}
				}
			})
		})
	}
}

func newBenchmarkDatasetEngine(dataset benchmarkDataset, withHook bool) *Engine {
	manager := NewProviderManager(
		NewAuthenticationProviderFunc(func(
			_ context.Context,
			credential benchmarkDatasetCredential,
		) (AuthenticationResult, error) {
			if _, ok := dataset.userPermissions[credential.UserID]; !ok {
				return AuthenticationResult{}, ErrUnauthenticated
			}
			return AuthenticationResult{
				Principal: Principal{ID: credential.UserID},
			}, nil
		}),
	)

	authorizer := AuthorizerFunc(func(_ context.Context, input AuthorizationModel) (Decision, error) {
		principal, ok := input.Principal.(Principal)
		if !ok || principal.ID == "" {
			return Decision{Allowed: false, Reason: "invalid_principal"}, nil
		}

		userPermissions, ok := dataset.userPermissions[principal.ID]
		if !ok {
			return Decision{Allowed: false, Reason: "user_not_found"}, nil
		}

		_, allowed := userPermissions[permissionKey(input.Action, input.Resource)]
		if !allowed {
			return Decision{Allowed: false, Reason: "no_permission"}, nil
		}
		return Decision{Allowed: true}, nil
	})

	opts := []EngineOption{
		WithAuthenticationManager(manager),
		WithAuthorizer(authorizer),
	}
	if withHook {
		opts = append(opts, WithHook(noopHook{}))
	}
	return NewEngine(opts...)
}

func newBenchmarkDataset(
	userCount int,
	permissionCount int,
	permissionsPerUser int,
	queryCount int,
) benchmarkDataset {
	randSource := gofakeit.New(42)
	permissions := make([]string, permissionCount)
	for i := 0; i < permissionCount; i++ {
		action := fmt.Sprintf("%s-%03d", normalizeFakeToken(randSource.Verb()), i/100)
		resource := fmt.Sprintf("%s-%03d", normalizeFakeToken(randSource.Noun()), i%100)
		permissions[i] = permissionKey(action, resource)
	}

	userIDs := make([]string, userCount)
	userPermissions := make(map[string]map[string]struct{}, userCount)
	for i := 0; i < userCount; i++ {
		userID := fmt.Sprintf("%s-%05d", normalizeFakeToken(randSource.Username()), i)
		userIDs[i] = userID

		assigned := make(map[string]struct{}, permissionsPerUser)
		for len(assigned) < permissionsPerUser {
			assigned[permissions[randSource.Number(0, len(permissions)-1)]] = struct{}{}
		}
		userPermissions[userID] = assigned
	}

	queries := make([]benchmarkDatasetQuery, queryCount)
	for i := 0; i < queryCount; i++ {
		userID := userIDs[randSource.Number(0, len(userIDs)-1)]
		assigned := userPermissions[userID]

		permission := samplePermission(randSource, assigned)
		allowed := true
		if i%2 == 1 {
			allowed = false
			for {
				candidate := permissions[randSource.Number(0, len(permissions)-1)]
				if _, exists := assigned[candidate]; !exists {
					permission = candidate
					break
				}
			}
		}

		action, resource := parsePermissionKey(permission)
		queries[i] = benchmarkDatasetQuery{
			userID:   userID,
			action:   action,
			resource: resource,
			allowed:  allowed,
		}
	}

	return benchmarkDataset{
		userPermissions: userPermissions,
		queries:         queries,
	}
}

func samplePermission(randSource *gofakeit.Faker, assigned map[string]struct{}) string {
	target := randSource.Number(0, len(assigned)-1)
	for permission := range assigned {
		if target == 0 {
			return permission
		}
		target--
	}
	return ""
}

func permissionKey(action string, resource string) string {
	return action + "|" + resource
}

func parsePermissionKey(key string) (string, string) {
	action, resource, found := strings.Cut(key, "|")
	if !found {
		return key, ""
	}
	return action, resource
}

func normalizeFakeToken(raw string) string {
	token := strings.ToLower(strings.TrimSpace(raw))
	token = strings.ReplaceAll(token, " ", "_")
	token = strings.ReplaceAll(token, "-", "_")
	if token == "" {
		return "x"
	}
	return token
}
