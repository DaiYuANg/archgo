package authx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	collectionlist "github.com/DaiYuANg/arcgo/collectionx/list"
)

type identityProviderChain struct {
	providers *collectionlist.ConcurrentList[IdentityProvider]
	logger    *slog.Logger
}

func newIdentityProviderChain(providers ...IdentityProvider) (*identityProviderChain, error) {
	chain := &identityProviderChain{
		providers: collectionlist.NewConcurrentList[IdentityProvider](),
		logger:    normalizeLogger(nil).With("component", "authx.provider-chain"),
	}
	if err := chain.Set(providers...); err != nil {
		return nil, err
	}
	return chain, nil
}

func (p *identityProviderChain) SetLogger(logger *slog.Logger) {
	if p == nil {
		return
	}
	p.logger = normalizeLogger(logger).With("component", "authx.provider-chain")
}

func (p *identityProviderChain) Set(providers ...IdentityProvider) error {
	if p == nil || p.providers == nil {
		return fmt.Errorf("%w: identity provider chain is nil", ErrInvalidAuthenticator)
	}
	validatedProviders, err := validateProviders(providers)
	if err != nil {
		return err
	}

	p.providers.Clear()
	p.providers.Add(validatedProviders...)
	p.logger.Debug("provider chain replaced", "providers", len(validatedProviders))
	return nil
}

func (p *identityProviderChain) Add(provider IdentityProvider) error {
	if p == nil || p.providers == nil {
		return fmt.Errorf("%w: identity provider chain is nil", ErrInvalidAuthenticator)
	}
	if provider == nil {
		return fmt.Errorf("%w: identity provider is nil", ErrInvalidAuthenticator)
	}

	p.providers.Add(provider)
	p.logger.Debug("provider added", "provider_type", fmt.Sprintf("%T", provider))
	return nil
}

func (p *identityProviderChain) LoadByPrincipal(ctx context.Context, principal string) (UserDetails, error) {
	if p == nil || p.providers == nil {
		return UserDetails{}, fmt.Errorf("%w: identity provider chain is nil", ErrInvalidAuthenticator)
	}

	providers := p.providers.Values()

	var firstError error
	for index, provider := range providers {
		p.logger.Debug("provider attempt", "provider_index", index, "provider_type", fmt.Sprintf("%T", provider))
		user, err := provider.LoadByPrincipal(ctx, principal)
		if err == nil {
			p.logger.Debug("provider matched", "provider_index", index, "provider_type", fmt.Sprintf("%T", provider))
			return user, nil
		}

		if errors.Is(err, ErrUnauthorized) {
			p.logger.Debug("provider unauthorized", "provider_index", index, "provider_type", fmt.Sprintf("%T", provider))
			continue
		}

		if firstError == nil {
			firstError = err
		}
		p.logger.Warn("provider failed", "provider_index", index, "provider_type", fmt.Sprintf("%T", provider), "error", err.Error())
	}

	if firstError != nil {
		return UserDetails{}, firstError
	}

	return UserDetails{}, ErrUnauthorized
}

func validateProviders(providers []IdentityProvider) ([]IdentityProvider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("%w: identity providers are required", ErrInvalidAuthenticator)
	}

	validatedProviders := make([]IdentityProvider, 0, len(providers))
	for _, provider := range providers {
		if provider == nil {
			return nil, fmt.Errorf("%w: identity provider is nil", ErrInvalidAuthenticator)
		}
		validatedProviders = append(validatedProviders, provider)
	}
	return validatedProviders, nil
}
