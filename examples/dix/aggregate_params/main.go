package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/DaiYuANg/arcgo/dix"
)

type DBConfig struct {
	DSN string
}

type RepositoryParams struct {
	Logger *slog.Logger
	Cfg    DBConfig
}

type Repository struct {
	params RepositoryParams
}

func main() {
	module := dix.NewModule("repository",
		dix.WithModuleProviders(
			dix.Provider0(func() DBConfig { return DBConfig{DSN: "postgres://demo"} }),
			dix.Provider2(func(logger *slog.Logger, cfg DBConfig) RepositoryParams {
				return RepositoryParams{Logger: logger, Cfg: cfg}
			}),
			dix.Provider1(func(params RepositoryParams) *Repository {
				return &Repository{params: params}
			}),
		),
	)

	app := dix.New("aggregate-params", dix.WithModule(module))
	if err := app.Build(); err != nil {
		panic(err)
	}
	if err := app.Start(context.Background()); err != nil {
		panic(err)
	}
	defer func() { _ = app.Stop(context.Background()) }()

	repo, err := dix.ResolveAs[*Repository](app.Container())
	if err != nil {
		panic(err)
	}

	fmt.Println("aggregate params example")
	fmt.Println(repo.params.Cfg.DSN)
}
