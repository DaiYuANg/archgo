package clientx

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type panicRetryPolicy struct{}

func (p panicRetryPolicy) Before(ctx context.Context, operation Operation) (context.Context, error) {
	return ctx, nil
}

func (p panicRetryPolicy) After(ctx context.Context, operation Operation, err error) error {
	return nil
}

func (p panicRetryPolicy) ShouldRetry(ctx context.Context, operation Operation, attempt int, err error) (bool, time.Duration) {
	panic("retry decider panic")
}

func TestInvokeWithPoliciesOrder(t *testing.T) {
	calls := make([]string, 0, 8)
	p1 := PolicyFuncs{
		BeforeFunc: func(ctx context.Context, operation Operation) (context.Context, error) {
			calls = append(calls, "before1")
			return context.WithValue(ctx, "v", "ok"), nil
		},
		AfterFunc: func(ctx context.Context, operation Operation, err error) error {
			calls = append(calls, "after1")
			return nil
		},
	}
	p2 := PolicyFuncs{
		BeforeFunc: func(ctx context.Context, operation Operation) (context.Context, error) {
			calls = append(calls, "before2")
			if got, _ := ctx.Value("v").(string); got != "ok" {
				t.Fatalf("expected ctx value propagated, got %q", got)
			}
			return ctx, nil
		},
		AfterFunc: func(ctx context.Context, operation Operation, err error) error {
			calls = append(calls, "after2")
			return nil
		},
	}

	out, err := InvokeWithPolicies(
		context.Background(),
		Operation{Protocol: ProtocolHTTP, Kind: OperationKindRequest, Op: "get"},
		func(ctx context.Context) (string, error) {
			calls = append(calls, "execute")
			return "done", nil
		},
		p1,
		p2,
	)
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	if out != "done" {
		t.Fatalf("unexpected output: %q", out)
	}

	expected := []string{"before1", "before2", "execute", "after2", "after1"}
	if !reflect.DeepEqual(calls, expected) {
		t.Fatalf("unexpected call order: got=%v want=%v", calls, expected)
	}
}

func TestInvokeWithPoliciesBeforeError(t *testing.T) {
	boom := errors.New("boom")
	called := false

	_, err := InvokeWithPolicies(
		context.Background(),
		Operation{Protocol: ProtocolTCP, Kind: OperationKindDial, Op: "dial"},
		func(ctx context.Context) (int, error) {
			called = true
			return 1, nil
		},
		PolicyFuncs{BeforeFunc: func(ctx context.Context, operation Operation) (context.Context, error) {
			return ctx, boom
		}},
	)
	if !errors.Is(err, boom) {
		t.Fatalf("expected before error, got %v", err)
	}
	if called {
		t.Fatal("expected invoke function not to be called")
	}
}

func TestInvokeWithPoliciesAfterErrorJoin(t *testing.T) {
	baseErr := errors.New("base")
	afterErr := errors.New("after")

	_, err := InvokeWithPolicies(
		context.Background(),
		Operation{Protocol: ProtocolUDP, Kind: OperationKindListen, Op: "listen"},
		func(ctx context.Context) (int, error) {
			return 0, baseErr
		},
		PolicyFuncs{AfterFunc: func(ctx context.Context, operation Operation, err error) error {
			return afterErr
		}},
	)
	if !errors.Is(err, baseErr) || !errors.Is(err, afterErr) {
		t.Fatalf("expected joined errors, got %v", err)
	}
}

func TestInvokeWithPoliciesBeforePanicIsolation(t *testing.T) {
	called := false

	out, err := InvokeWithPolicies(
		context.Background(),
		Operation{Protocol: ProtocolHTTP, Kind: OperationKindRequest, Op: "get"},
		func(ctx context.Context) (string, error) {
			called = true
			return "ok", nil
		},
		PolicyFuncs{
			BeforeFunc: func(ctx context.Context, operation Operation) (context.Context, error) {
				panic("before panic")
			},
		},
	)
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	if out != "ok" {
		t.Fatalf("unexpected output: %q", out)
	}
	if !called {
		t.Fatal("expected invoke function to be called")
	}
}

func TestInvokeWithPoliciesAfterPanicIsolation(t *testing.T) {
	_, err := InvokeWithPolicies(
		context.Background(),
		Operation{Protocol: ProtocolTCP, Kind: OperationKindDial, Op: "dial"},
		func(ctx context.Context) (int, error) {
			return 1, nil
		},
		PolicyFuncs{
			AfterFunc: func(ctx context.Context, operation Operation, err error) error {
				panic("after panic")
			},
		},
	)
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
}

func TestInvokeWithPoliciesRetryDeciderPanicIsolation(t *testing.T) {
	boom := errors.New("boom")
	attempts := 0

	_, err := InvokeWithPolicies(
		context.Background(),
		Operation{Protocol: ProtocolUDP, Kind: OperationKindDial, Op: "dial"},
		func(ctx context.Context) (int, error) {
			attempts++
			return 0, boom
		},
		panicRetryPolicy{},
	)
	if !errors.Is(err, boom) {
		t.Fatalf("expected boom error, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}
