package set

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderedSet_OrderAndDedupe(t *testing.T) {
	t.Parallel()

	var s OrderedSet[int]
	s.Add(1, 2, 2, 3, 1)

	require.Equal(t, []int{1, 2, 3}, s.Values())
	require.True(t, s.Contains(2))
}

func TestOrderedSet_RemoveReindex(t *testing.T) {
	t.Parallel()

	s := NewOrderedSet("a", "b", "c")
	require.True(t, s.Remove("b"))
	require.Equal(t, []string{"a", "c"}, s.Values())

	item, ok := s.At(1)
	require.True(t, ok)
	require.Equal(t, "c", item)
}

func TestNewOrderedSetWithCapacity(t *testing.T) {
	t.Parallel()

	s := NewOrderedSetWithCapacity(8, "a", "b", "a", "c")

	require.Equal(t, []string{"a", "b", "c"}, s.Values())
}
