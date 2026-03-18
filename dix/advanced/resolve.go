package advanced

import (
	"github.com/DaiYuANg/arcgo/dix"
	do "github.com/samber/do/v2"
)

func ResolveInjectorAs[T any](injector do.Injector) (T, error) {
	return do.InvokeNamed[T](injector, typedName[T]())
}

func MustResolveInjectorAs[T any](injector do.Injector) T {
	return do.MustInvokeNamed[T](injector, typedName[T]())
}

func ResolveRuntimeAs[T any](rt *dix.Runtime) (T, error) {
	if rt == nil {
		var zero T
		return zero, do.ErrServiceNotFound
	}
	return ResolveInjectorAs[T](rt.Raw())
}

func MustResolveRuntimeAs[T any](rt *dix.Runtime) T {
	return MustResolveInjectorAs[T](rt.Raw())
}

func ResolveNamedAs[T any](c *dix.Container, name string) (T, error) {
	return do.InvokeNamed[T](c.Raw(), name)
}

func MustResolveNamedAs[T any](c *dix.Container, name string) T {
	return do.MustInvokeNamed[T](c.Raw(), name)
}

func ResolveAssignableAs[T any](c *dix.Container) (T, error) {
	return do.InvokeAs[T](c.Raw())
}

func MustResolveAssignableAs[T any](c *dix.Container) T {
	return do.MustInvokeAs[T](c.Raw())
}
