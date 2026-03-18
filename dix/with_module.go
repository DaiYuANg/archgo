package dix

func WithModuleProviders(providers ...ProviderFunc) ModuleOption {
	return func(spec *moduleSpec) { spec.providers.Add(providers...) }
}

func WithModuleSetups(setups ...SetupFunc) ModuleOption {
	return func(spec *moduleSpec) { spec.setups.Add(setups...) }
}

func WithModuleInvokes(invokes ...InvokeFunc) ModuleOption {
	return func(spec *moduleSpec) { spec.invokes.Add(invokes...) }
}

func WithModuleHooks(hooks ...HookFunc) ModuleOption {
	return func(spec *moduleSpec) { spec.hooks.Add(hooks...) }
}

func WithModuleImports(modules ...Module) ModuleOption {
	return func(spec *moduleSpec) { spec.imports.Add(modules...) }
}

func WithModuleProfiles(profiles ...Profile) ModuleOption {
	return func(spec *moduleSpec) { spec.profiles.Add(profiles...) }
}

func WithModuleExcludeProfiles(profiles ...Profile) ModuleOption {
	return func(spec *moduleSpec) { spec.excludeProfiles.Add(profiles...) }
}

func WithModuleDescription(desc string) ModuleOption {
	return func(spec *moduleSpec) { spec.description = desc }
}

func WithModuleTags(tags ...string) ModuleOption {
	return func(spec *moduleSpec) { spec.tags.Add(tags...) }
}

func WithModuleSetup(fn func(*Container, Lifecycle) error) ModuleOption {
	return WithModuleSetups(Setup(fn))
}

func WithModuleDisabled(disabled bool) ModuleOption {
	return func(spec *moduleSpec) { spec.disabled = disabled }
}
