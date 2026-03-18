package dix_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DaiYuANg/arcgo/dix"
	dixadvanced "github.com/DaiYuANg/arcgo/dix/advanced"
	do "github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example providers

type Config struct {
	DSN  string
	Port int
}

func ProvideConfig() Config {
	return Config{
		DSN:  "sqlite://test.db",
		Port: 8080,
	}
}

type Database struct {
	dsn string
}

func NewDatabase(dsn string) *Database {
	return &Database{dsn: dsn}
}

func (d *Database) Connect() error {
	fmt.Println("Database connected:", d.dsn)
	return nil
}

type testGreeter interface {
	Greet() string
}

type testGreeterImpl struct{}

func (g *testGreeterImpl) Greet() string {
	return "hello"
}

func (d *Database) Close() error {
	fmt.Println("Database closed")
	return nil
}

func ProvideDatabase(cfg Config) *Database {
	return NewDatabase(cfg.DSN)
}

type Server struct {
	addr string
}

func ProvideServer(cfg Config) *Server {
	return &Server{
		addr: fmt.Sprintf(":%d", cfg.Port),
	}
}

func (s *Server) Start() error {
	fmt.Println("Server starting on", s.addr)
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	fmt.Println("Server stopped")
	return nil
}

// Example modules

var DatabaseModule = dix.NewModule("database",
	dix.WithModuleProviders(
		dix.Provider0(ProvideConfig),
		dix.Provider1(ProvideDatabase),
	),
	dix.WithModuleHooks(
		dix.OnStart(func(ctx context.Context, db *Database) error {
			return db.Connect()
		}),
		dix.OnStop(func(ctx context.Context, db *Database) error {
			return db.Close()
		}),
	),
)

var ServerModule = dix.NewModule("server",
	dix.WithModuleProviders(
		dix.Provider1(ProvideServer),
	),
	dix.WithModuleImports(DatabaseModule),
	dix.WithModuleHooks(
		dix.OnStart(func(ctx context.Context, s *Server) error {
			return s.Start()
		}),
		dix.OnStop(func(ctx context.Context, s *Server) error {
			return s.Stop(ctx)
		}),
	),
)

func buildRuntime(t *testing.T, app *dix.App) *dix.Runtime {
	t.Helper()
	rt, err := app.Build()
	require.NoError(t, err)
	require.NotNil(t, rt)
	return rt
}

// Tests

func TestApp_Build(t *testing.T) {
	app := dix.NewApp("testapp", ServerModule)
	rt := buildRuntime(t, app)

	db, err := dix.ResolveAs[*Database](rt.Container())
	require.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, "sqlite://test.db", db.dsn)

	cfg, err := dix.ResolveAs[Config](rt.Container())
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, dix.AppStateBuilt, rt.State())
}

func TestApp_BuildCreatesIndependentRuntimes(t *testing.T) {
	app := dix.NewApp("testapp", ServerModule)

	first := buildRuntime(t, app)
	second := buildRuntime(t, app)

	assert.NotSame(t, first, second)
	assert.NotSame(t, first.Container(), second.Container())
	assert.Equal(t, first.Name(), second.Name())
}

func TestRuntime_StartStop(t *testing.T) {
	rt := buildRuntime(t, dix.NewApp("testapp", ServerModule))
	ctx := context.Background()

	require.Equal(t, dix.AppStateBuilt, rt.State())
	require.NoError(t, rt.Start(ctx))
	assert.Equal(t, dix.AppStateStarted, rt.State())

	require.NoError(t, rt.Stop(ctx))
	assert.Equal(t, dix.AppStateStopped, rt.State())
}

func TestModule_WithProfiles(t *testing.T) {
	devOnlyCalled := false
	prodOnlyCalled := false

	devModule := dix.NewModule("dev-only",
		dix.WithModuleProfiles(dix.ProfileDev),
		dix.WithModuleProviders(
			dix.Provider0(func() string {
				devOnlyCalled = true
				return "dev"
			}),
		),
	)

	prodModule := dix.NewModule("prod-only",
		dix.WithModuleProfiles(dix.ProfileProd),
		dix.WithModuleProviders(
			dix.Provider0(func() string {
				prodOnlyCalled = true
				return "prod"
			}),
		),
	)

	appDev := dix.New("test",
		dix.WithProfile(dix.ProfileDev),
		dix.WithModules(devModule, prodModule),
	)

	rtDev := buildRuntime(t, appDev)
	devStr, err := dix.ResolveAs[string](rtDev.Container())
	require.NoError(t, err)
	assert.Equal(t, "dev", devStr)
	assert.True(t, devOnlyCalled)
	assert.False(t, prodOnlyCalled)

	appProd := dix.New("test",
		dix.WithProfile(dix.ProfileProd),
		dix.WithModules(devModule, prodModule),
	)

	rtProd := buildRuntime(t, appProd)
	prodStr, err := dix.ResolveAs[string](rtProd.Container())
	require.NoError(t, err)
	assert.Equal(t, "prod", prodStr)
}

