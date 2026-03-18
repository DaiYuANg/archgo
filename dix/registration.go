package dix

import "github.com/samber/lo"

// ServiceRef identifies a service in the container graph.
// Typed services should use TypedService[T](). Named services should use NamedService(name).
type ServiceRef struct {
	Name string
}

func TypedService[T any]() ServiceRef {
	return ServiceRef{Name: serviceNameOf[T]()}
}

func NamedService(name string) ServiceRef {
	return ServiceRef{Name: name}
}

type ProviderMetadata struct {
	Label        string
	Output       ServiceRef
	Dependencies []ServiceRef
	Raw          bool
}

type InvokeMetadata struct {
	Label        string
	Dependencies []ServiceRef
	Raw          bool
}

type HookKind string

const (
	HookKindStart HookKind = "start"
	HookKindStop  HookKind = "stop"
)

type HookMetadata struct {
	Label        string
	Kind         HookKind
	Dependencies []ServiceRef
	Raw          bool
}

type SetupMetadata struct {
	Label         string
	Dependencies  []ServiceRef
	Provides      []ServiceRef
	Overrides     []ServiceRef
	GraphMutation bool
	Raw           bool
}

func NewProviderFunc(register func(*Container), meta ProviderMetadata) ProviderFunc {
	return ProviderFunc{
		register: register,
		meta:     normalizeProviderMetadata(meta),
	}
}

func NewInvokeFunc(run func(*Container) error, meta InvokeMetadata) InvokeFunc {
	return InvokeFunc{
		run:  run,
		meta: normalizeInvokeMetadata(meta),
	}
}

func NewHookFunc(register func(*Container, Lifecycle), meta HookMetadata) HookFunc {
	return HookFunc{
		register: register,
		meta:     normalizeHookMetadata(meta),
	}
}

func NewSetupFunc(run func(*Container, Lifecycle) error, meta SetupMetadata) SetupFunc {
	return SetupFunc{
		run:  run,
		meta: normalizeSetupMetadata(meta),
	}
}

func normalizeProviderMetadata(meta ProviderMetadata) ProviderMetadata {
	if meta.Label == "" {
		meta.Label = "Provider"
	}
	meta.Dependencies = normalizeServiceRefs(meta.Dependencies)
	return meta
}

func normalizeInvokeMetadata(meta InvokeMetadata) InvokeMetadata {
	if meta.Label == "" {
		meta.Label = "Invoke"
	}
	meta.Dependencies = normalizeServiceRefs(meta.Dependencies)
	return meta
}

func normalizeHookMetadata(meta HookMetadata) HookMetadata {
	if meta.Label == "" {
		meta.Label = "Hook"
	}
	meta.Dependencies = normalizeServiceRefs(meta.Dependencies)
	return meta
}

func normalizeSetupMetadata(meta SetupMetadata) SetupMetadata {
	if meta.Label == "" {
		meta.Label = "Setup"
	}
	meta.Dependencies = normalizeServiceRefs(meta.Dependencies)
	meta.Provides = normalizeServiceRefs(meta.Provides)
	meta.Overrides = normalizeServiceRefs(meta.Overrides)
	return meta
}

func normalizeServiceRefs(refs []ServiceRef) []ServiceRef {
	return lo.Filter(refs, func(ref ServiceRef, _ int) bool {
		return ref.Name != ""
	})
}
