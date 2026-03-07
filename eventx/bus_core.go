package eventx

import (
	"log/slog"
	"reflect"
	"sync"

	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/DaiYuANg/arcgo/observability"
	"github.com/panjf2000/ants/v2"
	"github.com/samber/lo"
)

// Bus is an in-memory strongly typed event bus.
type Bus struct {
	closed        bool
	nextID        uint64
	subsByType    subscriptionTable
	parallel      bool
	middleware    []Middleware
	onAsyncErr    asyncErrorHandler
	antsPool      *ants.Pool
	asyncQueue    chan publishTask
	workerWG      sync.WaitGroup
	dispatchWG    sync.WaitGroup
	observability observability.Observability
	logger        *slog.Logger
}

const (
	metricDispatchTotal           = "eventx_dispatch_total"
	metricDispatchDurationMS      = "eventx_dispatch_duration_ms"
	metricAsyncEnqueueTotal       = "eventx_async_enqueue_total"
	metricAsyncEnqueueDurationMS  = "eventx_async_enqueue_duration_ms"
	metricAsyncDispatchErrorTotal = "eventx_async_dispatch_error_total"
)

// New creates a new Bus.
func New(opts ...Option) *Bus {
	cfg := defaultOptions()
	lo.ForEach(opts, func(opt Option, _ int) {
		if opt != nil {
			opt(&cfg)
		}
	})

	b := &Bus{
		subsByType:    collectionx.NewConcurrentTable[reflect.Type, uint64, *subscription](),
		parallel:      cfg.parallel,
		middleware:    cfg.middleware,
		onAsyncErr:    cfg.onAsyncError,
		observability: observability.Normalize(cfg.observability, nil),
	}
	b.logger = b.observability.Logger().With("component", "eventx.bus")

	// Initialize ants pool if enabled
	if cfg.useAntsPool {
		poolOpts := []ants.Option{
			ants.WithPreAlloc(true),
			ants.WithNonblocking(false),
		}
		if cfg.antsMaxBlockingCalls > 0 {
			poolOpts = append(poolOpts, ants.WithMaxBlockingTasks(cfg.antsMaxBlockingCalls))
		}

		pool, err := ants.NewPool(cfg.antsPoolSize, poolOpts...)
		if err != nil {
			b.logger.Error("failed to create ants pool", "error", err)
		} else {
			b.antsPool = pool
		}
	} else if cfg.asyncWorkers > 0 && cfg.asyncQueueSize > 0 {
		// Legacy mode: use channel-based worker pool
		b.asyncQueue = make(chan publishTask, cfg.asyncQueueSize)
		for i := 0; i < cfg.asyncWorkers; i++ {
			b.workerWG.Add(1)
			go b.workerLoop()
		}
	}

	return b
}

// Close stops accepting new events, drains async queue, and waits in-flight handlers.
func (b *Bus) Close() error {
	if b == nil {
		return nil
	}

	if b.closed {
		return nil
	}
	b.closed = true

	queue := b.asyncQueue
	if queue != nil {
		close(queue)
	}

	// Release ants pool if enabled
	if b.antsPool != nil {
		b.antsPool.Release()
	}

	// Wait for legacy worker pool if enabled
	if queue != nil {
		b.workerWG.Wait()
	}
	b.dispatchWG.Wait()
	return nil
}

// SubscriberCount returns active subscriber count.
func (b *Bus) SubscriberCount() int {
	if b == nil {
		return 0
	}
	return b.subsByType.Len()
}
