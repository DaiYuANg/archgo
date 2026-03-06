# logx

`logx` is a structured logging package built on `zerolog`, with option-based configuration and optional `slog` interop.

[Chinese](./README_ZH.md)

## Features

- Strongly typed levels (`TraceLevel`, `DebugLevel`, `InfoLevel`, ...)
- Console and file output
- File rotation via `lumberjack`
- Optional caller and global logger setup
- `slog` bridge helpers
- Oops integration helpers

## Quick Start

```go
logger, err := logx.New(
    logx.WithConsole(true),
    logx.WithLevel(logx.InfoLevel),
)
if err != nil {
    panic(err)
}
defer logger.Close()

logger.Info("service started", "service", "user-api")
```

## Common Scenarios

### 1) Development profile

```go
logger, err := logx.NewDevelopment()
if err != nil { panic(err) }
defer logger.Close()
```

### 2) Production profile

```go
logger, err := logx.NewProduction()
if err != nil { panic(err) }
defer logger.Close()
```

### 3) File output + rotation

```go
logger, err := logx.New(
    logx.WithConsole(false),
    logx.WithFile("./logs/app.log"),
    logx.WithFileRotation(100, 7, 20), // 100MB, 7 days, 20 backups
    logx.WithCompress(true),
)
```

### 4) Structured fields

```go
logger.WithField("request_id", reqID).Info("request accepted")
logger.WithFields(map[string]any{
    "order_id": orderID,
    "user_id":  userID,
}).Info("order placed")
```

### 5) Context and slog bridge

```go
slogLogger := logx.NewSlog(logger)
slogLogger.Info("hello", "module", "billing")
```

### 6) Attach trace/span IDs from context

```go
ctx := trace.ContextWithSpanContext(context.Background(), spanContext)

logx.WithFieldT(logger, "tenant", "acme").
    WithTraceContext(ctx).
    Info("request accepted")
```

## Level Helpers

- Parse from string: `ParseLevel("debug")`
- Panic on invalid level: `MustParseLevel("info")`
- Convenience constructors: `Trace()`, `Debug()`, `Info()`, ...

## Error/Oops Helpers

- `logger.WithError(err).Error("operation failed")`
- `logger.LogOops(err)`
- `logger.Oops()`, `logger.Oopsf(...)`, `logger.OopsWith(ctx)`

## Operational Notes

- Always call `Close()` when file output is enabled.
- `Sync()` is a no-op by design for `zerolog` (already synchronous write path).
- Use `WithGlobalLogger()` only when a process-wide logger is desired.

## Testing Tips

- Use `WithConsole(true)` in unit tests.
- Use temporary files for file rotation tests.
- Assert `GetLevel()` / `GetLevelString()` for config-level tests.

## FAQ

### Should I use `logx` logger directly or via `slog`?

Both are supported:

- Use `logx` methods for direct `zerolog` ergonomics.
- Use `NewSlog` if your app standardized on `slog` API surface.

### Is `Sync()` required before process exit?

`Sync()` is currently a no-op for this implementation.  
`Close()` is the important lifecycle call when file outputs are enabled.

### Can I set a global logger?

Yes, via `WithGlobalLogger()` or `SetGlobalLogger()`.  
Use it only if your process intentionally shares one logger instance.

## Troubleshooting

### No log file is created

Check:

- `WithFile(path)` is set.
- Process has write permissions for target directory.
- `WithConsole(false)` is not hiding the issue while file setup fails.

### Expected debug logs are missing

Verify log level (`WithLevel(DebugLevel)` or equivalent).  
Higher levels filter lower-severity logs.

### Log rotation does not behave as expected

Review `WithFileRotation(maxSizeMB, maxAgeDays, maxBackups)` values and units.  
`maxSize` is MB, not bytes.

## Anti-Patterns

- Creating short-lived logger instances per request.
- Forgetting `Close()` in services using file output.
- Logging high-cardinality, unbounded fields without sampling controls.
- Using panic/fatal level in recoverable business error paths.
