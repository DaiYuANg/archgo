package authx

import (
	"context"
	"testing"
)

type benchmarkCredential struct {
	Token string
}

type noopHook struct{}

func (noopHook) BeforeCheck(context.Context, any) error {
	return nil
}

func (noopHook) AfterCheck(context.Context, any, AuthenticationResult, error) {}

func (noopHook) BeforeCan(context.Context, AuthorizationModel) error {
	return nil
}

func (noopHook) AfterCan(context.Context, AuthorizationModel, Decision, error) {}

func newBenchmarkEngine(withHook bool) *Engine {
	manager := NewProviderManager(
		NewAuthenticationProviderFunc(func(_ context.Context, credential benchmarkCredential) (AuthenticationResult, error) {
			return AuthenticationResult{
				Principal: Principal{
					ID: credential.Token,
				},
			}, nil
		}),
	)
	authorizer := AuthorizerFunc(func(_ context.Context, input AuthorizationModel) (Decision, error) {
		_ = input
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

func BenchmarkEngineCheck(b *testing.B) {
	ctx := context.Background()
	credential := benchmarkCredential{Token: "u-1"}

	for _, benchCase := range []struct {
		name     string
		withHook bool
	}{
		{name: "NoHook", withHook: false},
		{name: "WithHook", withHook: true},
	} {
		b.Run(benchCase.name, func(b *testing.B) {
			engine := newBenchmarkEngine(benchCase.withHook)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				result, err := engine.Check(ctx, credential)
				if err != nil {
					b.Fatalf("check failed: %v", err)
				}
				if result.Principal == nil {
					b.Fatal("principal should not be nil")
				}
			}
		})
	}
}

func BenchmarkEngineCan(b *testing.B) {
	ctx := context.Background()
	model := AuthorizationModel{
		Principal: Principal{ID: "u-1"},
		Action:    "query",
		Resource:  "order",
		Context: map[string]any{
			"order_id": "1",
		},
	}

	for _, benchCase := range []struct {
		name     string
		withHook bool
	}{
		{name: "NoHook", withHook: false},
		{name: "WithHook", withHook: true},
	} {
		b.Run(benchCase.name, func(b *testing.B) {
			engine := newBenchmarkEngine(benchCase.withHook)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				decision, err := engine.Can(ctx, model)
				if err != nil {
					b.Fatalf("can failed: %v", err)
				}
				if !decision.Allowed {
					b.Fatal("decision should be allowed")
				}
			}
		})
	}
}

func BenchmarkEngineCheckThenCan(b *testing.B) {
	ctx := context.Background()
	credential := benchmarkCredential{Token: "u-1"}

	for _, benchCase := range []struct {
		name     string
		withHook bool
	}{
		{name: "NoHook", withHook: false},
		{name: "WithHook", withHook: true},
	} {
		b.Run(benchCase.name, func(b *testing.B) {
			engine := newBenchmarkEngine(benchCase.withHook)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				result, err := engine.Check(ctx, credential)
				if err != nil {
					b.Fatalf("check failed: %v", err)
				}

				decision, err := engine.Can(ctx, AuthorizationModel{
					Principal: result.Principal,
					Action:    "query",
					Resource:  "order",
					Context: map[string]any{
						"order_id": "1",
					},
				})
				if err != nil {
					b.Fatalf("can failed: %v", err)
				}
				if !decision.Allowed {
					b.Fatal("decision should be allowed")
				}
			}
		})
	}
}
