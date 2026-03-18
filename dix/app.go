package dix

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	do "github.com/samber/do/v2"
)

// App is the main application container.
type App struct {
	meta      AppMeta
	profile   Profile
	modules   []Module
	container *Container
	lifecycle *lifecycleImpl
	logger    *slog.Logger
	debug     debugSettings
	state     AppState
	built     bool
}

// AppOption configures an App.
type AppOption func(*App)

const DefaultAppName = "dix application"

// NewDefault creates an application with the default framework name.
func NewDefault(opts ...AppOption) *App {
	return New(DefaultAppName, opts...)
}

// New is the preferred constructor in v0.4.
// Everything goes through a single varargs configuration surface.
func New(name string, opts ...AppOption) *App {
	logger := defaultLogger()
	app := &App{
		meta:      AppMeta{Name: name},
		profile:   ProfileDefault,
		modules:   make([]Module, 0),
		container: newContainer(),
		lifecycle: newLifecycle(),
		logger:    logger,
		state:     AppStateCreated,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(app)
		}
	}
	app.syncInfrastructure()
	return app
}

// NewApp keeps backward compatibility with the v0.3 style.
func NewApp(name string, modules ...Module) *App {
	return New(name, WithModules(modules...))
}

// NewAppWithOptions keeps backward compatibility with the v0.3 style.
// Deprecated: prefer New(name, WithModules(...), WithProfile(...), ...).
func NewAppWithOptions(name string, opts []AppOption, modules ...Module) *App {
	merged := make([]AppOption, 0, len(opts)+1)
	merged = append(merged, WithModules(modules...))
	merged = append(merged, opts...)
	return New(name, merged...)
}

func WithProfile(profile Profile) AppOption {
	return func(a *App) { a.profile = profile }
}

// WithVersion sets application version metadata.
func WithVersion(version string) AppOption {
	return func(a *App) { a.meta.Version = version }
}

// WithDescription sets application description metadata.
func WithAppDescription(description string) AppOption {
	return func(a *App) { a.meta.Description = description }
}

// WithLogger sets the framework logger.
func WithLogger(logger *slog.Logger) AppOption {
	return func(a *App) {
		if logger != nil {
			a.logger = logger
		}
	}
}

// WithModules appends application modules.
func WithModules(modules ...Module) AppOption {
	return func(a *App) {
		a.modules = append(a.modules, modules...)
	}
}

// WithModule appends a single application module.
func WithModule(module Module) AppOption {
	return WithModules(module)
}

// WithDebugScopeTree logs do's scope tree after build.
func WithDebugScopeTree(enabled bool) AppOption {
	return func(a *App) { a.debug.scopeTree = enabled }
}

// WithDebugNamedServiceDependencies logs dependency trees for named services after build.
func WithDebugNamedServiceDependencies(names ...string) AppOption {
	return func(a *App) {
		a.debug.namedServiceDependencies = append(a.debug.namedServiceDependencies, names...)
	}
}

func defaultLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
}

func (a *App) syncInfrastructure() {
	a.container.logger = a.logger
	a.lifecycle.logger = a.logger
}

func (a *App) Name() string          { return a.meta.Name }
func (a *App) Profile() Profile      { return a.profile }
func (a *App) Container() *Container { return a.container }
func (a *App) Logger() *slog.Logger  { return a.logger }
func (a *App) State() AppState       { return a.state }
func (a *App) Raw() do.Injector      { return a.container.Raw() }
func (a *App) Meta() AppMeta         { return a.meta }

