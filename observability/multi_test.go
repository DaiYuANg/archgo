package observability

import (
	"context"
	"log/slog"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

type testBackend struct {
	logger *slog.Logger

	spanCount      atomic.Int64
	counterCount   atomic.Int64
	histogramCount atomic.Int64
}

func (t *testBackend) Logger() *slog.Logger {
	return NormalizeLogger(t.logger)
}

func (t *testBackend) StartSpan(ctx context.Context, name string, attrs ...Attribute) (context.Context, Span) {
	_ = name
	_ = attrs
	t.spanCount.Add(1)
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx, nopSpan{}
}

func (t *testBackend) AddCounter(ctx context.Context, name string, value int64, attrs ...Attribute) {
	_ = ctx
	_ = name
	_ = value
	_ = attrs
	t.counterCount.Add(1)
}

func (t *testBackend) RecordHistogram(ctx context.Context, name string, value float64, attrs ...Attribute) {
	_ = ctx
	_ = name
	_ = value
	_ = attrs
	t.histogramCount.Add(1)
}

func TestMulti(t *testing.T) {
	t.Parallel()

	a := &testBackend{}
	b := &testBackend{}

	obs := Multi(a, b)
	require.NotNil(t, obs)
	require.NotNil(t, obs.Logger())

	ctx, span := obs.StartSpan(context.Background(), "test")
	require.NotNil(t, ctx)
	require.NotNil(t, span)

	obs.AddCounter(ctx, "counter", 1)
	obs.RecordHistogram(ctx, "histogram", 1)
	span.End()

	require.EqualValues(t, 1, a.spanCount.Load())
	require.EqualValues(t, 1, b.spanCount.Load())
	require.EqualValues(t, 1, a.counterCount.Load())
	require.EqualValues(t, 1, b.counterCount.Load())
	require.EqualValues(t, 1, a.histogramCount.Load())
	require.EqualValues(t, 1, b.histogramCount.Load())
}
