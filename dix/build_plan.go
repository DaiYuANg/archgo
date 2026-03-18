package dix

import (
	"context"
	"fmt"
	"log/slog"

	collectionlist "github.com/DaiYuANg/arcgo/collectionx/list"
)

type buildPlan struct {
	spec    *appSpec
	modules *collectionlist.List[*moduleSpec]
}

func newBuildPlan(app *App) (*buildPlan, error) {
	if app == nil || app.spec == nil {
		return nil, fmt.Errorf("app is nil")
	}

	modules, err := flattenModuleList(&app.spec.modules, app.spec.profile)
	if err != nil {
		logger := app.spec.logger
		if logger != nil {
			logger.Error("module flatten failed", "app", app.Name(), "error", err)
		}
		return nil, fmt.Errorf("module flatten failed: %w", err)
	}

	plan := &buildPlan{
		spec:    app.spec,
		modules: modules,
	}

	return plan, validateTypedGraph(plan)
}

func (p *buildPlan) Build() (*Runtime, error) {
	if p == nil || p.spec == nil {
		return nil, fmt.Errorf("build plan is nil")
	}

	logger := p.spec.logger
	if logger == nil {
		logger = slog.Default()
	}
	debugEnabled := logger.Enabled(context.Background(), slog.LevelDebug)
	infoEnabled := logger.Enabled(context.Background(), slog.LevelInfo)

	if infoEnabled {
		logger.Info("building app", "app", p.spec.meta.Name, "profile", p.spec.profile)
	}

	rt := newRuntime(p.spec, p)
	ProvideValueT[*slog.Logger](rt.container, rt.logger)
	ProvideValueT[AppMeta](rt.container, p.spec.meta)
	ProvideValueT[Profile](rt.container, p.spec.profile)

	p.modules.Range(func(_ int, mod *moduleSpec) bool {
		if debugEnabled {
			logger.Debug("registering module", "module", mod.name)
		}
		mod.providers.Range(func(_ int, provider ProviderFunc) bool {
			provider.apply(rt.container)
			return true
		})
		return true
	})

	var setupErr error
	p.modules.Range(func(_ int, mod *moduleSpec) bool {
		mod.hooks.Range(func(_ int, hook HookFunc) bool {
			hook.bind(rt.container, rt.lifecycle)
			return true
		})

		mod.setups.Range(func(_ int, setup SetupFunc) bool {
			if debugEnabled {
				logger.Debug("running module setup", "module", mod.name, "label", setup.meta.Label)
			}
			if err := setup.apply(rt.container, rt.lifecycle); err != nil {
				logger.Error("module setup failed", "module", mod.name, "label", setup.meta.Label, "error", err)
				setupErr = fmt.Errorf("setup failed for module %s via %s: %w", mod.name, setup.meta.Label, err)
				return false
			}
			return true
		})
		if setupErr != nil {
			return false
		}
		return true
	})
	if setupErr != nil {
		return nil, setupErr
	}

	var buildErr error
	p.modules.Range(func(_ int, mod *moduleSpec) bool {
		var invokeErr error
		mod.invokes.Range(func(_ int, invoke InvokeFunc) bool {
			invokeErr = invoke.apply(rt.container)
			return invokeErr == nil
		})
		if invokeErr != nil {
			logger.Error("invoke failed", "module", mod.name, "error", invokeErr)
			buildErr = fmt.Errorf("invoke failed in module %s: %w", mod.name, invokeErr)
			return false
		}
		return true
	})
	if buildErr != nil {
		return nil, buildErr
	}

	rt.state = AppStateBuilt
	if infoEnabled {
		logger.Info("app built", "app", rt.Name(), "modules", p.modules.Len())
	}
	rt.logDebugInformation()
	return rt, nil
}
