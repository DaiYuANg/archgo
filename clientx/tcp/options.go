package tcp

import (
	"time"

	"github.com/DaiYuANg/arcgo/clientx"
)

type Option func(*DefaultClient)

func WithHooks(hooks ...clientx.Hook) Option {
	return func(c *DefaultClient) {
		c.hooks = append(c.hooks, hooks...)
	}
}

func WithPolicies(policies ...clientx.Policy) Option {
	return func(c *DefaultClient) {
		c.policies = append(c.policies, policies...)
	}
}

func WithConcurrencyLimit(maxInFlight int) Option {
	return func(c *DefaultClient) {
		c.policies = append(c.policies, clientx.NewConcurrencyLimitPolicy(maxInFlight))
	}
}

func WithTimeoutGuard(timeout time.Duration) Option {
	return func(c *DefaultClient) {
		c.policies = append(c.policies, clientx.NewTimeoutPolicy(timeout))
	}
}