func (a *App) Run() error {
	if err := a.Build(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	ctx := context.Background()
	if err := a.Start(ctx); err != nil {
		return fmt.Errorf("start failed: %w", err)
	}
	a.waitForShutdown()
	if err := a.Stop(ctx); err != nil {
		return fmt.Errorf("stop failed: %w", err)
	}
	return nil
}

func (a *App) Build() error {
	if a.built {
		a.logger.Debug("build skipped", "app", a.Name(), "reason", "already built")
		return nil
	}
	a.syncInfrastructure()
	a.logger.Info("building app", "app", a.Name(), "profile", a.profile)

	ProvideValueT[*slog.Logger](a.container, a.logger)
	ProvideValueT[AppMeta](a.container, a.meta)
	ProvideValueT[Profile](a.container, a.profile)

	flatModules, err := flattenModules(a.modules, a.profile)
	if err != nil {
		a.logger.Error("module flatten failed", "app", a.Name(), "error", err)
		return fmt.Errorf("module flatten failed: %w", err)
	}

	for _, mod := range flatModules {
		a.logger.Debug("registering module", "module", mod.Name)
		for _, provider := range mod.Providers {
			provider(a.container)
		}
	}

	for _, mod := range flatModules {
		if mod.Setup != nil {
			a.logger.Debug("running module setup", "module", mod.Name)
			if err := mod.Setup(a.container, a.lifecycle); err != nil {
				a.logger.Error("module setup failed", "module", mod.Name, "error", err)
				return fmt.Errorf("setup failed for module %s: %w", mod.Name, err)
			}
		}
		if mod.DoSetup != nil {
			a.logger.Debug("running do setup", "module", mod.Name)
			if err := mod.DoSetup(a.container.Raw()); err != nil {
				a.logger.Error("do setup failed", "module", mod.Name, "error", err)
				return fmt.Errorf("do setup failed for module %s: %w", mod.Name, err)
			}
		}
	}

	if err := a.Validate(); err != nil {
		a.logger.Error("validation failed", "app", a.Name(), "error", err)
		return err
	}

	for _, mod := range flatModules {
		for _, invoke := range mod.Invokes {
			if err := invoke(a.container); err != nil {
				a.logger.Error("invoke failed", "module", mod.Name, "error", err)
				return fmt.Errorf("invoke failed in module %s: %w", mod.Name, err)
			}
		}
	}

	a.built = true
	a.state = AppStateBuilt
	a.logger.Info("app built", "app", a.Name(), "modules", len(flatModules))
	a.logDebugInformation()
	return nil
}

func (a *App) Start(ctx context.Context) error {
	if !a.built {
		return fmt.Errorf("app must be built before starting, call Build() first")
	}
	a.logger.Info("starting app", "app", a.Name())
	if err := a.lifecycle.executeStartHooks(ctx, a.container); err != nil {
		a.logger.Error("app start failed", "app", a.Name(), "error", err)
		return err
	}
	a.state = AppStateStarted
	a.logger.Info("app started", "app", a.Name())
	return nil
}

func (a *App) Stop(ctx context.Context) error {
	if a.state != AppStateStarted {
		return fmt.Errorf("app must be started before stopping")
	}
	a.logger.Info("stopping app", "app", a.Name())
	if err := a.lifecycle.executeStopHooks(ctx, a.container); err != nil {
		a.logger.Error("stop hooks failed", "app", a.Name(), "error", err)
		return err
	}
	if err := a.container.Shutdown(ctx); err != nil {
		a.logger.Error("container shutdown failed", "app", a.Name(), "error", err)
		return fmt.Errorf("container shutdown failed: %w", err)
	}
	a.state = AppStateStopped
	a.logger.Info("app stopped", "app", a.Name())
	return nil
}

func (a *App) waitForShutdown() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

func (a *App) logDebugInformation() {
	if a.debug.scopeTree {
		injector := do.ExplainInjector(a.container.Raw())
		a.logger.Info("do scope tree", "app", a.Name(), "tree", injector.String())
	}
	for _, name := range a.debug.namedServiceDependencies {
		if desc, found := do.ExplainNamedService(a.container.Raw(), name); found {
			a.logger.Info("do named service dependencies", "app", a.Name(), "name", name, "dependencies", desc.String())
		} else {
			a.logger.Warn("do named service not found", "app", a.Name(), "name", name)
		}
	}
}

// LivenessHandler returns an HTTP handler for liveness checks.
func (a *App) LivenessHandler() http.HandlerFunc {
	return a.healthHandler(HealthKindLiveness)
}

// ReadinessHandler returns an HTTP handler for readiness checks.
func (a *App) ReadinessHandler() http.HandlerFunc {
	return a.healthHandler(HealthKindReadiness)
}

// HealthHandler returns an HTTP handler for all checks.
func (a *App) HealthHandler() http.HandlerFunc {
	return a.healthHandler(HealthKindGeneral)
}
