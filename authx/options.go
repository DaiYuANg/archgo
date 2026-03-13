package authx

type EngineOption func(*Engine)

func WithAuthenticationManager(manager AuthenticationManager) EngineOption {
	return func(engine *Engine) {
		engine.SetAuthenticationManager(manager)
	}
}

func WithAuthorizer(authorizer Authorizer) EngineOption {
	return func(engine *Engine) {
		engine.SetAuthorizer(authorizer)
	}
}

func WithHook(hook Hook) EngineOption {
	return func(engine *Engine) {
		engine.AddHook(hook)
	}
}
