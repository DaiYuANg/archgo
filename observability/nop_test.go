package observability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNop(t *testing.T) {
	t.Parallel()

	obs := Nop()
	require.NotNil(t, obs)
	require.NotNil(t, obs.Logger())

	ctx, span := obs.StartSpan(context.TODO(), "test")
	require.NotNil(t, ctx)
	require.NotNil(t, span)

	obs.AddCounter(context.Background(), "counter", 1, String("result", "ok"))
	obs.RecordHistogram(context.Background(), "histogram", 1.0, String("result", "ok"))

	span.SetAttributes(String("k", "v"))
	span.RecordError(context.Canceled)
	span.End()
}
