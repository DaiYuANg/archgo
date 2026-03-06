package authx

import (
	"context"
	"strings"

	collectionmapping "github.com/DaiYuANg/arcgo/collectionx/mapping"
)

// InMemoryIdentityProvider is a runtime-mutable identity provider.
type InMemoryIdentityProvider struct {
	users *collectionmapping.ConcurrentMap[string, UserDetails]
}

// NewInMemoryIdentityProvider creates an empty in-memory identity provider.
func NewInMemoryIdentityProvider() *InMemoryIdentityProvider {
	return &InMemoryIdentityProvider{
		users: collectionmapping.NewConcurrentMap[string, UserDetails](),
	}
}

// UpsertUser adds or replaces user details by principal.
func (p *InMemoryIdentityProvider) UpsertUser(user UserDetails) error {
	if p == nil || p.users == nil {
		return ErrInvalidAuthenticator
	}

	normalized := user.normalize()
	if err := normalized.validate(); err != nil {
		return err
	}

	p.users.Set(normalized.Principal, normalized)
	return nil
}

// RemoveUser removes user details by principal.
func (p *InMemoryIdentityProvider) RemoveUser(principal string) bool {
	if p == nil || p.users == nil {
		return false
	}
	return p.users.Delete(strings.TrimSpace(principal))
}

// LoadByPrincipal loads user details by principal.
func (p *InMemoryIdentityProvider) LoadByPrincipal(ctx context.Context, principal string) (UserDetails, error) {
	_ = ctx
	if p == nil || p.users == nil {
		return UserDetails{}, ErrInvalidAuthenticator
	}

	user, exists := p.users.Get(strings.TrimSpace(principal))
	if !exists {
		return UserDetails{}, ErrUnauthorized
	}

	return user, nil
}
