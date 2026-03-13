package clientx

import (
	"context"
	"time"
)

type timeoutPolicy struct {
	timeout time.Duration
}

type timeoutCancelKey struct{}

func NewTimeoutPolicy(timeout time.Duration) Policy {
	return &timeoutPolicy{timeout: timeout}
}

func (p *timeoutPolicy) Before(ctx context.Context, operation Operation) (context.Context, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if p.timeout <= 0 {
		return ctx, nil
	}

	if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) <= p.timeout {
		return ctx, nil
	}

	nextCtx, cancel := context.WithTimeout(ctx, p.timeout)
	return context.WithValue(nextCtx, timeoutCancelKey{}, cancel), nil
}

func (p *timeoutPolicy) After(ctx context.Context, operation Operation, err error) error {
	if ctx == nil {
		return nil
	}
	cancel, ok := ctx.Value(timeoutCancelKey{}).(context.CancelFunc)
	if ok && cancel != nil {
		cancel()
	}
	return nil
}
