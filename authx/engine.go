package authx

import (
	"context"

	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/samber/lo"
)

const (
	dependencyAuthnKey = "authn"
	dependencyAuthzKey = "authz"
)

// Engine separates authentication (Check) and authorization (Can).
type Engine struct {
	deps  collectionx.ConcurrentMap[string, any]
	hooks collectionx.ConcurrentList[Hook]
}

func NewEngine(opts ...EngineOption) *Engine {
	engine := &Engine{
		deps:  collectionx.NewConcurrentMap[string, any](),
		hooks: collectionx.NewConcurrentList[Hook](),
	}
	lo.ForEach(opts, func(opt EngineOption, _ int) {
		if opt != nil {
			opt(engine)
		}
	})
	return engine
}

func (engine *Engine) SetAuthenticationManager(manager AuthenticationManager) {
	if engine == nil || engine.deps == nil {
		return
	}
	if manager == nil {
		_ = engine.deps.Delete(dependencyAuthnKey)
		return
	}
	engine.deps.Set(dependencyAuthnKey, manager)
}

func (engine *Engine) SetAuthorizer(authorizer Authorizer) {
	if engine == nil || engine.deps == nil {
		return
	}
	if authorizer == nil {
		_ = engine.deps.Delete(dependencyAuthzKey)
		return
	}
	engine.deps.Set(dependencyAuthzKey, authorizer)
}

func (engine *Engine) AddHook(hook Hook) {
	if engine == nil || hook == nil {
		return
	}
	if engine.hooks == nil {
		engine.hooks = collectionx.NewConcurrentList[Hook]()
	}
	engine.hooks.Add(hook)
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

	if err := firstHookError(hooks, func(hook Hook) error {
		return hook.BeforeCheck(ctx, credential)
	}); err != nil {
		return AuthenticationResult{}, err
	}

	result, err := authn.Authenticate(ctx, credential)
	lo.ForEach(hooks, func(hook Hook, _ int) {
		hook.AfterCheck(ctx, credential, result, err)
	})
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

	if err := firstHookError(hooks, func(hook Hook) error {
		return hook.BeforeCan(ctx, input)
	}); err != nil {
		return Decision{}, err
	}

	decision, err := authorizer.Authorize(ctx, input)
	lo.ForEach(hooks, func(hook Hook, _ int) {
		hook.AfterCan(ctx, input, decision, err)
	})
	if err != nil {
		return Decision{}, err
	}
	return decision, nil
}

func (engine *Engine) snapshotCheckDependencies() (AuthenticationManager, []Hook) {
	if engine == nil {
		return nil, nil
	}

	var authn AuthenticationManager
	if engine.deps != nil {
		if value, ok := engine.deps.Get(dependencyAuthnKey); ok {
			authn, _ = value.(AuthenticationManager)
		}
	}

	hooks := []Hook(nil)
	if engine.hooks != nil {
		hooks = engine.hooks.Values()
	}
	return authn, hooks
}

func (engine *Engine) snapshotCanDependencies() (Authorizer, []Hook) {
	if engine == nil {
		return nil, nil
	}

	var authorizer Authorizer
	if engine.deps != nil {
		if value, ok := engine.deps.Get(dependencyAuthzKey); ok {
			authorizer, _ = value.(Authorizer)
		}
	}

	hooks := []Hook(nil)
	if engine.hooks != nil {
		hooks = engine.hooks.Values()
	}
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

func firstHookError(hooks []Hook, fn func(Hook) error) error {
	var firstErr error
	_, found := lo.Find(hooks, func(hook Hook) bool {
		firstErr = fn(hook)
		return firstErr != nil
	})
	if found {
		return firstErr
	}
	return nil
}
