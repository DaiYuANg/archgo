package main

import (
	"fmt"

	"github.com/DaiYuANg/arcgo/dix"
	dixadvanced "github.com/DaiYuANg/arcgo/dix/advanced"
	"github.com/DaiYuANg/arcgo/logx"
)

type AppConfig struct {
	Env string
}

func main() {
	logger, err := logx.NewDevelopment()
	if err != nil {
		panic(err)
	}

	app := dix.New(
		"override",
		dix.WithLogger(logger),
		dix.WithModule(
			dix.NewModule("override",
				dix.WithModuleProviders(
					dix.Provider0(func() AppConfig { return AppConfig{Env: "dev"} }),
				),
				dix.WithModuleSetups(
					dixadvanced.Override0(func() AppConfig { return AppConfig{Env: "prod"} }),
				),
			),
		),
	)

	if err := app.Validate(); err != nil {
		panic(err)
	}

	rt, err := app.Build()
	if err != nil {
		panic(err)
	}

	cfg, err := dix.ResolveAs[AppConfig](rt.Container())
	if err != nil {
		panic(err)
	}

	fmt.Println("override example")
	fmt.Println("env:", cfg.Env)
}
