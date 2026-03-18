package advanced

import (
	"github.com/DaiYuANg/arcgo/dix"
	do "github.com/samber/do/v2"
)

func DoSetup(fn func(do.Injector) error) dix.SetupFunc {
	return newSetup("DoSetup", func(c *dix.Container) error {
		return fn(c.Raw())
	}, nil, nil, nil, true, true)
}

func BindAlias[From, To any]() dix.SetupFunc {
	return newSetup("BindAlias", func(c *dix.Container) error {
		return do.As[From, To](c.Raw())
	}, []dix.ServiceRef{dix.TypedService[From]()}, []dix.ServiceRef{dix.TypedService[To]()}, nil, false, false)
}

func BindNamedAlias[From, To any](sourceName string, aliasName string) dix.SetupFunc {
	return newSetup("BindNamedAlias", func(c *dix.Container) error {
		return do.AsNamed[From, To](c.Raw(), sourceName, aliasName)
	}, []dix.ServiceRef{dix.NamedService(sourceName)}, []dix.ServiceRef{dix.NamedService(aliasName)}, nil, false, false)
}

func OverrideValue[T any](value T) dix.SetupFunc {
	return NamedOverrideValue(typedName[T](), value)
}

func NamedOverrideValue[T any](name string, value T) dix.SetupFunc {
	return newSetup("OverrideValue", func(c *dix.Container) error {
		do.OverrideNamedValue(c.Raw(), name, value)
		return nil
	}, nil, nil, []dix.ServiceRef{dix.NamedService(name)}, false, false)
}

func Override0[T any](fn func() T) dix.SetupFunc {
	return NamedOverride0(typedName[T](), fn)
}

func NamedOverride0[T any](name string, fn func() T) dix.SetupFunc {
	return newSetup("Override0", func(c *dix.Container) error {
		do.OverrideNamed(c.Raw(), name, func(do.Injector) (T, error) { return fn(), nil })
		return nil
	}, nil, nil, []dix.ServiceRef{dix.NamedService(name)}, false, false)
}

func Override1[T, D1 any](fn func(D1) T) dix.SetupFunc {
	return NamedOverride1(typedName[T](), fn)
}

func NamedOverride1[T, D1 any](name string, fn func(D1) T) dix.SetupFunc {
	return newSetup("Override1", func(c *dix.Container) error {
		do.OverrideNamed(c.Raw(), name, func(i do.Injector) (T, error) {
			d1, err := invokeTyped[D1](i)
			if err != nil {
				var zero T
				return zero, err
			}
			return fn(d1), nil
		})
		return nil
	}, []dix.ServiceRef{dix.TypedService[D1]()}, nil, []dix.ServiceRef{dix.NamedService(name)}, false, false)
}

func OverrideTransient0[T any](fn func() T) dix.SetupFunc {
	return NamedOverrideTransient0(typedName[T](), fn)
}

func NamedOverrideTransient0[T any](name string, fn func() T) dix.SetupFunc {
	return newSetup("OverrideTransient0", func(c *dix.Container) error {
		do.OverrideNamedTransient(c.Raw(), name, func(do.Injector) (T, error) { return fn(), nil })
		return nil
	}, nil, nil, []dix.ServiceRef{dix.NamedService(name)}, false, false)
}

func OverrideTransient1[T, D1 any](fn func(D1) T) dix.SetupFunc {
	return NamedOverrideTransient1(typedName[T](), fn)
}

func NamedOverrideTransient1[T, D1 any](name string, fn func(D1) T) dix.SetupFunc {
	return newSetup("OverrideTransient1", func(c *dix.Container) error {
		do.OverrideNamedTransient(c.Raw(), name, func(i do.Injector) (T, error) {
			d1, err := invokeTyped[D1](i)
			if err != nil {
				var zero T
				return zero, err
			}
			return fn(d1), nil
		})
		return nil
	}, []dix.ServiceRef{dix.TypedService[D1]()}, nil, []dix.ServiceRef{dix.NamedService(name)}, false, false)
}
