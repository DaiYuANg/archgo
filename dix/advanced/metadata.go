package advanced

import (
	collectionlist "github.com/DaiYuANg/arcgo/collectionx/list"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/samber/lo"
)

func serviceRefs(refs ...dix.ServiceRef) []dix.ServiceRef {
	if len(refs) == 0 {
		return nil
	}

	items := collectionlist.NewListWithCapacity[dix.ServiceRef](len(refs))
	items.MergeSlice(lo.Filter(refs, func(ref dix.ServiceRef, _ int) bool {
		return ref.Name != ""
	}))
	return items.Values()
}

func newProvider(
	label string,
	output dix.ServiceRef,
	register func(*dix.Container),
	deps ...dix.ServiceRef,
) dix.ProviderFunc {
	return dix.NewProviderFunc(register, dix.ProviderMetadata{
		Label:        label,
		Output:       output,
		Dependencies: serviceRefs(deps...),
	})
}

func newSetup(
	label string,
	run func(*dix.Container) error,
	dependencies []dix.ServiceRef,
	provides []dix.ServiceRef,
	overrides []dix.ServiceRef,
	graphMutation bool,
	raw bool,
) dix.SetupFunc {
	return dix.NewSetupFunc(func(c *dix.Container, _ dix.Lifecycle) error {
		return run(c)
	}, dix.SetupMetadata{
		Label:         label,
		Dependencies:  serviceRefs(dependencies...),
		Provides:      serviceRefs(provides...),
		Overrides:     serviceRefs(overrides...),
		GraphMutation: graphMutation,
		Raw:           raw,
	})
}
