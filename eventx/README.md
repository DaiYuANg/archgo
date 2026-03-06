# eventx

`eventx` 是一个内存级的强类型事件总线，提供：

- 泛型强类型订阅：`Subscribe[T Event](...)`
- 同步发布：`Publish`
- 异步发布：`PublishAsync`（worker 池）
- 可选并行分发：`WithParallelDispatch(true)`
- middleware 链：恢复 panic、观测耗时等
- 优雅关闭：`Close` 等待队列与 in-flight handler 完成

## 安装

```bash
go get github.com/DaiYuANg/arcgo/eventx
```

## 事件定义

```go
type UserCreated struct {
    ID int
}

func (e UserCreated) Name() string { return "user.created" }
```

## 快速开始

```go
package main

import (
    "context"
    "fmt"

    "github.com/DaiYuANg/arcgo/eventx"
)

type UserCreated struct {
    ID int
}

func (e UserCreated) Name() string { return "user.created" }

func main() {
    bus := eventx.New(
        eventx.WithMiddleware(eventx.RecoverMiddleware()),
    )
    defer bus.Close()

    unsubscribe, _ := eventx.Subscribe(bus, func(ctx context.Context, evt UserCreated) error {
        fmt.Println("user created:", evt.ID)
        return nil
    })
    defer unsubscribe()

    _ = bus.Publish(context.Background(), UserCreated{ID: 42})
}
```

## 设计说明

- 路由按 Go 运行时类型分发（`reflect.TypeOf(event)`），类型必须匹配订阅时的 `T`。
- `Event.Name()` 主要用于观测和语义表达，不作为路由键。
- `PublishAsync` 默认非阻塞入队；队列满返回 `ErrAsyncQueueFull`。
