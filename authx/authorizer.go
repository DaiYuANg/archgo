package authx

import (
	"context"
	"fmt"
	"maps"
	"strings"
)

// Request is an authorization request.
type Request struct {
	Action     string
	Resource   string
	Attributes map[string]string
}

// NewRequest creates an authorization request with copied attributes.
func NewRequest(action, resource string, attributes map[string]string) Request {
	return Request{
		Action:     strings.TrimSpace(action),
		Resource:   strings.TrimSpace(resource),
		Attributes: maps.Clone(attributes),
	}
}

// Validate validates authorization request.
func (r Request) Validate() error {
	if strings.TrimSpace(r.Action) == "" {
		return fmt.Errorf("%w: action is required", ErrInvalidRequest)
	}
	if strings.TrimSpace(r.Resource) == "" {
		return fmt.Errorf("%w: resource is required", ErrInvalidRequest)
	}
	return nil
}

// Decision is the authorization result.
type Decision struct {
	Allowed bool
	Reason  string
}

// Allow returns an allow decision.
func Allow(reason string) Decision {
	return Decision{
		Allowed: true,
		Reason:  strings.TrimSpace(reason),
	}
}

// Deny returns a deny decision.
func Deny(reason string) Decision {
	return Decision{
		Allowed: false,
		Reason:  strings.TrimSpace(reason),
	}
}

// Authorizer decides whether identity can access a target action/resource.
type Authorizer interface {
	Authorize(ctx context.Context, identity Identity, request Request) (Decision, error)
}

// AuthorizerFunc is a function-based authorizer.
type AuthorizerFunc func(ctx context.Context, identity Identity, request Request) (Decision, error)

// Authorize executes authorizer function.
func (f AuthorizerFunc) Authorize(ctx context.Context, identity Identity, request Request) (Decision, error) {
	if f == nil {
		return Decision{}, fmt.Errorf("%w: authorizer function is nil", ErrInvalidAuthorizer)
	}
	return f(ctx, identity, request)
}

// Require executes authorization and returns typed auth errors.
func Require(authorizer Authorizer, ctx context.Context, identity Identity, request Request) error {
	if authorizer == nil {
		return fmt.Errorf("%w: authorizer is nil", ErrInvalidAuthorizer)
	}
	if identity == nil || !identity.IsAuthenticated() {
		return ErrUnauthorized
	}
	if err := request.Validate(); err != nil {
		return err
	}

	decision, err := authorizer.Authorize(ctx, identity, request)
	if err != nil {
		return err
	}
	if !decision.Allowed {
		return ErrForbidden
	}
	return nil
}
