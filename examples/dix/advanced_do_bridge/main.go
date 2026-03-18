package main

import (
	"context"
	"fmt"

	"github.com/DaiYuANg/arcgo/dix"
	dixadvanced "github.com/DaiYuANg/arcgo/dix/advanced"
	"github.com/DaiYuANg/arcgo/logx"
	do "github.com/samber/do/v2"
)

type NamedValue string

func main() {
	logger, err := logx.NewDevelopment()
	if err != nil {
		panic(err)
	}

	module := dix.NewModule("advanced-bridge",
		dix.WithModuleSetups(
			dixadvanced.DoSetup(func(raw do.Injector) error {
				do.ProvideNamedValue(raw, "tenant.default", NamedValue("public"))
				return nil
			}),
		),
	)

	app := dix.New(
		"advanced-do-bridge",
		dix.WithDebugScopeTree(true),
		dix.WithDebugNamedServiceDependencies("tenant.default"),
		dix.WithLogger(logger),
		dix.WithModule(module),
	)

	rt, err := app.Build()
	if err != nil {
		panic(err)
	}
	if err := rt.Start(context.Background()); err != nil {
		panic(err)
	}
	defer func() { _ = rt.Stop(context.Background()) }()

	value, err := dixadvanced.ResolveNamedAs[NamedValue](rt.Container(), "tenant.default")
	if err != nil {
		panic(err)
	}

	fmt.Println("advanced do bridge example")
	fmt.Println(value)
}
