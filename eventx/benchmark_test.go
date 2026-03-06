package eventx

import (
	"context"
	"testing"
)

type benchmarkEvent struct {
	ID int
}

func (e benchmarkEvent) Name() string {
	return "benchmark.event"
}

func benchmarkBusWithSubscribers(b *testing.B, parallelDispatch bool, subscribers int) *Bus {
	b.Helper()

	bus := New(WithParallelDispatch(parallelDispatch))
	for i := 0; i < subscribers; i++ {
		_, err := Subscribe(bus, func(ctx context.Context, evt benchmarkEvent) error {
			_ = ctx
			_ = evt
			return nil
		})
		if err != nil {
			b.Fatalf("subscribe failed: %v", err)
		}
	}

	b.Cleanup(func() {
		_ = bus.Close()
	})
	return bus
}

func BenchmarkBusPublishSerial(b *testing.B) {
	bus := benchmarkBusWithSubscribers(b, false, 1)
	ctx := context.Background()
	evt := benchmarkEvent{ID: 1}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := bus.Publish(ctx, evt); err != nil {
			b.Fatalf("publish failed: %v", err)
		}
	}
}

func BenchmarkBusPublishParallelDispatch(b *testing.B) {
	bus := benchmarkBusWithSubscribers(b, true, 4)
	ctx := context.Background()
	evt := benchmarkEvent{ID: 1}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := bus.Publish(ctx, evt); err != nil {
			b.Fatalf("publish failed: %v", err)
		}
	}
}

func BenchmarkBusPublishConcurrentPublishers(b *testing.B) {
	bus := benchmarkBusWithSubscribers(b, false, 2)
	evt := benchmarkEvent{ID: 1}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			if err := bus.Publish(ctx, evt); err != nil {
				b.Fatalf("publish failed: %v", err)
			}
		}
	})
}
