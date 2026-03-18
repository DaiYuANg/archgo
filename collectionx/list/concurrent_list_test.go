package list

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConcurrentList_ParallelAdd(t *testing.T) {
	t.Parallel()

	var l ConcurrentList[int]

	const workers = 24
	const each = 150

	var wg sync.WaitGroup
	wg.Add(workers)

	for worker := 0; worker < workers; worker++ {
		worker := worker
		go func() {
			defer wg.Done()
			base := worker * each
			for i := 0; i < each; i++ {
				l.Add(base + i)
			}
		}()
	}

	wg.Wait()
	require.Equal(t, workers*each, l.Len())
}

func TestConcurrentList_InsertRemoveAndSnapshot(t *testing.T) {
	t.Parallel()

	l := NewConcurrentList(1, 3)
	require.True(t, l.AddAt(1, 2))
	require.Equal(t, []int{1, 2, 3}, l.Values())

	removed, ok := l.RemoveAt(1)
	require.True(t, ok)
	require.Equal(t, 2, removed)

	snapshot := l.Snapshot()
	l.Add(9)
	require.Equal(t, []int{1, 3}, snapshot.Values())
}

func TestConcurrentList_OptionAPIs(t *testing.T) {
	t.Parallel()

	var l ConcurrentList[string]
	l.Add("a", "b")

	opt := l.GetOption(0)
	require.True(t, opt.IsPresent())
	value, ok := opt.Get()
	require.True(t, ok)
	require.Equal(t, "a", value)

	removed := l.RemoveAtOption(1)
	require.True(t, removed.IsPresent())
	removedValue, ok := removed.Get()
	require.True(t, ok)
	require.Equal(t, "b", removedValue)

	require.True(t, l.GetOption(99).IsAbsent())
}

func TestConcurrentList_Merge(t *testing.T) {
	t.Parallel()

	left := NewConcurrentList(1, 2)
	right := NewList(3, 4)
	otherConcurrent := NewConcurrentList(5, 6)

	left.Merge(right).MergeConcurrent(otherConcurrent).MergeSlice([]int{7, 8})
	require.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8}, left.Values())
}

func TestNewConcurrentListWithCapacity(t *testing.T) {
	t.Parallel()

	l := NewConcurrentListWithCapacity[int](8, 1, 2, 3)

	require.Equal(t, []int{1, 2, 3}, l.Values())
	require.NotNil(t, l.core)
	require.GreaterOrEqual(t, cap(l.core.items), 8)
}
