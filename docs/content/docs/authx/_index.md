---
title: 'authx'
linkTitle: 'authx'
description: 'Extensible authentication and authorization abstraction for multiple scenarios'
weight: 1
---

## authx

`authx` is a Go authentication and authorization abstraction for multi-scenario use (HTTP / gRPC / CLI).

Core principles:

- Separation of authentication and authorization: `Check` / `Can`
- Authentication-mechanism agnostic: JWT / password / OTP and more
- Framework-agnostic core with scenario-specific integration layers

## Roadmap

- Module roadmap: [authx roadmap](./roadmap)
- Iteration execution plan: [authx iteration plan](./iteration-plan)
- New version note: [authx v0.3.0 release](./release-v0.3.0)
- Global roadmap: [ArcGo roadmap](../roadmap)

## Core API

- `Engine`: orchestrates authentication and authorization
- `ProviderManager`: manages providers for multiple credential types
- `AuthenticationProvider[C]`: generic provider abstraction
- `Authorizer`: authorization decision interface
- `Check(ctx, credential)`: authenticate
- `Can(ctx, AuthorizationModel)`: authorize
- `Hook`: before/after hooks for Check/Can

## Quick Start (Core)

```go
engine := authx.NewEngine(
    authx.WithAuthenticationManager(
        authx.NewProviderManager(
            authx.NewAuthenticationProviderFunc(func(
                _ context.Context,
                in UsernamePassword,
            ) (authx.AuthenticationResult, error) {
                return authx.AuthenticationResult{
                    Principal: authx.Principal{ID: in.Username},
                }, nil
            }),
        ),
    ),
    authx.WithAuthorizer(authx.AuthorizerFunc(func(
        _ context.Context,
        model authx.AuthorizationModel,
    ) (authx.Decision, error) {
        return authx.Decision{Allowed: true}, nil
    })),
)

result, err := engine.Check(ctx, UsernamePassword{Username: "alice", Password: "secret"})
if err != nil {
    panic(err)
}

decision, err := engine.Can(ctx, authx.AuthorizationModel{
    Principal: result.Principal,
    Action:    "query",
    Resource:  "order",
})
if err != nil {
    panic(err)
}
_ = decision
```

## HTTP Integrations

`authx/http` provides a unified Guard plus middleware integrations:

- `authx/http/std`
- `authx/http/gin`
- `authx/http/echo`
- `authx/http/fiber`

Unified extension points:

- `WithCredentialResolverFunc`
- `WithAuthorizationResolverFunc`

```go
guard := authhttp.NewGuard(
    engine,
    authhttp.WithCredentialResolverFunc(resolveCredential),
    authhttp.WithAuthorizationResolverFunc(resolveAuthorization),
)

router.Use(authstd.Require(guard))
// hot path: router.Use(authstd.RequireFast(guard))
```

## Examples

- `authx/http/examples/shared`
- `authx/http/examples/jwt`
- `authx/http/examples/std`
- `authx/http/examples/gin`
- `authx/http/examples/echo`
- `authx/http/examples/fiber`

## Testing and Benchmarks

```bash
go test ./authx/...

# core
go test ./authx -run ^$ -bench BenchmarkEngine -benchmem

# middleware
go test ./authx/http/std -run ^$ -bench BenchmarkRequire -benchmem
go test ./authx/http/gin -run ^$ -bench BenchmarkRequire -benchmem
go test ./authx/http/echo -run ^$ -bench BenchmarkRequire -benchmem
go test ./authx/http/fiber -run ^$ -bench BenchmarkRequire -benchmem
```
