package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	authxfx "github.com/DaiYuANg/arcgo/authx/fx"
	"github.com/DaiYuANg/arcgo/eventx"
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/httpx/adapter"
	httpxfiber "github.com/DaiYuANg/arcgo/httpx/adapter/fiber"
	httpxfx "github.com/DaiYuANg/arcgo/httpx/fx"
	"github.com/DaiYuANg/arcgo/logx"
	logxfx "github.com/DaiYuANg/arcgo/logx/fx"
	"github.com/DaiYuANg/arcgo/observabilityx"
	promobs "github.com/DaiYuANg/arcgo/observabilityx/prometheus"
	"github.com/gofiber/fiber/v2"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/fx"
)

func newAppModule() fx.Option {
	return fx.Options(
		logxfx.NewLogxModule(
			logx.WithConsole(true),
			logx.WithInfoLevel(),
		),
		authxfx.NewAuthxModule(),
		httpxfx.NewHttpxModule(),
		fx.Provide(
			newConfig,
			newPrometheusAdapter,
			newObservability,
			newJWTService,
			newEventBus,
			newStore,
			fx.Annotate(newAuthxEngineOptions, fx.ResultTags(`group:"authx_engine_options,flatten"`)),
			newGuard,
			newAuthMiddleware,
			newFiberAdapter,
			fx.Annotate(newHTTPServerOptions, fx.ResultTags(`group:"httpx_server_options,flatten"`)),
		),
		fx.Invoke(
			registerEventSubscribers,
			registerHTTPRoutes,
			registerInfraRoutes,
			startHTTPServer,
		),
	)
}

func newPrometheusAdapter(logger *slog.Logger) *promobs.Adapter {
	return promobs.New(
		promobs.WithLogger(logger),
		promobs.WithNamespace("arcgo_rbac_example"),
	)
}

func newObservability(logger *slog.Logger, prom *promobs.Adapter) observabilityx.Observability {
	return observabilityx.Multi(observabilityx.NopWithLogger(logger), prom)
}

func newEventBus(
	lc fx.Lifecycle,
	cfg appConfig,
	obs observabilityx.Observability,
	logger *slog.Logger,
) eventx.BusRuntime {
	bus := eventx.New(
		eventx.WithObservability(obs),
		eventx.WithAntsPool(cfg.Event.Workers),
		eventx.WithParallelDispatch(cfg.Event.Parallel),
		eventx.WithAsyncErrorHandler(func(ctx context.Context, event eventx.Event, err error) {
			if err == nil || event == nil {
				return
			}
			logx.WithError(logx.WithFields(logger, map[string]any{
				"event": event.Name(),
			}), err).Error("async event dispatch failed")
		}),
	)

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return bus.Close()
		},
	})
	return bus
}

func newFiberAdapter(
	cfg appConfig,
	logger *slog.Logger,
	obs observabilityx.Observability,
	authMW fiber.Handler,
) *httpxfiber.Adapter {
	fiberAdapter := httpxfiber.NewWithOptions(nil, httpxfiber.Options{
		Logger: logger,
		Huma: adapter.HumaOptions{
			Title:       "ArcGo RBAC Backend Scaffold",
			Version:     cfg.Version,
			Description: "httpx(fiber) + authx(jwt) + eventx + observabilityx + bun + fx",
			DocsPath:    cfg.docsPath(),
			OpenAPIPath: cfg.openAPIPath(),
		},
	})

	router := fiberAdapter.Router()
	router.Use(fiberrecover.New())
	router.Use(newRequestObservabilityMiddleware(obs))
	router.Use(newRequestLogMiddleware(logger))
	router.Use(authMW)
	return fiberAdapter
}

func newHTTPServerOptions(cfg appConfig, logger *slog.Logger, fiberAdapter *httpxfiber.Adapter) []httpx.ServerOption {
	return []httpx.ServerOption{
		httpx.WithAdapter(fiberAdapter),
		httpx.WithBasePath(cfg.basePath()),
		httpx.WithLogger(logx.WithFields(logger, map[string]any{"component": "httpx"})),
		httpx.WithOpenAPIInfo(
			"ArcGo RBAC Backend Scaffold",
			cfg.Version,
			"A reusable RBAC backend scaffold built with arcgo packages",
		),
	}
}

func registerInfraRoutes(server httpx.ServerRuntime, cfg appConfig, prom *promobs.Adapter) {
	server.Adapter().Handle(httpx.MethodGet, "/health", func(
		ctx context.Context,
		w http.ResponseWriter,
		r *http.Request,
	) error {
		_ = ctx
		_ = r
		_, err := w.Write([]byte("ok"))
		return err
	})

	server.Adapter().Handle(httpx.MethodGet, cfg.metricsPath(), func(
		ctx context.Context,
		w http.ResponseWriter,
		r *http.Request,
	) error {
		_ = ctx
		prom.Handler().ServeHTTP(w, r)
		return nil
	})
}

func startHTTPServer(
	lc fx.Lifecycle,
	cfg appConfig,
	logger *slog.Logger,
	server httpx.ServerRuntime,
) {
	var runCancel context.CancelFunc
	listenErrCh := make(chan error, 1)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			runCtx, cancel := context.WithCancel(context.Background())
			runCancel = cancel

			go func() {
				listenErrCh <- server.ListenAndServeContext(runCtx, cfg.addr())
			}()

			select {
			case err := <-listenErrCh:
				return err
			case <-time.After(200 * time.Millisecond):
			}

			logger.Info("rbac backend started",
				slog.String("address", cfg.addr()),
				slog.String("health", fmt.Sprintf("http://127.0.0.1%s/health", cfg.addr())),
				slog.String("docs", fmt.Sprintf("http://127.0.0.1%s%s", cfg.addr(), cfg.docsPath())),
				slog.String("openapi", fmt.Sprintf("http://127.0.0.1%s%s", cfg.addr(), cfg.openAPIPath())),
				slog.String("metrics", fmt.Sprintf("http://127.0.0.1%s%s", cfg.addr(), cfg.metricsPath())),
			)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if runCancel != nil {
				runCancel()
			}

			select {
			case err := <-listenErrCh:
				return err
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
				return nil
			}
		},
	})
}
