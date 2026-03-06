package authx

import "context"

type identityContextKey struct{}

// WithIdentity stores identity into context.
func WithIdentity(ctx context.Context, identity Identity) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if identity == nil {
		return ctx
	}
	authentication := NewAuthentication(identity, 0)
	return WithSecurityContext(ctx, NewSecurityContext(authentication))
}

// CurrentIdentity extracts identity from context.
func CurrentIdentity(ctx context.Context) (Identity, bool) {
	securityContext, ok := CurrentSecurityContext(ctx)
	if ok {
		identity := securityContext.Identity()
		return identity, identity != nil
	}
	return nil, false
}

// RequireIdentity extracts identity or returns ErrNoIdentity.
func RequireIdentity(ctx context.Context) (Identity, error) {
	identity, ok := CurrentIdentity(ctx)
	if !ok {
		return nil, ErrNoIdentity
	}
	return identity, nil
}