func TestModule_WithExcludeProfiles(t *testing.T) {
	called := false

	module := dix.NewModule("not-test",
		dix.WithModuleExcludeProfiles(dix.ProfileTest),
		dix.WithModuleProviders(
			dix.Provider0(func() string {
				called = true
				return "value"
			}),
		),
	)

	appTest := dix.New("test",
		dix.WithProfile(dix.ProfileTest),
		dix.WithModule(module),
	)

	rt := buildRuntime(t, appTest)
	_, err := dix.ResolveAs[string](rt.Container())
	assert.Error(t, err)
	assert.False(t, called)
}

func TestResolveAs(t *testing.T) {
	rt := buildRuntime(t, dix.NewApp("test",
		dix.NewModule("test",
			dix.WithModuleProviders(
				dix.Provider0(ProvideConfig),
			),
		),
	))

	cfg, err := dix.ResolveAs[Config](rt.Container())
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Port)
}

func TestMustResolveAs(t *testing.T) {
	rt := buildRuntime(t, dix.NewApp("test",
		dix.NewModule("test",
			dix.WithModuleProviders(
				dix.Provider0(ProvideConfig),
			),
		),
	))

	cfg := dix.MustResolveAs[Config](rt.Container())
	assert.Equal(t, 8080, cfg.Port)
}

func TestInvoke(t *testing.T) {
	invoked := false

	rt := buildRuntime(t, dix.NewApp("test",
		dix.NewModule("test",
			dix.WithModuleProviders(
				dix.Provider0(ProvideConfig),
			),
			dix.WithModuleInvokes(
				dix.Invoke1(func(cfg Config) {
					invoked = true
					assert.Equal(t, 8080, cfg.Port)
				}),
			),
		),
	))

	assert.True(t, invoked)
	assert.NotNil(t, rt)
}

func TestModule_Imports(t *testing.T) {
	baseModule := dix.NewModule("base",
		dix.WithModuleProviders(
			dix.Provider0(ProvideConfig),
		),
	)

	dependentModule := dix.NewModule("dependent",
		dix.WithModuleImports(baseModule),
		dix.WithModuleProviders(
			dix.Provider1(ProvideDatabase),
		),
	)

	rt := buildRuntime(t, dix.NewApp("test", dependentModule))

	cfg, err := dix.ResolveAs[Config](rt.Container())
	require.NoError(t, err)

	db, err := dix.ResolveAs[*Database](rt.Container())
	require.NoError(t, err)

	assert.NotNil(t, cfg)
	assert.NotNil(t, db)
}

func TestModule_ImportDeduplicatesSharedDependency(t *testing.T) {
	called := 0

	shared := dix.NewModule("shared",
		dix.WithModuleProviders(
			dix.Provider0(func() string {
				called++
				return "shared"
			}),
		),
	)

	left := dix.NewModule("left", dix.WithModuleImports(shared))
	right := dix.NewModule("right", dix.WithModuleImports(shared))

	rt := buildRuntime(t, dix.NewApp("test", left, right))
	value, err := dix.ResolveAs[string](rt.Container())
	require.NoError(t, err)
	assert.Equal(t, "shared", value)
	assert.Equal(t, 1, called)
}

func TestApp_ValidateGraph(t *testing.T) {
	app := dix.NewApp("test", DatabaseModule, ServerModule)
	require.NoError(t, app.Validate())
}

func TestApp_ValidateDetectsMissingDependency(t *testing.T) {
	app := dix.NewApp("validate-missing",
		dix.NewModule("broken",
			dix.WithModuleProviders(
				dix.Provider1(func(cfg Config) *Database {
					return &Database{dsn: cfg.DSN}
				}),
			),
		),
	)

	err := app.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), do.NameOf[Config]())
}

func TestApp_ValidateDetectsDuplicateProvider(t *testing.T) {
	app := dix.NewApp("validate-duplicate",
		dix.NewModule("dup",
			dix.WithModuleProviders(
				dix.Provider0(func() Config { return ProvideConfig() }),
				dix.Provider0(func() Config { return ProvideConfig() }),
			),
		),
	)

	err := app.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate provider output")
	assert.Contains(t, err.Error(), do.NameOf[Config]())
}

