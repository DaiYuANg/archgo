package http

import (
	"time"

	"github.com/DaiYuANg/arcgo/clientx"
	"resty.dev/v3"
)

type Option func(*DefaultClient)

func WithRequestMiddleware(fn func(*resty.Client, *resty.Request) error) Option {
	return func(c *DefaultClient) {
		c.Raw().AddRequestMiddleware(fn)
	}
}

func WithResponseMiddleware(fn func(*resty.Client, *resty.Response) error) Option {
	return func(c *DefaultClient) {
		c.Raw().AddResponseMiddleware(fn)
	}
}

func WithHeader(key, value string) Option {
	return func(c *DefaultClient) {
		c.Raw().SetHeader(key, value)
	}
}

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
