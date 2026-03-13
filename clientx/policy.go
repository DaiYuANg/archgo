package clientx

import (
	"context"
	"errors"
	"time"

	"github.com/samber/lo"
)

type OperationKind string

const (
	OperationKindUnknown OperationKind = "unknown"
	OperationKindRequest OperationKind = "request"
	OperationKindDial    OperationKind = "dial"
	OperationKindListen  OperationKind = "listen"
)

type Operation struct {
	Protocol Protocol
	Kind     OperationKind
	Op       string
	Network  string
	Addr     string
}

type Policy interface {
	Before(ctx context.Context, operation Operation) (context.Context, error)
	After(ctx context.Context, operation Operation, err error) error
}

// RetryDecider allows a policy to request re-execution with an optional delay.
type RetryDecider interface {
	ShouldRetry(ctx context.Context, operation Operation, attempt int, err error) (retry bool, wait time.Duration)
}

type PolicyFuncs struct {
	BeforeFunc func(ctx context.Context, operation Operation) (context.Context, error)
	AfterFunc  func(ctx context.Context, operation Operation, err error) error
}

func (p PolicyFuncs) Before(ctx context.Context, operation Operation) (context.Context, error) {
	if p.BeforeFunc != nil {
		return p.BeforeFunc(ctx, operation)
	}
	return ctx, nil
}

func (p PolicyFuncs) After(ctx context.Context, operation Operation, err error) error {
	if p.AfterFunc != nil {
		return p.AfterFunc(ctx, operation, err)
	}
	return nil
}

func InvokeWithPolicies[T any](
	ctx context.Context,
	operation Operation,
	fn func(context.Context) (T, error),
	policies ...Policy,
) (T, error) {
	var zero T
	if ctx == nil {
		ctx = context.Background()
	}
	if fn == nil {
		return zero, errors.New("invoke function is nil")
	}
	if operation.Protocol == "" {
		operation.Protocol = ProtocolUnknown
	}
	if operation.Kind == "" {
		operation.Kind = OperationKindUnknown
	}

	activePolicies := lo.Filter(policies, func(policy Policy, _ int) bool {
		return policy != nil
	})

	var result T
	for attempt := 1; ; attempt++ {
		attemptCtx := ctx
		applied := make([]Policy, 0, len(activePolicies))
		for _, policy := range activePolicies {
			nextCtx, err := callPolicyBefore(policy, attemptCtx, operation)
			if err != nil {
				return zero, applyAfterPolicies(applied, attemptCtx, operation, err)
			}
			attemptCtx = nextCtx
			applied = append(applied, policy)
		}

		execErr := error(nil)
		result, execErr = fn(attemptCtx)
		execErr = applyAfterPolicies(applied, attemptCtx, operation, execErr)
		if execErr == nil {
			return result, nil
		}

		retry, wait := decideRetry(activePolicies, ctx, operation, attempt, execErr)
		if !retry {
			return result, execErr
		}
		if sleepErr := sleepWithContext(ctx, wait); sleepErr != nil {
			return result, errors.Join(execErr, sleepErr)
		}
	}
}

func applyAfterPolicies(policies []Policy, ctx context.Context, operation Operation, baseErr error) error {
	err := baseErr
	for i := len(policies) - 1; i >= 0; i-- {
		policy := policies[i]
		afterErr, afterOK := callPolicyAfter(policy, ctx, operation, err)
		if !afterOK {
			continue
		}
		if afterErr != nil {
			err = errors.Join(err, afterErr)
		}
	}
	return err
}

func decideRetry(
	policies []Policy,
	ctx context.Context,
	operation Operation,
	attempt int,
	err error,
) (retry bool, wait time.Duration) {
	for _, policy := range policies {
		decider, ok := policy.(RetryDecider)
		if !ok {
			continue
		}
		shouldRetry, delay, retryOK := callShouldRetry(decider, ctx, operation, attempt, err)
		if !retryOK {
			continue
		}
		if !shouldRetry {
			continue
		}
		retry = true
		if delay > wait {
			wait = delay
		}
	}
	if wait < 0 {
		wait = 0
	}
	return retry, wait
}

func callPolicyBefore(
	policy Policy,
	ctx context.Context,
	operation Operation,
) (nextCtx context.Context, err error) {
	nextCtx = ctx
	defer func() {
		if recover() != nil {
			nextCtx = ctx
			err = nil
		}
	}()

	policyCtx, policyErr := policy.Before(ctx, operation)
	if policyCtx != nil {
		nextCtx = policyCtx
	}
	return nextCtx, policyErr
}

func callPolicyAfter(
	policy Policy,
	ctx context.Context,
	operation Operation,
	err error,
) (afterErr error, ok bool) {
	ok = true
	defer func() {
		if recover() != nil {
			afterErr = nil
			ok = false
		}
	}()
	return policy.After(ctx, operation, err), ok
}

func callShouldRetry(
	decider RetryDecider,
	ctx context.Context,
	operation Operation,
	attempt int,
	err error,
) (retry bool, wait time.Duration, ok bool) {
	ok = true
	defer func() {
		if recover() != nil {
			retry = false
			wait = 0
			ok = false
		}
	}()
	retry, wait = decider.ShouldRetry(ctx, operation, attempt, err)
	return retry, wait, ok
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
