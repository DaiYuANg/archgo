package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/DaiYuANg/arcgo/dix"
	dixadvanced "github.com/DaiYuANg/arcgo/dix/advanced"
	"github.com/DaiYuANg/arcgo/logx"
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
	logger, err := logx.NewDevelopment()
	if err != nil {
		panic(err)
	}

	serviceModule := dix.NewModule("greeter",
		dix.WithModuleProviders(
			dix.Provider1(func(logger *slog.Logger) *englishGreeter {
				return &englishGreeter{logger: logger}
			}),
			dixadvanced.NamedValue("locale.default", "en-US"),
			dixadvanced.NamedProvider1[*englishGreeter, *slog.Logger]("greeter.en", func(logger *slog.Logger) *englishGreeter {
				return &englishGreeter{logger: logger}
			}),
		),
		dix.WithModuleSetups(
			dixadvanced.BindAlias[*englishGreeter, Greeter](),
			dixadvanced.BindNamedAlias[*englishGreeter, Greeter]("greeter.en", "greeter.en.alias"),
		),
	)

	app := dix.New("named-alias", dix.WithModule(serviceModule), dix.WithLogger(logger))
	rt, err := app.Build()
	if err != nil {
		panic(err)
	}
	if err := rt.Start(context.Background()); err != nil {
		panic(err)
	}
	defer func() { _ = rt.Stop(context.Background()) }()

	locale, err := dixadvanced.ResolveNamedAs[string](rt.Container(), "locale.default")
	if err != nil {
		panic(err)
	}
	fmt.Println("locale:", locale)

	greeter, err := dix.ResolveAs[Greeter](rt.Container())
	if err != nil {
		panic(err)
	}
	fmt.Println("implicit/assignable alias:", greeter.Greet())

	namedAlias, err := dixadvanced.ResolveNamedAs[Greeter](rt.Container(), "greeter.en.alias")
	if err != nil {
		panic(err)
	}
	fmt.Println("named explicit alias:", namedAlias.Greet())
}
