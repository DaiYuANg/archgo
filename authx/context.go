package authx

import "context"

type principalContextKey struct{}

func WithPrincipal(ctx context.Context, principal any) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (any, bool) {
	principal := ctx.Value(principalContextKey{})
	if principal == nil {
		return nil, false
	}
	return principal, true
}

func PrincipalFromContextAs[T any](ctx context.Context) (T, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(T)
	return principal, ok
}