func TestApp_ValidateDoesNotEscapeForCoreSetup(t *testing.T) {
	app := dix.NewApp("validate-setup",
		dix.NewModule("health",
			dix.WithModuleSetup(func(c *dix.Container, _ dix.Lifecycle) error {
				c.RegisterHealthCheck("ok", func(ctx context.Context) error { return nil })
				return nil
			}),
			dix.WithModuleProviders(
				dix.Provider1(func(cfg Config) *Database {
					return &Database{dsn: cfg.DSN}
				}),
			),
		),
	)

	err := app.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), do.NameOf[Config]())
}

func TestApp_ValidateAdvancedAliasDependency(t *testing.T) {
	app := dix.NewApp("validate-alias",
		dix.NewModule("alias",
			dix.WithModuleProviders(
				dix.Provider0(func() *testGreeterImpl { return &testGreeterImpl{} }),
			),
			dix.WithModuleSetups(
				dixadvanced.BindAlias[*testGreeterImpl, testGreeter](),
			),
			dix.WithModuleInvokes(
				dix.Invoke1(func(testGreeter) {}),
			),
		),
	)

	require.NoError(t, app.Validate())
}

func TestProvider4AndAggregateDependencyStruct(t *testing.T) {
	type Params struct {
		Config Config
		DB     *Database
		Server *Server
		Label  string
	}

	module := dix.NewModule("test",
		dix.WithModuleProviders(
			dix.Provider0(ProvideConfig),
			dix.Provider1(ProvideDatabase),
			dix.Provider1(ProvideServer),
			dix.Provider0(func() string { return "ok" }),
			dix.Provider4(func(cfg Config, db *Database, srv *Server, label string) Params {
				return Params{Config: cfg, DB: db, Server: srv, Label: label}
			}),
		),
	)

	rt := buildRuntime(t, dix.NewApp("test", module))
	params, err := dix.ResolveAs[Params](rt.Container())
	require.NoError(t, err)
	assert.Equal(t, 8080, params.Config.Port)
	assert.Equal(t, "sqlite://test.db", params.DB.dsn)
	assert.Equal(t, ":8080", params.Server.addr)
	assert.Equal(t, "ok", params.Label)
}

func TestLifecycleHookReceivesContext(t *testing.T) {
	type ctxKey string

	const key ctxKey = "trace"
	received := ""

	module := dix.NewModule("ctx",
		dix.WithModuleProviders(dix.Provider0(func() string { return "value" })),
		dix.WithModuleHooks(
			dix.OnStart(func(ctx context.Context, value string) error {
				received = ctx.Value(key).(string) + ":" + value
				return nil
			}),
		),
	)

	rt := buildRuntime(t, dix.NewApp("test", module))
	require.NoError(t, rt.Start(context.WithValue(context.Background(), key, "abc")))
	require.NoError(t, rt.Stop(context.Background()))
	assert.Equal(t, "abc:value", received)
}

