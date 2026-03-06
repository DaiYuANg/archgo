# configx

`configx` is a layered configuration loader built on top of `koanf` and `validator`.

[Chinese](./README_ZH.md)

## What It Supports

- `.env` loading (`WithDotenv`)
- Config file loading (`WithFiles`)
- Environment variable loading (`WithEnvPrefix`)
- Custom source precedence (`WithPriority`)
- Defaults via map or struct (`WithDefaults`, `WithDefaultsStruct`)
- Optional validation (`WithValidateLevel`, `WithValidator`)
- Generic and non-generic loading entry points

## Loading Flow

`configx` merges sources by priority. Later sources override earlier ones.

Default priority:

1. dotenv
2. files
3. env vars

## Quick Start

```go
type AppConfig struct {
    Name string `mapstructure:"name" validate:"required"`
    Port int    `mapstructure:"port" validate:"required,min=1,max=65535"`
}

var cfg AppConfig
err := configx.Load(&cfg,
    configx.WithDotenv(),
    configx.WithFiles("config.yaml"),
    configx.WithEnvPrefix("APP"),
    configx.WithValidateLevel(configx.ValidateLevelRequired),
)
if err != nil {
    panic(err)
}
```

## Common Scenarios

### 1) Local development (`.env` first)

```go
err := configx.Load(&cfg,
    configx.WithDotenv(".env", ".env.local"),
    configx.WithIgnoreDotenvError(true),
)
```

### 2) File + environment override

```go
err := configx.Load(&cfg,
    configx.WithFiles("config.yaml"),
    configx.WithEnvPrefix("APP"),
    configx.WithPriority(configx.SourceFile, configx.SourceEnv),
)
```

### 3) Bootstrap with defaults only

```go
err := configx.Load(&cfg,
    configx.WithDefaults(map[string]any{
        "name": "my-service",
        "port": 8080,
    }),
)
```

### 4) Defaults from a struct

```go
type DefaultCfg struct {
    Name string `mapstructure:"name"`
    Port int    `mapstructure:"port"`
}

err := configx.Load(&cfg,
    configx.WithDefaultsStruct(DefaultCfg{Name: "svc", Port: 8080}),
)
```

### 5) Generic loading API

```go
result := configx.LoadT[AppConfig](
    configx.WithFiles("config.yaml"),
)
if result.IsError() {
    panic(result.Error())
}
cfg := result.MustGet()
```

### 6) Explicit `Config` object usage

```go
c, err := configx.LoadConfig(
    configx.WithFiles("config.yaml"),
)
if err != nil {
    panic(err)
}

name := c.GetString("app.name")
port := c.GetInt("app.port")
exists := c.Exists("app.debug")
all := c.All()
_, _, _, _ = name, port, exists, all
```

## Validation Modes

- `ValidateLevelNone`: no validation
- `ValidateLevelStruct`: run struct validation
- `ValidateLevelRequired`: required tags enforced (same struct validation path)

If you need custom validators/tags:

```go
v := validator.New(validator.WithRequiredStructEnabled())
err := configx.Load(&cfg,
    configx.WithValidator(v),
    configx.WithValidateLevel(configx.ValidateLevelRequired),
)
```

## Environment Key Mapping

With `WithEnvPrefix("APP")`:

- `APP_DATABASE_HOST` -> `database.host`
- `APP_SERVER_READ_TIMEOUT` -> `server.read.timeout`

## Production Tips

- Keep source precedence explicit in production builds.
- Use defaults for non-critical values to reduce startup failures.
- Use validation for critical fields (ports, credentials, hostnames).
- Keep `.env` optional in production unless explicitly required.

## Test Patterns

- Use `WithDefaults` for deterministic tests.
- Avoid real env dependencies in unit tests unless test isolates `os.Environ`.
- Use `LoadT[T]` in tests to reduce boilerplate.

## FAQ

### Which source should have highest priority?

In most services, environment variables should be highest priority in production.  
A common order is: defaults -> file -> env.

### Should I use `Load` or `LoadConfig`?

- Use `Load` if you just need one typed struct.
- Use `LoadConfig` when you also need dynamic getters (`GetString`, `Exists`, `All`) after load.

### Map defaults vs struct defaults?

- `WithDefaults(map[string]any)` is explicit and dynamic.
- `WithDefaultsStruct` is convenient when you already have typed default config structs.

## Troubleshooting

### Environment values are not taking effect

Check these first:

- `WithEnvPrefix` matches actual env key prefix.
- `WithPriority` places `SourceEnv` after other sources.
- Env keys map to dot-path format (`APP_DB_HOST` -> `db.host`).

### Validation does not run

Validation is disabled unless configured.  
Set `WithValidateLevel(...)`, or wire `WithValidator(...)` plus validation level.

### `.env` file missing crashes startup

Use `WithIgnoreDotenvError(true)` in environments where `.env` is optional.

### `WithDefaultsStruct` fails for unsupported types

The struct-to-map conversion is reflection-based.  
Keep defaults structs simple and export fields with predictable `mapstructure` tags.

## Anti-Patterns

- Relying on implicit source precedence in production.
- Reading config from process env directly in business code after adopting `configx`.
- Disabling validation for critical fields (ports, credentials, URLs).
- Mixing unrelated prefixes across multiple services in shared environments.
