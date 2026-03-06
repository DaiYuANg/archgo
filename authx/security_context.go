package authx

import (
	"context"
	"time"
)

type securityContextKey struct{}

// Authentication is the authenticated principal object returned by manager.
type Authentication struct {
	identity        Identity
	policyVersion   int64
	authenticatedAt time.Time
}

// NewAuthentication builds a new authentication snapshot.
func NewAuthentication(identity Identity, policyVersion int64) Authentication {
	return Authentication{
		identity:        identity,
		policyVersion:   policyVersion,
		authenticatedAt: time.Now(),
	}
}

// Identity returns authenticated identity.
func (a Authentication) Identity() Identity {
	return a.identity
}

// PolicyVersion returns policy version used when authentication object is created.
func (a Authentication) PolicyVersion() int64 {
	return a.policyVersion
}

// AuthenticatedAt returns authentication creation time.
func (a Authentication) AuthenticatedAt() time.Time {
	return a.authenticatedAt
}

// IsAuthenticated reports whether the authentication has an authenticated identity.
func (a Authentication) IsAuthenticated() bool {
	return a.identity != nil && a.identity.IsAuthenticated()
}

// SecurityContext stores security state for current request lifecycle.
type SecurityContext struct {
	authentication Authentication
}

// NewSecurityContext creates a security context with authentication object.
func NewSecurityContext(authentication Authentication) SecurityContext {
	return SecurityContext{authentication: authentication}
}

// Authentication returns authentication object.
func (s SecurityContext) Authentication() Authentication {
	return s.authentication
}

// Identity returns identity from authentication object.
func (s SecurityContext) Identity() Identity {
	return s.authentication.Identity()
}

// WithSecurityContext stores security context into runtime context.
func WithSecurityContext(ctx context.Context, securityContext SecurityContext) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	identity := securityContext.Identity()
	if identity == nil {
		return ctx
	}

	ctx = context.WithValue(ctx, securityContextKey{}, securityContext)
	return context.WithValue(ctx, identityContextKey{}, identity)
}

// CurrentSecurityContext extracts security context.
func CurrentSecurityContext(ctx context.Context) (SecurityContext, bool) {
	if ctx == nil {
		return SecurityContext{}, false
	}

	securityContext, ok := ctx.Value(securityContextKey{}).(SecurityContext)
	if ok && securityContext.Identity() != nil {
		return securityContext, true
	}

	identity, identityExists := ctx.Value(identityContextKey{}).(Identity)
	if identityExists && identity != nil {
		return NewSecurityContext(NewAuthentication(identity, 0)), true
	}

	return SecurityContext{}, false
}

// RequireSecurityContext extracts security context or returns ErrNoIdentity.
func RequireSecurityContext(ctx context.Context) (SecurityContext, error) {
	securityContext, ok := CurrentSecurityContext(ctx)
	if !ok {
		return SecurityContext{}, ErrNoIdentity
	}
	return securityContext, nil
}

// CurrentAuthentication extracts current authentication object.
func CurrentAuthentication(ctx context.Context) (Authentication, bool) {
	securityContext, ok := CurrentSecurityContext(ctx)
	if !ok {
		return Authentication{}, false
	}
	authentication := securityContext.Authentication()
	if !authentication.IsAuthenticated() {
		return Authentication{}, false
	}
	return authentication, true
}

// RequireAuthentication extracts current authentication object or returns ErrNoIdentity.
func RequireAuthentication(ctx context.Context) (Authentication, error) {
	authentication, ok := CurrentAuthentication(ctx)
	if !ok {
		return Authentication{}, ErrNoIdentity
	}
	return authentication, nil
}
