package dix

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	collectionlist "github.com/DaiYuANg/arcgo/collectionx/list"
	"github.com/samber/lo"
)

// AppOption configures an App specification during construction.
type AppOption func(*appSpec)

const DefaultAppName = "dix application"

// NewDefault creates an application with the default framework name.
func NewDefault(opts ...AppOption) *App {
	return New(DefaultAppName, opts...)
}

// New creates an immutable application specification.
func New(name string, opts ...AppOption) *App {
	spec := &appSpec{
		meta:    AppMeta{Name: name},
		profile: ProfileDefault,
		logger:  defaultLogger(),
	}

	lo.ForEach(lo.Filter(opts, func(opt AppOption, _ int) bool {
		return opt != nil
	}), func(opt AppOption, _ int) {
		opt(spec)
	})

	return &App{spec: spec}
}

// NewApp keeps backward compatibility with the v0.3 style constructor surface.
func NewApp(name string, modules ...Module) *App {
	return New(name, WithModules(modules...))
}

// NewAppWithOptions keeps backward compatibility with the v0.3 style.
// Deprecated: prefer New(name, WithModules(...), WithProfile(...), ...).
func NewAppWithOptions(name string, opts []AppOption, modules ...Module) *App {
	merged := collectionlist.NewListWithCapacity[AppOption](len(opts)+1, WithModules(modules...))
	merged.MergeSlice(opts)
	return New(name, merged.Values()...)
}

func WithProfile(profile Profile) AppOption {
	return func(spec *appSpec) { spec.profile = profile }
}

// WithVersion sets application version metadata.
func WithVersion(version string) AppOption {
	return func(spec *appSpec) { spec.meta.Version = version }
}

// WithDescription sets application description metadata.
func WithAppDescription(description string) AppOption {
	return func(spec *appSpec) { spec.meta.Description = description }
}

// WithLogger sets the framework logger.
func WithLogger(logger *slog.Logger) AppOption {
	return func(spec *appSpec) {
		if logger != nil {
			spec.logger = logger
		}
	}
}

// WithModules appends application modules.
func WithModules(modules ...Module) AppOption {
	return func(spec *appSpec) {
		spec.modules.Add(modules...)
	}
}

// WithModule appends a single application module.
func WithModule(module Module) AppOption {
	return WithModules(module)
}

// WithDebugScopeTree logs do's scope tree after build.
func WithDebugScopeTree(enabled bool) AppOption {
	return func(spec *appSpec) { spec.debug.scopeTree = enabled }
}

// WithDebugNamedServiceDependencies logs dependency trees for named services after build.
func WithDebugNamedServiceDependencies(names ...string) AppOption {
	return func(spec *appSpec) {
		spec.debug.namedServiceDependencies.Add(names...)
	}
}

func defaultLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
}

func (a *App) Name() string {
	if a == nil || a.spec == nil {
		return ""
	}
	return a.spec.meta.Name
}

func (a *App) Profile() Profile {
	if a == nil || a.spec == nil {
		return ""
	}
	return a.spec.profile
}

func (a *App) Logger() *slog.Logger {
	if a == nil || a.spec == nil {
		return nil
	}
	return a.spec.logger
}

func (a *App) Meta() AppMeta {
	if a == nil || a.spec == nil {
		return AppMeta{}
	}
	return a.spec.meta
}

func (a *App) Modules() []Module {
	if a == nil || a.spec == nil {
		return nil
	}
	return a.spec.modules.Values()
}

// Build compiles the immutable App spec into a Runtime.
func (a *App) Build() (*Runtime, error) {
	plan, err := newBuildPlan(a)
	if err != nil {
		return nil, err
	}
	return plan.Build()
}

// Run builds a Runtime, starts it, waits for shutdown signals, and stops it.
func (a *App) Run() error {
	rt, err := a.Build()
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	ctx := context.Background()
	if err := rt.Start(ctx); err != nil {
		return fmt.Errorf("start failed: %w", err)
	}

	waitForShutdown()

	if err := rt.Stop(ctx); err != nil {
		return fmt.Errorf("stop failed: %w", err)
	}

	return nil
}

func waitForShutdown() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
