package authx

import (
	"context"
	"sync"
)

// Engine separates authentication (Check) and authorization (Can).
type Engine struct {
	mu    sync.RWMutex
	authn AuthenticationManager
	authz Authorizer
	hooks []Hook
}

func NewEngine(opts ...EngineOption) *Engine {
	engine := &Engine{}
	for _, opt := range opts {
		if opt != nil {
			opt(engine)
		}
	}
	return engine
}

func (engine *Engine) SetAuthenticationManager(manager AuthenticationManager) {
	if engine == nil {
		return
	}
	engine.mu.Lock()
	engine.authn = manager
	engine.mu.Unlock()
}

func (engine *Engine) SetAuthorizer(authorizer Authorizer) {
	if engine == nil {
		return
	}
	engine.mu.Lock()
	engine.authz = authorizer
	engine.mu.Unlock()
}

func (engine *Engine) AddHook(hook Hook) {
	if engine == nil || hook == nil {
		return
	}
	engine.mu.Lock()
	next := make([]Hook, len(engine.hooks)+1)
	copy(next, engine.hooks)
	next[len(engine.hooks)] = hook
	engine.hooks = next
	engine.mu.Unlock()
}

// Check authenticates credential and returns principal.
func (engine *Engine) Check(ctx context.Context, credential any) (AuthenticationResult, error) {
	if credential == nil {
		return AuthenticationResult{}, ErrInvalidAuthenticationCredential
	}

	authn, hooks := engine.snapshotCheckDependencies()
	if authn == nil {
		return AuthenticationResult{}, ErrAuthenticationManagerNotConfigured
	}

	for _, hook := range hooks {
		if err := hook.BeforeCheck(ctx, credential); err != nil {
			return AuthenticationResult{}, err
		}
	}

	result, err := authn.Authenticate(ctx, credential)
	for _, hook := range hooks {
		hook.AfterCheck(ctx, credential, result, err)
	}
	if err != nil {
		return AuthenticationResult{}, err
	}
	return result, nil
}

// Can authorizes principal access to action/resource.
func (engine *Engine) Can(ctx context.Context, input AuthorizationModel) (Decision, error) {
	if err := validateAuthorizationModel(input); err != nil {
		return Decision{}, err
	}

	authorizer, hooks := engine.snapshotCanDependencies()
	if authorizer == nil {
		return Decision{}, ErrAuthorizerNotConfigured
	}

	for _, hook := range hooks {
		if err := hook.BeforeCan(ctx, input); err != nil {
			return Decision{}, err
		}
	}

	decision, err := authorizer.Authorize(ctx, input)
	for _, hook := range hooks {
		hook.AfterCan(ctx, input, decision, err)
	}
	if err != nil {
		return Decision{}, err
	}
	return decision, nil
}

func (engine *Engine) snapshotCheckDependencies() (AuthenticationManager, []Hook) {
	if engine == nil {
		return nil, nil
	}

	engine.mu.RLock()
	authn := engine.authn
	hooks := engine.hooks
	engine.mu.RUnlock()
	return authn, hooks
}

func (engine *Engine) snapshotCanDependencies() (Authorizer, []Hook) {
	if engine == nil {
		return nil, nil
	}

	engine.mu.RLock()
	authorizer := engine.authz
	hooks := engine.hooks
	engine.mu.RUnlock()
	return authorizer, hooks
}

func validateAuthorizationModel(input AuthorizationModel) error {
	if input.Action == "" || input.Resource == "" {
		return ErrInvalidAuthorizationModel
	}
	if input.Principal == nil {
		return ErrInvalidAuthorizationModel
	}
	return nil
}
