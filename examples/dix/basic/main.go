package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/logx"
)

type Config struct {
	Port int
}

type Server struct {
	Logger *slog.Logger
	Config Config
}

func main() {
	configModule := dix.NewModule("config",
		dix.WithModuleProviders(
			dix.Provider0(func() Config { return Config{Port: 8080} }),
		),
	)

	logger, err := logx.NewDevelopment()
	if err != nil {
		panic(err)
	}
	serverModule := dix.NewModule("server",
		dix.WithModuleImports(configModule),
		dix.WithModuleProviders(
			dix.Provider2(func(logger *slog.Logger, cfg Config) *Server {
				return &Server{Logger: logger, Config: cfg}
			}),
		),
		dix.WithModuleSetup(func(c *dix.Container, lc dix.Lifecycle) error {
			c.RegisterLivenessCheck("process", func(context.Context) error { return nil })
			c.RegisterReadinessCheck("bootstrap", func(context.Context) error {
				server, ok := dix.ResolveOptionalAs[*Server](c)
				if !ok || server == nil {
					return errors.New("server not ready")
				}
				return nil
			})
			return nil
		}),
	)

	app := dix.NewDefault(
		dix.WithProfile(dix.ProfileDev),
		dix.WithVersion("0.5.0"),
		dix.WithModules(serverModule),
		dix.WithLogger(logger),
	)

	rt, err := app.Build()
	if err != nil {
		panic(err)
	}
	if err := rt.Start(context.Background()); err != nil {
		panic(err)
	}
	defer func() { _ = rt.Stop(context.Background()) }()

	fmt.Println("basic app built and started")
	fmt.Printf("health: %v\n", rt.CheckHealth(context.Background()).Healthy())
	fmt.Printf("liveness: %v\n", rt.CheckLiveness(context.Background()).Healthy())
	fmt.Printf("readiness: %v\n", rt.CheckReadiness(context.Background()).Healthy())

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", rt.HealthHandler())
	mux.HandleFunc("/livez", rt.LivenessHandler())
	mux.HandleFunc("/readyz", rt.ReadinessHandler())
	_ = mux

	time.Sleep(100 * time.Millisecond)
}
