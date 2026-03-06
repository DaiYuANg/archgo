# logx

`logx` 是基于 `zerolog` 的结构化日志组件，支持 Option 配置、文件轮转和 `slog` 互通。

[English](./README.md)

## 核心能力

- 强类型日志级别（`TraceLevel` / `DebugLevel` / `InfoLevel` 等）
- 控制台与文件双输出
- 基于 `lumberjack` 的文件轮转
- 可选调用者信息和全局 logger
- `slog` 适配辅助方法
- oops 错误集成

## 快速开始

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

## 常见场景

### 1) 开发环境

```go
logger, err := logx.NewDevelopment()
if err != nil { panic(err) }
defer logger.Close()
```

### 2) 生产环境

```go
logger, err := logx.NewProduction()
if err != nil { panic(err) }
defer logger.Close()
```

### 3) 文件落盘与轮转

```go
logger, err := logx.New(
    logx.WithConsole(false),
    logx.WithFile("./logs/app.log"),
    logx.WithFileRotation(100, 7, 20),
    logx.WithCompress(true),
)
```

### 4) 结构化字段

```go
logger.WithField("request_id", reqID).Info("request accepted")
logger.WithFields(map[string]any{
    "order_id": orderID,
    "user_id":  userID,
}).Info("order placed")
```

### 5) `slog` 互通

```go
slogLogger := logx.NewSlog(logger)
slogLogger.Info("hello", "module", "billing")
```

### 6) 从上下文注入 trace/span

```go
ctx := trace.ContextWithSpanContext(context.Background(), spanContext)

logx.WithFieldT(logger, "tenant", "acme").
    WithTraceContext(ctx).
    Info("request accepted")
```

## 级别与错误

- 字符串解析：`ParseLevel("debug")`
- 强制解析：`MustParseLevel("info")`
- 错误附加：`logger.WithError(err).Error("operation failed")`
- oops：`LogOops` / `Oops` / `Oopsf` / `OopsWith`

## 实践建议

- 开启文件输出时，务必在退出前 `Close()`。
- 若要全局 logger，请显式使用 `WithGlobalLogger()`。
- 测试中优先使用控制台输出与临时文件。
