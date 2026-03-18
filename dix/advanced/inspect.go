package advanced

import (
	collectionmap "github.com/DaiYuANg/arcgo/collectionx/mapping"
	"github.com/DaiYuANg/arcgo/dix"
	do "github.com/samber/do/v2"
	"github.com/samber/lo"
)

type Inspection struct {
	ScopeTree         string
	ProvidedServices  []do.ServiceDescription
	InvokedServices   []do.ServiceDescription
	NamedDependencies map[string]string
}

type InspectOptions struct {
	IncludeScopeTree        bool
	IncludeProvidedServices bool
	IncludeInvokedServices  bool
	IncludeNamedDeps        bool
}

func DefaultInspectOptions() InspectOptions {
	return InspectOptions{
		IncludeScopeTree:        true,
		IncludeProvidedServices: true,
		IncludeInvokedServices:  true,
		IncludeNamedDeps:        true,
	}
}

func ExplainScopeTree(rt *dix.Runtime) string {
	if rt == nil {
		return ""
	}

	explainedScope := do.ExplainInjector(rt.Raw())
	return explainedScope.String()
}

func ListProvidedServices(rt *dix.Runtime) []do.ServiceDescription {
	if rt == nil {
		return nil
	}

	return rt.Raw().ListProvidedServices()
}

func ListInvokedServices(rt *dix.Runtime) []do.ServiceDescription {
	if rt == nil {
		return nil
	}

	return rt.Raw().ListInvokedServices()
}

func ExplainNamedDependencies(rt *dix.Runtime, namedServices ...string) map[string]string {
	if rt == nil || len(namedServices) == 0 {
		return nil
	}

	dependencies := collectionmap.NewMapWithCapacity[string, string](len(namedServices))
	lo.ForEach(namedServices, func(name string, _ int) {
		if desc, found := do.ExplainNamedService(rt.Raw(), name); found {
			dependencies.Set(name, desc.String())
		}
	})

	return dependencies.All()
}

func InspectRuntime(rt *dix.Runtime, namedServices ...string) Inspection {
	return InspectRuntimeWithOptions(rt, DefaultInspectOptions(), namedServices...)
}

func InspectRuntimeWithOptions(rt *dix.Runtime, opts InspectOptions, namedServices ...string) Inspection {
	if rt == nil {
		return Inspection{}
	}

	var scopeTree string
	if opts.IncludeScopeTree {
		scopeTree = ExplainScopeTree(rt)
	}

	var provided []do.ServiceDescription
	if opts.IncludeProvidedServices {
		provided = ListProvidedServices(rt)
	}

	var invoked []do.ServiceDescription
	if opts.IncludeInvokedServices {
		invoked = ListInvokedServices(rt)
	}

	var namedDependencies map[string]string
	if opts.IncludeNamedDeps && len(namedServices) > 0 {
		namedDependencies = ExplainNamedDependencies(rt, namedServices...)
	}

	return Inspection{
		ScopeTree:         scopeTree,
		ProvidedServices:  provided,
		InvokedServices:   invoked,
		NamedDependencies: namedDependencies,
	}
}
