package clientx

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestConcurrencyLimitPolicySerialize(t *testing.T) {
	policy := NewConcurrencyLimitPolicy(1)
	var active int32
	var maxActive int32

	fn := func(ctx context.Context) (int, error) {
		current := atomic.AddInt32(&active, 1)
		defer atomic.AddInt32(&active, -1)
		for {
			seen := atomic.LoadInt32(&maxActive)
			if current <= seen || atomic.CompareAndSwapInt32(&maxActive, seen, current) {
				break
			}
		}
		time.Sleep(30 * time.Millisecond)
		return 1, nil
	}

	var wg sync.WaitGroup
	wg.Add(2)
	for range 2 {
		go func() {
			defer wg.Done()
			_, err := InvokeWithPolicies(
				context.Background(),
				Operation{Protocol: ProtocolHTTP, Kind: OperationKindRequest, Op: "get"},
				fn,
				policy,
			)
			if err != nil {
				t.Errorf("invoke failed: %v", err)
			}
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&maxActive); got != 1 {
		t.Fatalf("expected max concurrency 1, got %d", got)
	}
}

func TestConcurrencyLimitPolicyRespectsContextCancel(t *testing.T) {
	policy := NewConcurrencyLimitPolicy(1)
	op := Operation{Protocol: ProtocolTCP, Kind: OperationKindDial, Op: "dial"}

	ctx := context.Background()
	_, err := policy.Before(ctx, op)
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err = policy.Before(timeoutCtx, op)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}

	if err := policy.After(context.Background(), op, nil); err != nil {
		t.Fatalf("release failed: %v", err)
	}
}

func TestConcurrencyLimitPolicyReleaseAfterError(t *testing.T) {
	policy := NewConcurrencyLimitPolicy(1)
	boom := errors.New("boom")

	for i := range 2 {
		_, err := InvokeWithPolicies(
			context.Background(),
			Operation{Protocol: ProtocolUDP, Kind: OperationKindDial, Op: "dial"},
			func(ctx context.Context) (int, error) {
				return 0, boom
			},
			policy,
		)
		if !errors.Is(err, boom) {
			t.Fatalf("round %d expected boom error, got %v", i, err)
		}
	}
}
