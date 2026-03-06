package authx

import (
	"maps"
	"slices"
	"strings"

	"github.com/samber/lo"
)

// Identity describes an authenticated runtime principal.
type Identity interface {
	ID() string
	Type() string
	Name() string
	Principal() any
	Roles() []string
	Permissions() []string
	Attributes() map[string]string
	IsAuthenticated() bool
}

// IdentityOption configures BasicIdentity.
type IdentityOption func(identity *BasicIdentity)

// BasicIdentity is the default immutable identity implementation.
type BasicIdentity struct {
	id            string
	subjectType   string
	name          string
	principal     any
	roles         []string
	permissions   []string
	attributes    map[string]string
	authenticated bool
}

// NewIdentity creates a new identity snapshot.
func NewIdentity(id, subjectType, name string, opts ...IdentityOption) *BasicIdentity {
	normalizedID := strings.TrimSpace(id)

	identity := &BasicIdentity{
		id:            normalizedID,
		subjectType:   strings.TrimSpace(subjectType),
		name:          strings.TrimSpace(name),
		principal:     nil,
		roles:         make([]string, 0),
		permissions:   make([]string, 0),
		attributes:    make(map[string]string),
		authenticated: normalizedID != "",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(identity)
		}
	}

	return identity
}

// AnonymousIdentity creates an unauthenticated anonymous identity.
func AnonymousIdentity() *BasicIdentity {
	return NewIdentity(
		"",
		"anonymous",
		"anonymous",
		WithAuthenticated(false),
	)
}

// WithRoles appends identity roles.
func WithRoles(roles ...string) IdentityOption {
	return func(identity *BasicIdentity) {
		if identity == nil {
			return
		}
		identity.roles = uniqueStrings(append(identity.roles, roles...))
	}
}

// WithPrincipal sets business principal payload.
func WithPrincipal(principal any) IdentityOption {
	return func(identity *BasicIdentity) {
		if identity == nil {
			return
		}
		identity.principal = principal
	}
}

// WithPermissions appends identity permissions.
func WithPermissions(permissions ...string) IdentityOption {
	return func(identity *BasicIdentity) {
		if identity == nil {
			return
		}
		identity.permissions = uniqueStrings(append(identity.permissions, permissions...))
	}
}

// WithAttributes sets identity attributes.
func WithAttributes(attributes map[string]string) IdentityOption {
	return func(identity *BasicIdentity) {
		if identity == nil {
			return
		}

		identity.attributes = make(map[string]string, len(attributes))
		for k, v := range attributes {
			trimmedKey := strings.TrimSpace(k)
			if trimmedKey == "" {
				continue
			}
			identity.attributes[trimmedKey] = strings.TrimSpace(v)
		}
	}
}

// WithAuthenticated explicitly sets authenticated state.
func WithAuthenticated(authenticated bool) IdentityOption {
	return func(identity *BasicIdentity) {
		if identity == nil {
			return
		}
		identity.authenticated = authenticated
	}
}

// ID returns identity id.
func (i *BasicIdentity) ID() string {
	return i.id
}

// Type returns subject type.
func (i *BasicIdentity) Type() string {
	return i.subjectType
}

// Name returns display name.
func (i *BasicIdentity) Name() string {
	return i.name
}

// Principal returns business principal payload.
func (i *BasicIdentity) Principal() any {
	return i.principal
}

// Roles returns a copy of roles.
func (i *BasicIdentity) Roles() []string {
	return slices.Clone(i.roles)
}

// Permissions returns a copy of permissions.
func (i *BasicIdentity) Permissions() []string {
	return slices.Clone(i.permissions)
}

// Attributes returns a copy of attributes.
func (i *BasicIdentity) Attributes() map[string]string {
	return maps.Clone(i.attributes)
}

// IsAuthenticated indicates whether identity is authenticated.
func (i *BasicIdentity) IsAuthenticated() bool {
	return i.authenticated
}

func uniqueStrings(values []string) []string {
	return lo.Uniq(lo.FilterMap(values, func(value string, _ int) (string, bool) {
		trimmed := strings.TrimSpace(value)
		return trimmed, trimmed != ""
	}))
}
