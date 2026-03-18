package set

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultiSet_BasicOps(t *testing.T) {
	t.Parallel()

	var s MultiSet[string]
	s.Add("a", "a", "b")
	s.AddN("c", 3)

	require.Equal(t, 6, s.Len())
	require.Equal(t, 3, s.UniqueLen())
	require.Equal(t, 2, s.Count("a"))
	require.Equal(t, 3, s.Count("c"))

	require.True(t, s.Remove("a"))
	require.Equal(t, 1, s.Count("a"))
	require.Equal(t, 2, s.RemoveN("c", 2))
	require.Equal(t, 1, s.Count("c"))
}

func TestMultiSet_CopySemantics(t *testing.T) {
	t.Parallel()

	s := NewMultiSet(1, 1, 2, 3)
	all := s.AllCounts()
	all[1] = 99
	require.Equal(t, 2, s.Count(1))
}

func TestNewMultiSetWithCapacity(t *testing.T) {
	t.Parallel()

	s := NewMultiSetWithCapacity(8, "a", "a", "b")

	require.Equal(t, 3, s.Len())
	require.Equal(t, 2, s.Count("a"))
	require.Equal(t, 1, s.Count("b"))
}
