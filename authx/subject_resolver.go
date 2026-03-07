package authx

import (
	"context"
	"fmt"
	"strings"
)

// SubjectResolver resolves an authenticated Identity to an authorization subject string.
// This allows flexible mapping between identity and the subject used in policy evaluation.
//
// Use cases:
//   - Multi-tenant subject namespacing: tenant/{tenantID}/user/{userID}
//   - Service account mapping: external API key → internal service subject
//   - Custom subject naming strategies
//   - Domain-aware authorization contexts
type SubjectResolver interface {
	ResolveSubject(ctx context.Context, identity Identity) (string, error)
}

// SubjectResolverFunc is a function-based subject resolver.
type SubjectResolverFunc func(ctx context.Context, identity Identity) (string, error)

// ResolveSubject implements SubjectResolver interface.
func (f SubjectResolverFunc) ResolveSubject(ctx context.Context, identity Identity) (string, error) {
	if f == nil {
		return "", fmt.Errorf("%w: subject resolver function is nil", ErrInvalidRequest)
	}
	return f(ctx, identity)
}

// DefaultSubjectResolver is the default resolver that uses identity.ID() or identity.Name().
type DefaultSubjectResolver struct{}

// NewDefaultSubjectResolver creates a default subject resolver.
func NewDefaultSubjectResolver() *DefaultSubjectResolver {
	return &DefaultSubjectResolver{}
}

// ResolveSubject returns identity.ID() if available, otherwise identity.Name().
func (r *DefaultSubjectResolver) ResolveSubject(ctx context.Context, identity Identity) (string, error) {
	if identity == nil || !identity.IsAuthenticated() {
		return "", fmt.Errorf("%w: cannot resolve subject from unauthenticated identity", ErrUnauthorized)
	}

	subject := strings.TrimSpace(identity.ID())
	if subject == "" {
		subject = strings.TrimSpace(identity.Name())
	}
	if subject == "" {
		return "", fmt.Errorf("%w: identity has no ID or name for subject resolution", ErrUnauthorized)
	}

	return subject, nil
}

// TenantSubjectResolver resolves identity to a tenant-namespaced subject.
// Format: tenant/{tenantID}/subject/{subjectKey}
type TenantSubjectResolver struct {
	tenantExtractor func(Identity) string
	subjectKey      string
}

// TenantSubjectResolverConfig configures a tenant subject resolver.
type TenantSubjectResolverConfig struct {
	// TenantExtractor extracts tenant ID from identity.
	// If nil, uses identity.Attributes()["tenant"] by default.
	TenantExtractor func(Identity) string
	// SubjectKey is the key used in the subject path.
	// Defaults to "subject" if empty.
	SubjectKey string
}

// NewTenantSubjectResolver creates a tenant-namespaced subject resolver.
func NewTenantSubjectResolver(cfg TenantSubjectResolverConfig) *TenantSubjectResolver {
	resolver := &TenantSubjectResolver{
		subjectKey: "subject",
	}

	if cfg.TenantExtractor != nil {
		resolver.tenantExtractor = cfg.TenantExtractor
	} else {
		resolver.tenantExtractor = func(identity Identity) string {
			if identity == nil || identity.Attributes() == nil {
				return ""
			}
			return strings.TrimSpace(identity.Attributes()["tenant"])
		}
	}

	if cfg.SubjectKey != "" {
		resolver.subjectKey = strings.TrimSpace(cfg.SubjectKey)
	}

	return resolver
}

// ResolveSubject returns a tenant-namespaced subject string.
// Format: tenant/{tenantID}/{subjectKey}/{identityID}
func (r *TenantSubjectResolver) ResolveSubject(ctx context.Context, identity Identity) (string, error) {
	if identity == nil || !identity.IsAuthenticated() {
		return "", fmt.Errorf("%w: cannot resolve subject from unauthenticated identity", ErrUnauthorized)
	}

	tenantID := r.tenantExtractor(identity)
	if tenantID == "" {
		return "", fmt.Errorf("%w: tenant ID is required for tenant subject resolution", ErrInvalidRequest)
	}

	identityID := strings.TrimSpace(identity.ID())
	if identityID == "" {
		identityID = strings.TrimSpace(identity.Name())
	}
	if identityID == "" {
		return "", fmt.Errorf("%w: identity has no ID or name for subject resolution", ErrUnauthorized)
	}

	return fmt.Sprintf("tenant/%s/%s/%s", tenantID, r.subjectKey, identityID), nil
}

// PrefixSubjectResolver prepends a fixed prefix to the resolved subject.
// Format: {prefix}/{identityID}
type PrefixSubjectResolver struct {
	prefix string
}

// NewPrefixSubjectResolver creates a prefix-based subject resolver.
func NewPrefixSubjectResolver(prefix string) *PrefixSubjectResolver {
	return &PrefixSubjectResolver{
		prefix: strings.TrimSpace(prefix),
	}
}

// ResolveSubject returns a prefixed subject string.
func (r *PrefixSubjectResolver) ResolveSubject(ctx context.Context, identity Identity) (string, error) {
	if identity == nil || !identity.IsAuthenticated() {
		return "", fmt.Errorf("%w: cannot resolve subject from unauthenticated identity", ErrUnauthorized)
	}

	identityID := strings.TrimSpace(identity.ID())
	if identityID == "" {
		identityID = strings.TrimSpace(identity.Name())
	}
	if identityID == "" {
		return "", fmt.Errorf("%w: identity has no ID or name for subject resolution", ErrUnauthorized)
	}

	if r.prefix == "" {
		return identityID, nil
	}

	return fmt.Sprintf("%s/%s", r.prefix, identityID), nil
}

// MappedSubjectResolver applies a custom mapping function to resolve subject.
type MappedSubjectResolver struct {
	mapper func(ctx context.Context, identity Identity) (string, error)
}

// NewMappedSubjectResolver creates a mapped subject resolver.
func NewMappedSubjectResolver(mapper func(ctx context.Context, identity Identity) (string, error)) *MappedSubjectResolver {
	return &MappedSubjectResolver{
		mapper: mapper,
	}
}

// ResolveSubject applies the custom mapper function.
func (r *MappedSubjectResolver) ResolveSubject(ctx context.Context, identity Identity) (string, error) {
	if r.mapper == nil {
		return "", fmt.Errorf("%w: subject mapper function is nil", ErrInvalidRequest)
	}
	return r.mapper(ctx, identity)
}

// ComposedSubjectResolver composes multiple resolvers in a chain.
// The first non-error result is returned.
type ComposedSubjectResolver struct {
	resolvers []SubjectResolver
}

// NewComposedSubjectResolver creates a composed subject resolver.
func NewComposedSubjectResolver(resolvers ...SubjectResolver) *ComposedSubjectResolver {
	return &ComposedSubjectResolver{
		resolvers: resolvers,
	}
}

// ResolveSubject tries each resolver in order until one succeeds.
func (r *ComposedSubjectResolver) ResolveSubject(ctx context.Context, identity Identity) (string, error) {
	if len(r.resolvers) == 0 {
		return "", fmt.Errorf("%w: no subject resolvers configured", ErrInvalidRequest)
	}

	var lastErr error
	for _, resolver := range r.resolvers {
		if resolver == nil {
			continue
		}
		subject, err := resolver.ResolveSubject(ctx, identity)
		if err == nil && subject != "" {
			return subject, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("%w: all subject resolvers failed", ErrInvalidRequest)
}
