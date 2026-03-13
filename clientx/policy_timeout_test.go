package clientx

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTimeoutPolicyTriggersDeadline(t *testing.T) {
	policy := NewTimeoutPolicy(30 * time.Millisecond)
	start := time.Now()

	_, err := InvokeWithPolicies(
		context.Background(),
		Operation{Protocol: ProtocolHTTP, Kind: OperationKindRequest, Op: "get"},
		func(ctx context.Context) (int, error) {
			<-ctx.Done()
			return 0, ctx.Err()
		},
		policy,
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}

	elapsed := time.Since(start)
	if elapsed < 20*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Fatalf("unexpected elapsed time: %v", elapsed)
	}
}

func TestTimeoutPolicyKeepsShorterParentDeadline(t *testing.T) {
	policy := NewTimeoutPolicy(200 * time.Millisecond)
	parentCtx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := InvokeWithPolicies(
		parentCtx,
		Operation{Protocol: ProtocolTCP, Kind: OperationKindDial, Op: "dial"},
		func(ctx context.Context) (int, error) {
			<-ctx.Done()
			return 0, ctx.Err()
		},
		policy,
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}

	if elapsed := time.Since(start); elapsed > 120*time.Millisecond {
		t.Fatalf("expected shorter parent deadline to apply, got elapsed %v", elapsed)
	}
}

func TestTimeoutPolicyTightensLongerParentDeadline(t *testing.T) {
	policy := NewTimeoutPolicy(30 * time.Millisecond)
	parentCtx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := InvokeWithPolicies(
		parentCtx,
		Operation{Protocol: ProtocolUDP, Kind: OperationKindDial, Op: "dial"},
		func(ctx context.Context) (int, error) {
			<-ctx.Done()
			return 0, ctx.Err()
		},
		policy,
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}

	if elapsed := time.Since(start); elapsed > 150*time.Millisecond {
		t.Fatalf("expected timeout guard to tighten deadline, got elapsed %v", elapsed)
	}
}

func TestTimeoutPolicyNoopWhenDisabled(t *testing.T) {
	policy := NewTimeoutPolicy(0)

	_, err := InvokeWithPolicies(
		context.Background(),
		Operation{Protocol: ProtocolHTTP, Kind: OperationKindRequest, Op: "get"},
		func(ctx context.Context) (int, error) {
			if _, ok := ctx.Deadline(); ok {
				t.Fatal("expected no deadline from timeout policy")
			}
			return 1, nil
		},
		policy,
	)
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
}

func TestTimeoutPolicyWorksWithRetryPolicy(t *testing.T) {
	attempts := 0
	timeoutPolicy := NewTimeoutPolicy(50 * time.Millisecond)
	retryPolicy := NewRetryPolicy(RetryPolicyConfig{
		MaxAttempts: 3,
		BaseDelay:   time.Millisecond,
		MaxDelay:    time.Millisecond,
		JitterRatio: 0,
	})

	result, err := InvokeWithPolicies(
		context.Background(),
		Operation{Protocol: ProtocolHTTP, Kind: OperationKindRequest, Op: "get"},
		func(ctx context.Context) (string, error) {
			attempts++
			if attempts < 3 {
				return "", context.DeadlineExceeded
			}
			return "ok", nil
		},
		timeoutPolicy,
		retryPolicy,
	)
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	if result != "ok" {
		t.Fatalf("unexpected result: %q", result)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}
