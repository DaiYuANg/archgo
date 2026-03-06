package authx

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/DaiYuANg/arcgo/observability"
)

// ManagerOption configures manager construction.
type ManagerOption func(cfg *managerConfig) error

type managerConfig struct {
	providers     []IdentityProvider
	sources       []PolicySource
	logger        *slog.Logger
	observability observability.Observability
}

// WithProvider appends one authentication provider to provider chain.
func WithProvider(provider IdentityProvider) ManagerOption {
	return func(cfg *managerConfig) error {
		if cfg == nil {
			return fmt.Errorf("%w: manager config is nil", ErrInvalidAuthenticator)
		}
		if provider == nil {
			return fmt.Errorf("%w: identity provider is nil", ErrInvalidAuthenticator)
		}
		cfg.providers = append(cfg.providers, provider)
		return nil
	}
}

// WithSource appends one policy source.
func WithSource(source PolicySource) ManagerOption {
	return func(cfg *managerConfig) error {
		if cfg == nil {
			return fmt.Errorf("%w: manager config is nil", ErrInvalidPolicy)
		}
		if source == nil {
			return fmt.Errorf("%w: policy source is nil", ErrInvalidPolicy)
		}
		cfg.sources = append(cfg.sources, source)
		return nil
	}
}

// WithLogger sets slog logger used by authx runtime nodes.
func WithLogger(logger *slog.Logger) ManagerOption {
	return func(cfg *managerConfig) error {
		if cfg == nil {
			return fmt.Errorf("%w: manager config is nil", ErrInvalidAuthenticator)
		}
		cfg.logger = logger
		return nil
	}
}

// WithObservability sets optional observability integration.
func WithObservability(obs observability.Observability) ManagerOption {
	return func(cfg *managerConfig) error {
		if cfg == nil {
			return fmt.Errorf("%w: manager config is nil", ErrInvalidAuthenticator)
		}
		cfg.observability = obs
		return nil
	}
}

// TypedProviderLoader loads typed domain payload by principal.
type TypedProviderLoader[T any] interface {
	LoadByPrincipal(ctx context.Context, principal string) (T, error)
}

// TypedProviderMapper maps typed domain payload into AuthX user details.
type TypedProviderMapper[T any] interface {
	MapToUserDetails(ctx context.Context, principal string, payload T) (UserDetails, error)
}

// MappedProvider is a generic provider contract for custom user model mapping.
type MappedProvider[T any] interface {
	TypedProviderLoader[T]
	TypedProviderMapper[T]
}

type mappedIdentityProvider[T any] struct {
	provider MappedProvider[T]
}

// NewMappedIdentityProvider creates a generic strong-typed provider adapter.
func NewMappedIdentityProvider[T any](provider MappedProvider[T]) (IdentityProvider, error) {
	if provider == nil {
		return nil, fmt.Errorf("%w: mapped provider is nil", ErrInvalidAuthenticator)
	}

	return &mappedIdentityProvider[T]{
		provider: provider,
	}, nil
}

// WithMappedProvider appends a generic strong-typed provider adapter.
func WithMappedProvider[T any](provider MappedProvider[T]) ManagerOption {
	return func(cfg *managerConfig) error {
		adaptedProvider, err := NewMappedIdentityProvider(provider)
		if err != nil {
			return err
		}
		return WithProvider(adaptedProvider)(cfg)
	}
}

func (p *mappedIdentityProvider[T]) LoadByPrincipal(ctx context.Context, principal string) (UserDetails, error) {
	if p == nil || p.provider == nil {
		return UserDetails{}, fmt.Errorf("%w: mapped provider is not configured", ErrInvalidAuthenticator)
	}

	payload, err := p.provider.LoadByPrincipal(ctx, principal)
	if err != nil {
		return UserDetails{}, err
	}

	details, err := p.provider.MapToUserDetails(ctx, principal, payload)
	if err != nil {
		return UserDetails{}, err
	}
	details.Payload = payload
	return details, nil
}
