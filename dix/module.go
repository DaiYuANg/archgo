package dix

import (
	"fmt"
	"strings"

	do "github.com/samber/do/v2"
)

// ProviderFunc is a function that registers providers in a container.
type ProviderFunc func(c *Container)

// InvokeFunc is a function that executes an invoke with a container.
type InvokeFunc func(c *Container) error

// Module is the core building block for composing applications.
type Module struct {
	Name            string
	Description     string
	Providers       []ProviderFunc
	Invokes         []InvokeFunc
	Imports         []Module
	Profiles        []Profile
	ExcludeProfiles []Profile
	Disabled        bool
	Setup           SetupFunc
	DoSetup         DoSetupFunc
	Tags            []string
}

// SetupFunc is called during container build.
type SetupFunc func(c *Container, lc Lifecycle) error

// DoSetupFunc is a narrow escape hatch for do-specific integration work.
// Keep this at the module/framework boundary, not in normal business code.
type DoSetupFunc func(raw do.Injector) error

// NewModule creates a new Module with the given name and options.
func NewModule(name string, opts ...ModuleOption) Module {
	m := Module{
		Name:      name,
		Providers: make([]ProviderFunc, 0),
		Invokes:   make([]InvokeFunc, 0),
		Imports:   make([]Module, 0),
		Profiles:  make([]Profile, 0),
		Tags:      make([]string, 0),
		Disabled:  false,
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// flattenModules recursively flattens all imported modules.
func flattenModules(modules []Module, profile Profile) ([]Module, error) {
	result := make([]Module, 0)
	visited := make(map[string]struct{})
	visiting := make(map[string]struct{})

	var walk func(mod Module, path []string) error
	walk = func(mod Module, path []string) error {
		if mod.Disabled || !isActiveForProfile(mod, profile) {
			return nil
		}
		key := moduleKey(mod)
		if _, ok := visited[key]; ok {
			return nil
		}
		if _, ok := visiting[key]; ok {
			return fmt.Errorf("module import cycle detected: %s -> %s", formatModulePath(path), key)
		}
		visiting[key] = struct{}{}
		path = append(path, key)
		for _, imported := range mod.Imports {
			if err := walk(imported, path); err != nil {
				return err
			}
		}
		delete(visiting, key)
		visited[key] = struct{}{}
		result = append(result, mod)
		return nil
	}

	for _, mod := range modules {
		if err := walk(mod, nil); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func moduleKey(mod Module) string {
	if mod.Name != "" {
		return mod.Name
	}
	return fmt.Sprintf("<anonymous:%p>", &mod)
}
func formatModulePath(path []string) string {
	if len(path) == 0 {
		return "<root>"
	}
	return strings.Join(path, " -> ")
}
func isActiveForProfile(mod Module, profile Profile) bool {
	if mod.Disabled {
		return false
	}
	for _, p := range mod.ExcludeProfiles {
		if p == profile {
			return false
		}
	}
	if len(mod.Profiles) > 0 {
		for _, p := range mod.Profiles {
			if p == profile {
				return true
			}
		}
		return false
	}
	return true
}
