package main

import (
	"context"
	"fmt"

	"github.com/DaiYuANg/arcgo/dix"
	do "github.com/samber/do/v2"
)

type NamedValue string

func main() {
	module := dix.NewModule("advanced-bridge",
		dix.WithModuleDoSetup(func(raw do.Injector) error {
			do.ProvideNamedValue(raw, "tenant.default", NamedValue("public"))
			return nil
		}),
	)

	app := dix.New(
		"advanced-do-bridge",
		dix.WithDebugScopeTree(true),
		dix.WithDebugNamedServiceDependencies("tenant.default"),
		dix.WithModule(module),
	)

	if err := app.Build(); err != nil {
		panic(err)
	}
	if err := app.Start(context.Background()); err != nil {
		panic(err)
	}
	defer func() { _ = app.Stop(context.Background()) }()

	value, err := dix.ResolveNamedAs[NamedValue](app.Container(), "tenant.default")
	if err != nil {
		panic(err)
	}

	fmt.Println("advanced do bridge example")
	fmt.Println(value)
}
