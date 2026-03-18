package advanced

import (
	"github.com/DaiYuANg/arcgo/dix"
	do "github.com/samber/do/v2"
)

func NamedValue[T any](name string, value T) dix.ProviderFunc {
	return newProvider("NamedValue", dix.NamedService(name), func(c *dix.Container) {
		do.ProvideNamedValue(c.Raw(), name, value)
	})
}

func NamedProvider0[T any](name string, fn func() T) dix.ProviderFunc {
	return newProvider("NamedProvider0", dix.NamedService(name), func(c *dix.Container) {
		do.ProvideNamed(c.Raw(), name, func(do.Injector) (T, error) { return fn(), nil })
	})
}

func NamedProvider1[T, D1 any](name string, fn func(D1) T) dix.ProviderFunc {
	return newProvider("NamedProvider1", dix.NamedService(name), func(c *dix.Container) {
		do.ProvideNamed(c.Raw(), name, func(i do.Injector) (T, error) {
			d1, err := invokeTyped[D1](i)
			if err != nil {
				var zero T
				return zero, err
			}
			return fn(d1), nil
		})
	}, dix.TypedService[D1]())
}

func NamedProvider2[T, D1, D2 any](name string, fn func(D1, D2) T) dix.ProviderFunc {
	return newProvider("NamedProvider2", dix.NamedService(name), func(c *dix.Container) {
		do.ProvideNamed(c.Raw(), name, func(i do.Injector) (T, error) {
			d1, err := invokeTyped[D1](i)
			if err != nil {
				var zero T
				return zero, err
			}
			d2, err := invokeTyped[D2](i)
			if err != nil {
				var zero T
				return zero, err
			}
			return fn(d1, d2), nil
		})
	}, dix.TypedService[D1](), dix.TypedService[D2]())
}

func NamedProvider3[T, D1, D2, D3 any](name string, fn func(D1, D2, D3) T) dix.ProviderFunc {
	return newProvider("NamedProvider3", dix.NamedService(name), func(c *dix.Container) {
		do.ProvideNamed(c.Raw(), name, func(i do.Injector) (T, error) {
			d1, err := invokeTyped[D1](i)
			if err != nil {
				var zero T
				return zero, err
			}
			d2, err := invokeTyped[D2](i)
			if err != nil {
				var zero T
				return zero, err
			}
			d3, err := invokeTyped[D3](i)
			if err != nil {
				var zero T
				return zero, err
			}
			return fn(d1, d2, d3), nil
		})
	}, dix.TypedService[D1](), dix.TypedService[D2](), dix.TypedService[D3]())
}

func TransientProvider0[T any](fn func() T) dix.ProviderFunc {
	name := typedName[T]()
	return newProvider("TransientProvider0", dix.TypedService[T](), func(c *dix.Container) {
		do.ProvideNamedTransient(c.Raw(), name, func(do.Injector) (T, error) { return fn(), nil })
	})
}

func TransientProvider1[T, D1 any](fn func(D1) T) dix.ProviderFunc {
	name := typedName[T]()
	return newProvider("TransientProvider1", dix.TypedService[T](), func(c *dix.Container) {
		do.ProvideNamedTransient(c.Raw(), name, func(i do.Injector) (T, error) {
			d1, err := invokeTyped[D1](i)
			if err != nil {
				var zero T
				return zero, err
			}
			return fn(d1), nil
		})
	}, dix.TypedService[D1]())
}

func NamedTransientProvider0[T any](name string, fn func() T) dix.ProviderFunc {
	return newProvider("NamedTransientProvider0", dix.NamedService(name), func(c *dix.Container) {
		do.ProvideNamedTransient(c.Raw(), name, func(do.Injector) (T, error) { return fn(), nil })
	})
}

func NamedTransientProvider1[T, D1 any](name string, fn func(D1) T) dix.ProviderFunc {
	return newProvider("NamedTransientProvider1", dix.NamedService(name), func(c *dix.Container) {
		do.ProvideNamedTransient(c.Raw(), name, func(i do.Injector) (T, error) {
			d1, err := invokeTyped[D1](i)
			if err != nil {
				var zero T
				return zero, err
			}
			return fn(d1), nil
		})
	}, dix.TypedService[D1]())
}
