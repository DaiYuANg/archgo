package authx

import (
	"context"
	"fmt"
)

// PrincipalAs casts identity principal payload to target type.
func PrincipalAs[T any](identity Identity) (T, bool) {
	var zero T
	if identity == nil {
		return zero, false
	}

	value, ok := identity.Principal().(T)
	if !ok {
		return zero, false
	}
	return value, true
}

// AuthenticationPrincipalAs casts authentication principal payload to target type.
func AuthenticationPrincipalAs[T any](authentication Authentication) (T, bool) {
	return PrincipalAs[T](authentication.Identity())
}

// CurrentPrincipalAs extracts and casts principal payload from context.
func CurrentPrincipalAs[T any](ctx context.Context) (T, bool) {
	identity, ok := CurrentIdentity(ctx)
	if !ok {
		var zero T
		return zero, false
	}
	return PrincipalAs[T](identity)
}

// RequirePrincipalAs extracts and casts principal payload from context or returns error.
func RequirePrincipalAs[T any](ctx context.Context) (T, error) {
	principal, ok := CurrentPrincipalAs[T](ctx)
	if ok {
		return principal, nil
	}
	var zero T
	return zero, fmt.Errorf("%w: principal type mismatch or unavailable", ErrNoIdentity)
}
