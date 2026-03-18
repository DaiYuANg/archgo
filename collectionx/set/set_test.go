package set

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSet_ZeroValueAndBasicOps(t *testing.T) {
	t.Parallel()

	var s Set[int]

	s.Add(1, 2, 2, 3)

	require.Equal(t, 3, s.Len())
	require.True(t, s.Contains(1))
	require.False(t, s.Contains(9))

	require.True(t, s.Remove(2))
	require.False(t, s.Remove(2))
	require.Equal(t, 2, s.Len())

	s.Clear()
	require.True(t, s.IsEmpty())
}

func TestSet_MathOperations(t *testing.T) {
	t.Parallel()

	left := NewSet(1, 2, 3)
	right := NewSet(3, 4, 5)

	require.ElementsMatch(t, []int{1, 2, 3, 4, 5}, left.Union(right).Values())
	require.ElementsMatch(t, []int{3}, left.Intersect(right).Values())
	require.ElementsMatch(t, []int{1, 2}, left.Difference(right).Values())
}

func TestSet_RangeStop(t *testing.T) {
	t.Parallel()

	s := NewSet("a", "b", "c")
	visited := 0

	s.Range(func(item string) bool {
		visited++
		return item != ""
	})

	require.Equal(t, 3, visited)

	visited = 0
	s.Range(func(item string) bool {
		visited++
		return false
	})
	require.Equal(t, 1, visited)
}

func TestSet_Merge(t *testing.T) {
	t.Parallel()

	left := NewSet(1, 2)
	right := NewSet(2, 3)

	left.Merge(right).MergeSlice([]int{3, 4, 5})
	require.ElementsMatch(t, []int{1, 2, 3, 4, 5}, left.Values())
}

func TestNewSetWithCapacity(t *testing.T) {
	t.Parallel()

	s := NewSetWithCapacity(8, 1, 2, 2, 3)

	require.Equal(t, 3, s.Len())
	require.True(t, s.Contains(1))
	require.True(t, s.Contains(2))
	require.True(t, s.Contains(3))
}
