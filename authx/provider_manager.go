package authx

import (
	"context"
	"fmt"
	"reflect"

	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/samber/lo"
)

// ProviderManager routes authentication credential to provider by credential concrete type.
type ProviderManager struct {
	providers collectionx.ConcurrentMap[reflect.Type, AuthenticationProvider]
}

func NewProviderManager(providers ...AuthenticationProvider) *ProviderManager {
	manager := &ProviderManager{providers: collectionx.NewConcurrentMap[reflect.Type, AuthenticationProvider]()}
	manager.Register(providers...)
	return manager
}

func (manager *ProviderManager) Register(providers ...AuthenticationProvider) {
	if manager == nil || manager.providers == nil {
		return
	}

	validProviders := lo.Filter(providers, func(provider AuthenticationProvider, _ int) bool {
		return provider != nil && provider.CredentialType() != nil
	})

	lo.ForEach(validProviders, func(provider AuthenticationProvider, _ int) {
		manager.providers.Set(provider.CredentialType(), provider)
	})
}

func (manager *ProviderManager) Authenticate(
	ctx context.Context,
	credential any,
) (AuthenticationResult, error) {
	if credential == nil {
		return AuthenticationResult{}, ErrInvalidAuthenticationCredential
	}
	if manager == nil || manager.providers == nil {
		return AuthenticationResult{}, ErrAuthenticationManagerNotConfigured
	}

	provider, ok := manager.providers.GetOption(reflect.TypeOf(credential)).Get()
	if !ok {
		return AuthenticationResult{}, fmt.Errorf("%w: %v", ErrAuthenticationProviderNotFound, reflect.TypeOf(credential))
	}

	result, err := provider.AuthenticateAny(ctx, credential)
	if err != nil {
		return AuthenticationResult{}, fmt.Errorf("%w: %v", ErrUnauthenticated, err)
	}
	return result, nil
}
