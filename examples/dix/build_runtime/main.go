package main

import (
	"context"
	"fmt"

	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/logx"
)

type Greeting struct {
	Message string
}

func main() {
	logger, err := logx.NewDevelopment()
	if err != nil {
		panic(err)
	}

	module := dix.NewModule("greeting",
		dix.WithModuleProviders(
			dix.Provider0(func() Greeting {
				return Greeting{Message: "hello runtime"}
			}),
		),
	)

	app := dix.New(
		"build-runtime",
		dix.WithVersion("0.6.0"),
		dix.WithLogger(logger),
		dix.WithModule(module),
	)

	first, err := app.Build()
	if err != nil {
		panic(err)
	}
	second, err := app.Build()
	if err != nil {
		panic(err)
	}

	greeting, err := dix.ResolveAs[Greeting](first.Container())
	if err != nil {
		panic(err)
	}

	if err := first.Start(context.Background()); err != nil {
		panic(err)
	}
	defer func() { _ = first.Stop(context.Background()) }()

	fmt.Println("build runtime example")
	fmt.Println("app name:", app.Name())
	fmt.Println("first runtime started:", first.State() == dix.AppStateStarted)
	fmt.Println("second runtime built:", second.State() == dix.AppStateBuilt)
	fmt.Println("independent runtimes:", first != second)
	fmt.Println(greeting.Message)
}
