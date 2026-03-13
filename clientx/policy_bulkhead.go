package clientx

import "context"

type concurrencyLimitPolicy struct {
	sem chan struct{}
}

func NewConcurrencyLimitPolicy(maxInFlight int) Policy {
	if maxInFlight <= 0 {
		maxInFlight = 1
	}
	return &concurrencyLimitPolicy{sem: make(chan struct{}, maxInFlight)}
}

func (p *concurrencyLimitPolicy) Before(ctx context.Context, operation Operation) (context.Context, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case p.sem <- struct{}{}:
		return ctx, nil
	case <-ctx.Done():
		return ctx, ctx.Err()
	}
}

func (p *concurrencyLimitPolicy) After(ctx context.Context, operation Operation, err error) error {
	select {
	case <-p.sem:
	default:
	}
	return nil
}
