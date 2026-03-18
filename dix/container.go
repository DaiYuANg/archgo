package dix

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	do "github.com/samber/do/v2"
	mo "github.com/samber/mo"
)

// Container wraps samber/do.Injector.
// Most code should stay on the typed dix helpers.
// Raw() exists as an explicit escape hatch for advanced integrations.
type Container struct {
	injector     do.Injector
	healthChecks []healthCheckEntry
	logger       *slog.Logger
}

func newContainer() *Container {
	return &Container{
		injector:     do.New(),
		healthChecks: make([]healthCheckEntry, 0),
		logger:       slog.Default(),
	}
}

// Raw returns the underlying do injector for advanced integrations.
func (c *Container) Raw() do.Injector { return c.injector }

// Injector returns the underlying do injector.
// Deprecated: prefer Raw() to make advanced usage explicit at call sites.
func (c *Container) Injector() do.Injector { return c.injector }

func (c *Container) Shutdown(ctx context.Context) error {
	return c.injector.ShutdownWithContext(ctx)
}

func ProvideT[T any](c *Container, fn func() T) {
	do.Provide(c.injector, func(i do.Injector) (T, error) { return fn(), nil })
}
func Provide1T[T, D1 any](c *Container, fn func(D1) T) {
	do.Provide(c.injector, func(i do.Injector) (T, error) {
		d1, err := do.Invoke[D1](i)
		if err != nil {
			var zero T
			return zero, err
		}
		return fn(d1), nil
	})
}
func Provide2T[T, D1, D2 any](c *Container, fn func(D1, D2) T) {
	do.Provide(c.injector, func(i do.Injector) (T, error) {
		d1, err := do.Invoke[D1](i)
		if err != nil {
			var zero T
			return zero, err
		}
		d2, err := do.Invoke[D2](i)
		if err != nil {
			var zero T
			return zero, err
		}
		return fn(d1, d2), nil
	})
}
func Provide3T[T, D1, D2, D3 any](c *Container, fn func(D1, D2, D3) T) {
	do.Provide(c.injector, func(i do.Injector) (T, error) {
		d1, err := do.Invoke[D1](i)
		if err != nil {
			var zero T
			return zero, err
		}
		d2, err := do.Invoke[D2](i)
		if err != nil {
			var zero T
			return zero, err
		}
		d3, err := do.Invoke[D3](i)
		if err != nil {
			var zero T
			return zero, err
		}
		return fn(d1, d2, d3), nil
	})
}
func Provide4T[T, D1, D2, D3, D4 any](c *Container, fn func(D1, D2, D3, D4) T) {
	do.Provide(c.injector, func(i do.Injector) (T, error) {
		d1, err := do.Invoke[D1](i)
		if err != nil {
			var zero T
			return zero, err
		}
		d2, err := do.Invoke[D2](i)
		if err != nil {
			var zero T
			return zero, err
		}
		d3, err := do.Invoke[D3](i)
		if err != nil {
			var zero T
			return zero, err
		}
		d4, err := do.Invoke[D4](i)
		if err != nil {
			var zero T
			return zero, err
		}
		return fn(d1, d2, d3, d4), nil
	})
}
func Provide5T[T, D1, D2, D3, D4, D5 any](c *Container, fn func(D1, D2, D3, D4, D5) T) {
	do.Provide(c.injector, func(i do.Injector) (T, error) {
		d1, err := do.Invoke[D1](i)
		if err != nil {
			var zero T
			return zero, err
		}
		d2, err := do.Invoke[D2](i)
		if err != nil {
			var zero T
			return zero, err
		}
		d3, err := do.Invoke[D3](i)
		if err != nil {
			var zero T
			return zero, err
		}
		d4, err := do.Invoke[D4](i)
		if err != nil {
			var zero T
			return zero, err
		}
		d5, err := do.Invoke[D5](i)
		if err != nil {
			var zero T
			return zero, err
		}
		return fn(d1, d2, d3, d4, d5), nil
	})
}
func Provide6T[T, D1, D2, D3, D4, D5, D6 any](c *Container, fn func(D1, D2, D3, D4, D5, D6) T) {
	do.Provide(c.injector, func(i do.Injector) (T, error) {
		d1, err := do.Invoke[D1](i)
		if err != nil {
			var zero T
			return zero, err
		}
		d2, err := do.Invoke[D2](i)
		if err != nil {
			var zero T
			return zero, err
		}
		d3, err := do.Invoke[D3](i)
		if err != nil {
			var zero T
			return zero, err
		}
		d4, err := do.Invoke[D4](i)
		if err != nil {
			var zero T
			return zero, err
		}
		d5, err := do.Invoke[D5](i)
		if err != nil {
			var zero T
			return zero, err
		}
		d6, err := do.Invoke[D6](i)
		if err != nil {
			var zero T
			return zero, err
		}
		return fn(d1, d2, d3, d4, d5, d6), nil
	})
}

func ProvideValueT[T any](c *Container, value T) { do.ProvideValue(c.injector, value) }
func ResolveAs[T any](c *Container) (T, error)   { return do.Invoke[T](c.injector) }

func ResolveOptionalAs[T any](c *Container) (value T, ok bool) {
	option := ResolveOptionAs[T](c)
	return option.Get()
}

// ResolveOptionAs resolves an optional dependency as mo.Option.
func ResolveOptionAs[T any](c *Container) mo.Option[T] {
	value, err := ResolveAs[T](c)
	if err != nil {
		return mo.None[T]()
	}
	return mo.Some(value)
}
func ResolveOrElse[T any](c *Container, fallback T) T {
	value, ok := ResolveOptionalAs[T](c)
	if !ok {
		return fallback
	}
	return value
}
func MustResolveAs[T any](c *Container) T {
	result, err := ResolveAs[T](c)
	if err != nil {
		panic(fmt.Sprintf("failed to resolve dependency: %v", err))
	}
	return result
}

// Backward compatibility types.
type Definition struct {
	Name       string
	Kind       DefinitionKind
	Value      any
	Provider   any
	ModuleName string
	Lazy       bool
	Transient  bool
}

type DefinitionKind string

const (
	DefinitionValue    DefinitionKind = "value"
	DefinitionProvider DefinitionKind = "provider"
)

func (c *Container) Register(def Definition) error {
	switch def.Kind {
	case DefinitionValue:
		if def.Name != "" {
			do.ProvideNamedValue(c.injector, def.Name, def.Value)
		} else {
			do.ProvideValue(c.injector, def.Value)
		}
		return nil
	case DefinitionProvider:
		return errors.New("provider definition registration is not implemented; use typed ProviderN helpers instead")
	default:
		return fmt.Errorf("unknown definition kind: %v", def.Kind)
	}
}

func (c *Container) Resolve(target any) error {
	return errors.New("Resolve(target) is not supported; use ResolveAs[T]() for type-safe resolution")
}
