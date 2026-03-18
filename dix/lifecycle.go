package dix

import (
	"context"
	"fmt"
	"log/slog"

	collectionlist "github.com/DaiYuANg/arcgo/collectionx/list"
	"github.com/samber/lo"
)

// StartHook is executed when the application starts.
type StartHook func(ctx context.Context) error

// StopHook is executed when the application stops.
type StopHook func(ctx context.Context) error

// Lifecycle manages application lifecycle hooks.
type Lifecycle interface {
	OnStart(hook StartHook)
	OnStop(hook StopHook)
}

type HookFunc struct {
	register func(*Container, Lifecycle)
	meta     HookMetadata
}

func (h HookFunc) bind(c *Container, lc Lifecycle) {
	if h.register != nil {
		h.register(c, lc)
	}
}

// lifecycleImpl is the internal implementation.
type lifecycleImpl struct {
	startHooks collectionlist.List[StartHook]
	stopHooks  collectionlist.List[StopHook]
	logger     *slog.Logger
}

func newLifecycle() *lifecycleImpl {
	return &lifecycleImpl{
		logger: slog.Default(),
	}
}

func (l *lifecycleImpl) OnStart(hook StartHook) {
	l.startHooks.Add(hook)
}

func (l *lifecycleImpl) OnStop(hook StopHook) {
	l.stopHooks.Add(hook)
}

func (l *lifecycleImpl) executeStartHooks(ctx context.Context, _ *Container) error {
	for i, hook := range l.startHooks.Values() {
		if err := hook(ctx); err != nil {
			if l.logger != nil {
				l.logger.Error("start hook failed", "index", i, "error", err)
			}
			return fmt.Errorf("start hook %d failed: %w", i, err)
		}
	}
	return nil
}

func (l *lifecycleImpl) executeStopHooks(ctx context.Context, _ *Container) error {
	stopHooks := lo.Reverse(l.stopHooks.Values())
	for i, hook := range stopHooks {
		if err := hook(ctx); err != nil {
			if l.logger != nil {
				l.logger.Error("stop hook failed", "index", i, "error", err)
			}
			return fmt.Errorf("stop hook %d failed: %w", i, err)
		}
	}
	return nil
}

func OnStart0(fn func(context.Context) error) HookFunc {
	return NewHookFunc(func(_ *Container, lc Lifecycle) {
		lc.OnStart(fn)
	}, HookMetadata{
		Label: "OnStart0",
		Kind:  HookKindStart,
	})
}

func OnStop0(fn func(context.Context) error) HookFunc {
	return NewHookFunc(func(_ *Container, lc Lifecycle) {
		lc.OnStop(fn)
	}, HookMetadata{
		Label: "OnStop0",
		Kind:  HookKindStop,
	})
}

func OnStart[T any](fn func(context.Context, T) error) HookFunc {
	return NewHookFunc(func(c *Container, lc Lifecycle) {
		lc.OnStart(func(ctx context.Context) error {
			t, err := ResolveAs[T](c)
			if err != nil {
				return fmt.Errorf("resolving dependency: %w", err)
			}
			return fn(ctx, t)
		})
	}, HookMetadata{
		Label:        "OnStart",
		Kind:         HookKindStart,
		Dependencies: []ServiceRef{TypedService[T]()},
	})
}

func OnStop[T any](fn func(context.Context, T) error) HookFunc {
	return NewHookFunc(func(c *Container, lc Lifecycle) {
		lc.OnStop(func(ctx context.Context) error {
			t, err := ResolveAs[T](c)
			if err != nil {
				return fmt.Errorf("resolving dependency: %w", err)
			}
			return fn(ctx, t)
		})
	}, HookMetadata{
		Label:        "OnStop",
		Kind:         HookKindStop,
		Dependencies: []ServiceRef{TypedService[T]()},
	})
}

func OnStart2[T1, T2 any](fn func(context.Context, T1, T2) error) HookFunc {
	return NewHookFunc(func(c *Container, lc Lifecycle) {
		lc.OnStart(func(ctx context.Context) error {
			t1, err := ResolveAs[T1](c)
			if err != nil {
				return err
			}
			t2, err := ResolveAs[T2](c)
			if err != nil {
				return err
			}
			return fn(ctx, t1, t2)
		})
	}, HookMetadata{
		Label:        "OnStart2",
		Kind:         HookKindStart,
		Dependencies: []ServiceRef{TypedService[T1](), TypedService[T2]()},
	})
}

func OnStop2[T1, T2 any](fn func(context.Context, T1, T2) error) HookFunc {
	return NewHookFunc(func(c *Container, lc Lifecycle) {
		lc.OnStop(func(ctx context.Context) error {
			t1, err := ResolveAs[T1](c)
			if err != nil {
				return err
			}
			t2, err := ResolveAs[T2](c)
			if err != nil {
				return err
			}
			return fn(ctx, t1, t2)
		})
	}, HookMetadata{
		Label:        "OnStop2",
		Kind:         HookKindStop,
		Dependencies: []ServiceRef{TypedService[T1](), TypedService[T2]()},
	})
}

func OnStart3[T1, T2, T3 any](fn func(context.Context, T1, T2, T3) error) HookFunc {
	return NewHookFunc(func(c *Container, lc Lifecycle) {
		lc.OnStart(func(ctx context.Context) error {
			t1, err := ResolveAs[T1](c)
			if err != nil {
				return err
			}
			t2, err := ResolveAs[T2](c)
			if err != nil {
				return err
			}
			t3, err := ResolveAs[T3](c)
			if err != nil {
				return err
			}
			return fn(ctx, t1, t2, t3)
		})
	}, HookMetadata{
		Label:        "OnStart3",
		Kind:         HookKindStart,
		Dependencies: []ServiceRef{TypedService[T1](), TypedService[T2](), TypedService[T3]()},
	})
}

func OnStop3[T1, T2, T3 any](fn func(context.Context, T1, T2, T3) error) HookFunc {
	return NewHookFunc(func(c *Container, lc Lifecycle) {
		lc.OnStop(func(ctx context.Context) error {
			t1, err := ResolveAs[T1](c)
			if err != nil {
				return err
			}
			t2, err := ResolveAs[T2](c)
			if err != nil {
				return err
			}
			t3, err := ResolveAs[T3](c)
			if err != nil {
				return err
			}
			return fn(ctx, t1, t2, t3)
		})
	}, HookMetadata{
		Label:        "OnStop3",
		Kind:         HookKindStop,
		Dependencies: []ServiceRef{TypedService[T1](), TypedService[T2](), TypedService[T3]()},
	})
}
