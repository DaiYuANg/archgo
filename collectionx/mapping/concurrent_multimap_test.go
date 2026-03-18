package mapping

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConcurrentMultiMap_ParallelPut(t *testing.T) {
	t.Parallel()

	var m ConcurrentMultiMap[int, int]

	const workers = 16
	const each = 120

	var wg sync.WaitGroup
	wg.Add(workers)

	for worker := 0; worker < workers; worker++ {
		worker := worker
		go func() {
			defer wg.Done()
			for i := 0; i < each; i++ {
				m.Put(worker, i)
			}
		}()
	}

	wg.Wait()
	require.Equal(t, workers, m.Len())
	require.Equal(t, workers*each, m.ValueCount())
}

func TestConcurrentMultiMap_OptionAndSnapshot(t *testing.T) {
	t.Parallel()

	var m ConcurrentMultiMap[string, int]
	m.PutAll("a", 1, 2, 3)

	opt := m.GetOption("a")
	require.True(t, opt.IsPresent())
	values, ok := opt.Get()
	require.True(t, ok)
	require.Equal(t, []int{1, 2, 3}, values)

	snapshot := m.Snapshot()
	m.Put("a", 4)
	require.Equal(t, []int{1, 2, 3}, snapshot.Get("a"))
}

func TestNewConcurrentMultiMapWithCapacity(t *testing.T) {
	t.Parallel()

	m := NewConcurrentMultiMapWithCapacity[string, int](8)
	m.PutAll("a", 1, 2)

	require.Equal(t, 1, m.Len())
	require.Equal(t, 2, m.ValueCount())
}
