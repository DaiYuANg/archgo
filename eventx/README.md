# eventx

`eventx` is an in-memory, strongly typed event bus for Go services.

[Chinese](./README_ZH.md)

## Core Capabilities

- Generic typed subscriptions: `Subscribe[T Event]`
- Synchronous publishing: `Publish`
- Asynchronous publishing with queue/workers: `PublishAsync`
- Optional parallel dispatch for handlers of the same event type
- Middleware pipeline (global + per-subscriber)
- Graceful shutdown with in-flight draining (`Close`)

## Event Contract

```go
type Event interface {
    Name() string
}
```

`Name()` is for semantics/observability. Routing is based on Go runtime type.

## Quick Start

```go
type UserCreated struct { ID int }
func (e UserCreated) Name() string { return "user.created" }

bus := eventx.New()
defer bus.Close()

unsub, err := eventx.Subscribe(bus, func(ctx context.Context, evt UserCreated) error {
    fmt.Println(evt.ID)
    return nil
})
if err != nil { panic(err) }
defer unsub()

_ = bus.Publish(context.Background(), UserCreated{ID: 42})
```

## Dispatch Modes

### 1) Deterministic serial dispatch (default)

```go
bus := eventx.New()
```

### 2) Parallel handler dispatch per event

```go
bus := eventx.New(eventx.WithParallelDispatch(true))
```

## Async Publishing

```go
bus := eventx.New(
    eventx.WithAsyncWorkers(8),
    eventx.WithAsyncQueueSize(1024),
    eventx.WithAsyncErrorHandler(func(ctx context.Context, evt eventx.Event, err error) {
        // log/metric/report
    }),
)

err := bus.PublishAsync(ctx, UserCreated{ID: 1})
if errors.Is(err, eventx.ErrAsyncQueueFull) {
    // apply backpressure or fallback strategy
}
```

Behavior notes:

- If async queue/workers are disabled, `PublishAsync` falls back to sync `Publish`.
- When queue is full, `PublishAsync` returns `ErrAsyncQueueFull`.

## Middleware

### Global middleware

```go
bus := eventx.New(
    eventx.WithMiddleware(
        eventx.RecoverMiddleware(),
        eventx.ObserveMiddleware(func(ctx context.Context, evt eventx.Event, d time.Duration, err error) {
            // metrics
        }),
    ),
)
```

### Per-subscriber middleware

```go
_, _ = eventx.Subscribe(
    bus,
    handler,
    eventx.WithSubscriberMiddleware(mySubscriberMw),
)
```

Execution order:

- Global middleware wraps subscriber middleware.
- Middleware order is preserved as provided.

## Error Handling

- `Publish` returns aggregated handler errors (`errors.Join` semantics).
- A panic inside handlers can be converted to error via `RecoverMiddleware`.
- Async errors can be observed via `WithAsyncErrorHandler`.

## Unsubscribe & Lifecycle

- `Subscribe` returns an idempotent `unsubscribe` function.
- `Close` stops new publishes, drains async queue, and waits for in-flight dispatch.
- Calling `Close` multiple times is safe.

## Useful APIs

- `bus.SubscriberCount()` to inspect active subscriptions.
- `eventx.ErrBusClosed`, `eventx.ErrNilEvent`, `eventx.ErrNilBus`, `eventx.ErrNilHandler` for typed error branches.

## Testing Tips

- Use serial dispatch in unit tests for deterministic ordering.
- Call `defer bus.Close()` in each test to avoid worker leakage.
- Use explicit event types per test to avoid accidental shared subscriptions.

## FAQ

### Is `Event.Name()` used for routing?

No. Routing is based on the concrete Go type of the event.  
`Name()` is mainly semantic metadata for logs/metrics/tracing.

### Can one subscriber receive multiple event types?

Use separate `Subscribe[T]` calls per type.  
Each subscription is bound to one generic type `T`.

### Can I recover panics from handlers?

Yes. Add `RecoverMiddleware()` globally or per-subscriber.

## Troubleshooting

### `PublishAsync` returns `ErrAsyncQueueFull`

Options:

- Increase queue size (`WithAsyncQueueSize`).
- Increase workers (`WithAsyncWorkers`).
- Add upstream backpressure/retry policy.
- Fallback to `Publish` for critical events.

### Handlers appear to run in unexpected order

- Serial mode preserves snapshot iteration order.
- Parallel mode (`WithParallelDispatch(true)`) does not guarantee ordering.
- If order matters, keep parallel dispatch disabled for that bus.

### `Close` hangs during shutdown

Usually caused by long-running handlers or blocked downstream calls.  
Pass cancellable contexts and enforce timeouts in handlers.

## Anti-Patterns

- Using one global bus for all domains without clear ownership boundaries.
- Publishing high-volume firehose traffic without queue/backpressure planning.
- Enabling parallel dispatch while also requiring strict order guarantees.
- Ignoring async errors when business-critical events are involved.
