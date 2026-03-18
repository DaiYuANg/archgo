package main

import (
	"context"
	"fmt"

	"github.com/DaiYuANg/arcgo/dix"
	dixadvanced "github.com/DaiYuANg/arcgo/dix/advanced"
	"github.com/DaiYuANg/arcgo/logx"
	do "github.com/samber/do/v2"
)

type AppConfig struct {
	Name string
}

type RequestContext struct {
	RequestID string
}

type ScopedService struct {
	Config  AppConfig
	Request RequestContext
}

func main() {
	logger, err := logx.NewDevelopment()
	if err != nil {
		panic(err)
	}

	app := dix.New(
		"runtime-scope",
		dix.WithLogger(logger),
		dix.WithModule(
			dix.NewModule("root",
				dix.WithModuleProviders(
					dix.Provider0(func() AppConfig {
						return AppConfig{Name: "arcgo"}
					}),
				),
			),
		),
	)

	rt, err := app.Build()
	if err != nil {
		panic(err)
	}
	if err := rt.Start(context.Background()); err != nil {
		panic(err)
	}
	defer func() { _ = rt.Stop(context.Background()) }()

	requestScope := dixadvanced.Scope(rt, "request-42", func(injector do.Injector) {
		dixadvanced.ProvideScopedValue(injector, RequestContext{RequestID: "req-42"})
		dixadvanced.ProvideScoped2(injector, func(cfg AppConfig, req RequestContext) ScopedService {
			return ScopedService{Config: cfg, Request: req}
		})
	})

	service, err := dixadvanced.ResolveScopedAs[ScopedService](requestScope)
	if err != nil {
		panic(err)
	}

	_, rootCanResolveRequest := dixadvanced.ResolveRuntimeAs[RequestContext](rt)

	fmt.Println("runtime scope example")
	fmt.Println(service.Config.Name)
	fmt.Println(service.Request.RequestID)
	fmt.Println("root sees request context:", rootCanResolveRequest == nil)
}
