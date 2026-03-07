# EventX 使用示例

本目录包含 `eventx` 包的使用示例，展示不同场景下的最佳实践。

## 示例列表

### 1. 基础示例 (basic)

展示事件总线的基本用法，包括：
- 创建事件总线
- 定义和发布事件
- 同步和异步事件处理

**运行：**
```bash
go run ./eventx/examples/basic/
```

**示例场景：**
- 订单创建事件（同步处理）
- 订单支付事件（异步处理）
- 多个订阅者监听同一事件

### 2. 中间件示例 (middleware)

展示中间件的使用，包括：
- 全局中间件（日志记录、Panic 恢复）
- 订阅者级别中间件（权限检查）
- 中间件的执行顺序

**运行：**
```bash
go run ./eventx/examples/middleware/
```

**示例场景：**
- 用户注册事件处理
- 用户登录事件处理
- Panic 恢复中间件演示

### 3. FX 集成示例 (fx)

展示如何与 Uber FX 框架集成：
- 使用 `eventx/fx` 模块
- 集成 `logx/fx` 日志模块
- 依赖注入事件总线和日志
- FX 生命周期管理

**运行：**
```bash
go run ./eventx/examples/fx/
```

**示例场景：**
- 通知系统（邮件、短信、推送）
- 结构化日志记录
- 完整的 FX 应用示例

**输出示例：**
```
=== EventX + FX + LogX 集成示例 ===

2026-03-07 23:17:12 INF 🚀 正在注册通知订阅者...
2026-03-07 23:17:12 INF ✅ 所有通知订阅者已注册完成
2026-03-07 23:17:12 INF 📨 开始发布通知事件...

=== 发布通知事件 ===
2026-03-07 23:17:12 INF 发布通知事件 index=1 total=4 type=email user_id=user-001
2026-03-07 23:17:12 INF 发布通知事件 index=2 total=4 type=sms user_id=user-001
2026-03-07 23:17:12 INF 发布通知事件 index=3 total=4 type=push user_id=user-002
2026-03-07 23:17:12 INF 发布通知事件 index=4 total=4 type=email user_id=user-002
2026-03-07 23:17:12 INF ✅ 所有通知已发布到异步队列

✅ 所有通知已发布到异步队列
2026-03-07 23:17:12 INF 📧 发送邮件 msg_type=email user_id=user-002
   📧 发送邮件给用户 user-002: 订单已发货
2026-03-07 23:17:12 INF 🔔 发送推送 msg_type=push user_id=user-002
   🔔 发送推送给用户 user-002: 您有新的消息
```

### 4. 高性能示例 (benchmark)

展示高性能批量处理场景：
- 多协程池配置
- 并行分发
- 异步错误处理
- 性能统计

**运行：**
```bash
go run ./eventx/examples/benchmark/
```

**示例场景：**
- 库存变更事件批量处理
- 多消费者并发处理
- 吞吐量统计

## 核心概念

### 1. 事件定义

所有事件必须实现 `Event` 接口：

```go
type Event interface {
    Name() string
}
```

示例：
```go
type OrderCreatedEvent struct {
    OrderID string
    UserID  string
    Amount  float64
}

func (e OrderCreatedEvent) Name() string {
    return "order.created"
}
```

### 2. 订阅事件

使用泛型订阅特定类型的事件：

```go
unsubscribe, err := eventx.Subscribe[OrderCreatedEvent](bus, 
    func(ctx context.Context, event OrderCreatedEvent) error {
        // 处理事件
        return nil
    },
    // 可选：订阅者中间件
    eventx.WithSubscriberMiddleware(...),
)
```

### 3. 发布事件

**同步发布：**
```go
err := bus.Publish(ctx, event)
```

**异步发布：**
```go
err := bus.PublishAsync(ctx, event)
```

### 4. 配置选项

```go
bus := eventx.New(
    // 使用 ants 协程池（推荐）
    eventx.WithAntsPool(10),
    
    // 并行分发
    eventx.WithParallelDispatch(true),
    
    // 全局中间件
    eventx.WithMiddleware(middleware1, middleware2),
    
    // 异步错误处理
    eventx.WithAsyncErrorHandler(handler),
)
```

## 最佳实践

### 1. 选择合适的发布方式

- **同步发布**：需要立即知道处理结果，事件处理快速
- **异步发布**：事件处理耗时，不需要立即知道结果

### 2. 使用中间件

常用中间件：
- `RecoverMiddleware()`: 恢复 panic，防止总线崩溃
- 日志记录中间件
- 性能监控中间件

### 3. 合理配置协程池

```go
// 低负载场景
eventx.WithAntsPool(4)

// 高负载场景
eventx.WithAntsPool(20)
```

### 4. 清理资源

应用退出时关闭总线：

```go
defer bus.Close()
```

### 5. 取消订阅

使用返回的取消订阅函数：

```go
unsubscribe, err := eventx.Subscribe[MyEvent](bus, handler)
// 不再需要时
unsubscribe()
```

## 性能提示

1. **并行分发**：当有多个订阅者时启用
2. **协程池大小**：根据 CPU 核心数和负载调整
3. **避免阻塞**：异步事件处理中避免长时间阻塞操作
4. **批量处理**：使用异步发布处理批量事件

## 常见问题

### Q: 什么时候使用同步 vs 异步发布？

**A:** 
- 同步：需要事务一致性，事件处理快速（<10ms）
- 异步：事件处理耗时，可以最终一致性

### Q: 如何处理异步事件的错误？

**A:** 使用 `WithAsyncErrorHandler` 配置全局错误处理器。

### Q: 如何保证事件处理顺序？

**A:** 
- 同一事件类型的多个订阅者：不保证顺序
- 需要顺序处理：使用单个订阅者，内部队列处理
