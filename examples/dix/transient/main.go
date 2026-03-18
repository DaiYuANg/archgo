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

	counter := 0
	app := dix.New(
		"transient",
		dix.WithLogger(logger),
		dix.WithModule(
			dix.NewModule("transient",
				dix.WithModuleProviders(
					dixadvanced.TransientProvider0(func() int {
						counter++
						return counter
					}),
				),
			),
		),
	)

	rt, err := app.Build()
	if err != nil {
		panic(err)
	}

	first, err := dix.ResolveAs[int](rt.Container())
	if err != nil {
		panic(err)
	}
	second, err := dix.ResolveAs[int](rt.Container())
	if err != nil {
		panic(err)
	}

	fmt.Println("transient example")
	fmt.Println("first:", first)
	fmt.Println("second:", second)
	fmt.Println("different:", first != second)
}
