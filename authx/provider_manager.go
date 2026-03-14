package authx

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// ProviderManager routes authentication credential to provider by credential concrete type.
type ProviderManager struct {
	mu        sync.RWMutex
	providers map[reflect.Type]AuthenticationProvider
}

func NewProviderManager(providers ...AuthenticationProvider) *ProviderManager {
	manager := &ProviderManager{providers: make(map[reflect.Type]AuthenticationProvider)}
	manager.Register(providers...)
	return manager
}

func (manager *ProviderManager) Register(providers ...AuthenticationProvider) {
	if manager == nil || len(providers) == 0 {
		return
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()
	for _, provider := range providers {
		if provider == nil {
			continue
		}
		credentialType := provider.CredentialType()
		if credentialType == nil {
			continue
		}
		manager.providers[credentialType] = provider
	}
}

func (manager *ProviderManager) Authenticate(
	ctx context.Context,
	credential any,
) (AuthenticationResult, error) {
	if credential == nil {
		return AuthenticationResult{}, ErrInvalidAuthenticationCredential
	}
	if manager == nil {
		return AuthenticationResult{}, ErrAuthenticationManagerNotConfigured
	}

	credentialType := reflect.TypeOf(credential)
	manager.mu.RLock()
	provider, ok := manager.providers[credentialType]
	manager.mu.RUnlock()
	if !ok {
		return AuthenticationResult{}, fmt.Errorf("%w: %v", ErrAuthenticationProviderNotFound, credentialType)
	}

	result, err := provider.AuthenticateAny(ctx, credential)
	if err != nil {
		return AuthenticationResult{}, fmt.Errorf("%w: %v", ErrUnauthenticated, err)
	}
	return result, nil
}
