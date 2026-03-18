package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/DaiYuANg/arcgo/dix"
)

type Greeter interface {
	Greet() string
}

type englishGreeter struct {
	logger *slog.Logger
}

func (g *englishGreeter) Greet() string {
	g.logger.Info("greet invoked", "lang", "en")
	return "hello"
}

func main() {
	serviceModule := dix.NewModule("greeter",
		dix.WithModuleProviders(
			dix.Provider1(func(logger *slog.Logger) *englishGreeter {
				return &englishGreeter{logger: logger}
			}),
			// named service wrapper around do.ProvideNamed
			dix.ProviderFunc(func(c *dix.Container) {
				dix.ProvideNamedValueT[string](c, "locale.default", "en-US")
				dix.ProvideNamed1T[*englishGreeter, *slog.Logger](c, "greeter.en", func(logger *slog.Logger) *englishGreeter {
					return &englishGreeter{logger: logger}
				})
			}),
		),
		dix.WithModuleSetup(func(c *dix.Container, lc dix.Lifecycle) error {
			// explicit alias wrapper around do.As / do.AsNamed
			if err := dix.BindAlias[*englishGreeter, Greeter](c); err != nil {
				return err
			}
			if err := dix.BindNamedAlias[*englishGreeter, Greeter](c, "greeter.en.alias", "greeter"); err != nil {
				return err
			}
			return nil
		}),
	)

	app := dix.New("named-alias", dix.WithModule(serviceModule))
	if err := app.Build(); err != nil {
		panic(err)
	}
	if err := app.Start(context.Background()); err != nil {
		panic(err)
	}
	defer func() { _ = app.Stop(context.Background()) }()

	locale, err := dix.ResolveNamedAs[string](app.Container(), "locale.default")
	if err != nil {
		panic(err)
	}
	fmt.Println("locale:", locale)

	greeter, err := dix.ResolveAssignableAs[Greeter](app.Container())
	if err != nil {
		panic(err)
	}
	fmt.Println("implicit/assignable alias:", greeter.Greet())

	namedAlias, err := dix.ResolveNamedAs[Greeter](app.Container(), "greeter.en.alias")
	if err != nil {
		panic(err)
	}
	fmt.Println("named explicit alias:", namedAlias.Greet())
}
