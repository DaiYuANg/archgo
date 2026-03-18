package dix

// ModuleOption configures a Module
type ModuleOption func(*Module)

func WithModuleProviders(providers ...ProviderFunc) ModuleOption {
	return func(m *Module) { m.Providers = append(m.Providers, providers...) }
}
func WithModuleInvokes(invokes ...InvokeFunc) ModuleOption {
	return func(m *Module) { m.Invokes = append(m.Invokes, invokes...) }
}
func WithModuleImports(modules ...Module) ModuleOption {
	return func(m *Module) { m.Imports = append(m.Imports, modules...) }
}
func WithModuleProfiles(profiles ...Profile) ModuleOption {
	return func(m *Module) { m.Profiles = append(m.Profiles, profiles...) }
}
func WithModuleExcludeProfiles(profiles ...Profile) ModuleOption {
	return func(m *Module) { m.ExcludeProfiles = append(m.ExcludeProfiles, profiles...) }
}
func WithModuleDescription(desc string) ModuleOption {
	return func(m *Module) { m.Description = desc }
}
func WithModuleTags(tags ...string) ModuleOption {
	return func(m *Module) { m.Tags = append(m.Tags, tags...) }
}
func WithModuleSetup(fn SetupFunc) ModuleOption {
	return func(m *Module) { m.Setup = fn }
}
func WithModuleDoSetup(fn DoSetupFunc) ModuleOption {
	return func(m *Module) { m.DoSetup = fn }
}
func WithModuleDisabled(disabled bool) ModuleOption {
	return func(m *Module) { m.Disabled = disabled }
}
