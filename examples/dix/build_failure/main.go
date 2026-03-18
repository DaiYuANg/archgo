package main

import (
	"fmt"

	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/arcgo/logx"
)

type MissingDependency struct {
	Name string
}

type NeedsMissingDependency struct {
	Name string
}

func main() {
	logger, err := logx.NewDevelopment()
	if err != nil {
		panic(err)
	}

	app := dix.New(
		"build-failure",
		dix.WithLogger(logger),
		dix.WithModule(
			dix.NewModule("broken",
				dix.WithModuleProviders(
					dix.Provider1(func(dep MissingDependency) NeedsMissingDependency {
						return NeedsMissingDependency{Name: dep.Name}
					}),
				),
				dix.WithModuleInvokes(
					dix.Invoke1(func(value NeedsMissingDependency) {
						fmt.Println(value.Name)
					}),
				),
			),
		),
	)

	if err := app.Validate(); err != nil {
		fmt.Println("validate error:", err)
	} else {
		fmt.Println("validate error: <nil>")
	}

	_, err = app.Build()
	fmt.Println("build error:", err != nil)
}
