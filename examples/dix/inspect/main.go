package main

import (
	"fmt"

	"github.com/DaiYuANg/arcgo/dix"
	dixadvanced "github.com/DaiYuANg/arcgo/dix/advanced"
	"github.com/DaiYuANg/arcgo/logx"
)

func main() {
	logger, err := logx.NewDevelopment()
	if err != nil {
		panic(err)
	}

	app := dix.New(
		"inspect",
		dix.WithLogger(logger),
		dix.WithModule(
			dix.NewModule("inspect",
				dix.WithModuleProviders(
					dix.Provider0(func() string { return "hello" }),
					dixadvanced.NamedValue("tenant.default", "public"),
				),
			),
		),
	)

	rt, err := app.Build()
	if err != nil {
		panic(err)
	}

	_, err = dix.ResolveAs[string](rt.Container())
	if err != nil {
		panic(err)
	}
	_, err = dixadvanced.ResolveNamedAs[string](rt.Container(), "tenant.default")
	if err != nil {
		panic(err)
	}

	report := dixadvanced.InspectRuntime(rt, "tenant.default")
	fmt.Println("inspect example")
	fmt.Println("provided:", len(report.ProvidedServices))
	fmt.Println("invoked:", len(report.InvokedServices))
	fmt.Println("has tenant deps:", report.NamedDependencies["tenant.default"] != "")
}
