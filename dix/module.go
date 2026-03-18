package dix

import "github.com/samber/lo"

// ModuleOption configures a Module during construction.
type ModuleOption func(*moduleSpec)

// NewModule creates an immutable module specification.
func NewModule(name string, opts ...ModuleOption) Module {
	spec := &moduleSpec{name: name}
	lo.ForEach(lo.Filter(opts, func(opt ModuleOption, _ int) bool {
		return opt != nil
	}), func(opt ModuleOption, _ int) {
		opt(spec)
	})
	return Module{spec: spec}
}

func (m Module) Name() string {
	if m.spec == nil {
		return ""
	}
	return m.spec.name
}

func (m Module) Description() string {
	if m.spec == nil {
		return ""
	}
	return m.spec.description
}

func (m Module) Tags() []string {
	if m.spec == nil {
		return nil
	}
	return m.spec.tags.Values()
}

func (m Module) Imports() []Module {
	if m.spec == nil {
		return nil
	}
	return m.spec.imports.Values()
}