func TestContainerRegisterProviderDefinitionReturnsError(t *testing.T) {
	rt := buildRuntime(t, dix.NewApp("test"))
	err := rt.Container().Register(dix.Definition{Kind: dix.DefinitionProvider})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

// Examples

func ExampleApp() {
	app := dix.NewApp("myapp",
		DatabaseModule,
		ServerModule,
	)
	_ = app
}

func ExampleNewModule() {
	module := dix.NewModule("example",
		dix.WithModuleProviders(
			dix.Provider0(func() string { return "hello" }),
			dix.Provider1(func(s string) int { return len(s) }),
		),
		dix.WithModuleSetup(func(c *dix.Container, lc dix.Lifecycle) error {
			lc.OnStart(func(ctx context.Context) error {
				s, _ := dix.ResolveAs[string](c)
				fmt.Println("Starting with:", s)
				return nil
			})
			return nil
		}),
	)

	app := dix.NewApp("test", module)
	_, _ = app.Build()
}

func ExampleWithModuleHooks() {
	module := dix.NewModule("example",
		dix.WithModuleProviders(
			dix.Provider0(func() *http.Server {
				return &http.Server{Addr: ":8080"}
			}),
		),
		dix.WithModuleHooks(
			dix.OnStart(func(ctx context.Context, s *http.Server) error {
				go s.ListenAndServe()
				return nil
			}),
			dix.OnStop(func(ctx context.Context, s *http.Server) error {
				return s.Shutdown(ctx)
			}),
		),
	)

	app := dix.NewApp("test", module)
	_ = app
}

func TestResolveOptionalAs(t *testing.T) {
	rt := buildRuntime(t, dix.NewApp("test"))
	_, ok := dix.ResolveOptionalAs[string](rt.Container())
	assert.False(t, ok)
}

func TestResolveOrElse(t *testing.T) {
	rt := buildRuntime(t, dix.NewApp("test",
		dix.NewModule("deps",
			dix.WithModuleProviders(
				dix.Provider0(func() string { return "configured" }),
			),
		),
	))

	assert.Equal(t, "configured", dix.ResolveOrElse[string](rt.Container(), "fallback"))
	assert.Equal(t, 42, dix.ResolveOrElse[int](rt.Container(), 42))
}

func TestProfileFromEnv(t *testing.T) {
	t.Setenv("ARCGO_DIX_PROFILE", string(dix.ProfileDev))
	assert.Equal(t, dix.ProfileDev, dix.ProfileFromEnv("ARCGO_DIX_PROFILE", dix.ProfileProd))

	t.Setenv("ARCGO_DIX_PROFILE", "invalid")
	assert.Equal(t, dix.ProfileProd, dix.ProfileFromEnv("ARCGO_DIX_PROFILE", dix.ProfileProd))
}

func TestWithDoSetup(t *testing.T) {
	called := false
	module := dix.NewModule("advanced",
		dix.WithModuleSetups(
			dixadvanced.DoSetup(func(raw do.Injector) error {
				called = raw != nil
				return nil
			}),
		),
	)
	buildRuntime(t, dix.NewApp("test", module))
	assert.True(t, called)
}

func TestHealthCheckReport(t *testing.T) {
	module := dix.NewModule("health",
		dix.WithModuleSetup(func(c *dix.Container, lc dix.Lifecycle) error {
			c.RegisterHealthCheck("db", func(ctx context.Context) error { return nil })
			c.RegisterHealthCheck("cache", func(ctx context.Context) error { return fmt.Errorf("down") })
			return nil
		}),
	)

	rt := buildRuntime(t, dix.NewApp("test", module))
	report := rt.CheckHealth(context.Background())
	assert.False(t, report.Healthy())
	require.Error(t, report.Error())
	assert.Contains(t, report.Error().Error(), "cache")
}

func TestRuntime_HealthHandlers(t *testing.T) {
	module := dix.NewModule("health",
		dix.WithModuleSetup(func(c *dix.Container, lc dix.Lifecycle) error {
			c.RegisterLivenessCheck("live", func(ctx context.Context) error { return nil })
			c.RegisterReadinessCheck("ready", func(ctx context.Context) error { return fmt.Errorf("booting") })
			return nil
		}),
	)

	rt := buildRuntime(t, dix.NewApp("health", module))

	liveReq := httptest.NewRequest(http.MethodGet, "/livez", nil)
	liveRes := httptest.NewRecorder()
	rt.LivenessHandler()(liveRes, liveReq)
	assert.Equal(t, http.StatusOK, liveRes.Code)

	readyReq := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	readyRes := httptest.NewRecorder()
	rt.ReadinessHandler()(readyRes, readyReq)
	assert.Equal(t, http.StatusServiceUnavailable, readyRes.Code)
}

func TestNew_WithModulesOption(t *testing.T) {
	rt := buildRuntime(t, dix.New("test",
		dix.WithProfile(dix.ProfileDev),
		dix.WithModule(DatabaseModule),
	))

	logger, err := dix.ResolveAs[*slog.Logger](rt.Container())
	require.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestHealthKinds(t *testing.T) {
	mod := dix.NewModule("health",
		dix.WithModuleSetup(func(c *dix.Container, lc dix.Lifecycle) error {
			c.RegisterLivenessCheck("live", func(ctx context.Context) error { return nil })
			c.RegisterReadinessCheck("ready", func(ctx context.Context) error { return nil })
			return nil
		}),
	)

	rt := buildRuntime(t, dix.New("health-app", dix.WithModule(mod)))
	live := rt.CheckLiveness(context.Background())
	ready := rt.CheckReadiness(context.Background())

	assert.True(t, live.Healthy())
	assert.True(t, ready.Healthy())
	assert.Len(t, live.Checks, 1)
	assert.Len(t, ready.Checks, 1)
}

func TestNewDefault(t *testing.T) {
	app := dix.NewDefault()
	assert.Equal(t, dix.DefaultAppName, app.Name())
}
