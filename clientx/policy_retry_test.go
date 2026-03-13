package clientx

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryPolicyRetriesUntilSuccess(t *testing.T) {
	attempts := 0
	policy := NewRetryPolicy(RetryPolicyConfig{
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
				return "", WrapError(ProtocolHTTP, "get", "example", context.DeadlineExceeded)
			}
			return "ok", nil
		},
		policy,
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

func TestRetryPolicyStopsAtMaxAttempts(t *testing.T) {
	attempts := 0
	policy := NewRetryPolicy(RetryPolicyConfig{
		MaxAttempts: 3,
		BaseDelay:   time.Millisecond,
		MaxDelay:    time.Millisecond,
		JitterRatio: 0,
	})

	_, err := InvokeWithPolicies(
		context.Background(),
		Operation{Protocol: ProtocolTCP, Kind: OperationKindDial, Op: "dial"},
		func(ctx context.Context) (int, error) {
			attempts++
			return 0, WrapError(ProtocolTCP, "dial", "127.0.0.1:1", context.DeadlineExceeded)
		},
		policy,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryPolicySkipsNonRetryableError(t *testing.T) {
	attempts := 0
	policy := NewRetryPolicy(RetryPolicyConfig{MaxAttempts: 5})

	_, err := InvokeWithPolicies(
		context.Background(),
		Operation{Protocol: ProtocolUDP, Kind: OperationKindDial, Op: "dial"},
		func(ctx context.Context) (int, error) {
			attempts++
			return 0, WrapErrorWithKind(ProtocolUDP, "dial", "127.0.0.1:1", ErrorKindCodec, errors.New("codec"))
		},
		policy,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}

func TestRetryPolicyContextCancelDuringBackoff(t *testing.T) {
	attempts := 0
	policy := NewRetryPolicy(RetryPolicyConfig{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    100 * time.Millisecond,
		JitterRatio: 0,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_, err := InvokeWithPolicies(
		ctx,
		Operation{Protocol: ProtocolHTTP, Kind: OperationKindRequest, Op: "get"},
		func(ctx context.Context) (int, error) {
			attempts++
			return 0, WrapError(ProtocolHTTP, "get", "example", context.DeadlineExceeded)
		},
		policy,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled error, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}
