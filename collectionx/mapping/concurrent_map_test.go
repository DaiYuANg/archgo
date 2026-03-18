package mapping

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConcurrentMap_ParallelSet(t *testing.T) {
	t.Parallel()

	var m ConcurrentMap[int, int]

	const workers = 20
	const each = 200

	var wg sync.WaitGroup
	wg.Add(workers)

	for worker := 0; worker < workers; worker++ {
		worker := worker
		go func() {
			defer wg.Done()
			base := worker * each
			for i := 0; i < each; i++ {
				m.Set(base+i, i)
			}
		}()
	}

	wg.Wait()
	require.Equal(t, workers*each, m.Len())
}

func TestConcurrentMap_GetOrStore(t *testing.T) {
	t.Parallel()

	var m ConcurrentMap[string, int]

	value, loaded := m.GetOrStore("a", 1)
	require.False(t, loaded)
	require.Equal(t, 1, value)

	value, loaded = m.GetOrStore("a", 9)
	require.True(t, loaded)
	require.Equal(t, 1, value)
}

func TestConcurrentMap_LoadAndDelete(t *testing.T) {
	t.Parallel()

	var m ConcurrentMap[string, string]
	m.Set("k", "v")

	value, ok := m.LoadAndDelete("k")
	require.True(t, ok)
	require.Equal(t, "v", value)

	_, ok = m.Get("k")
	require.False(t, ok)
}

func TestConcurrentMap_OptionAPIs(t *testing.T) {
	t.Parallel()

	var m ConcurrentMap[string, int]
	m.Set("x", 42)

	opt := m.GetOption("x")
	require.True(t, opt.IsPresent())
	value, ok := opt.Get()
	require.True(t, ok)
	require.Equal(t, 42, value)

	deleted := m.LoadAndDeleteOption("x")
	require.True(t, deleted.IsPresent())
	deletedValue, ok := deleted.Get()
	require.True(t, ok)
	require.Equal(t, 42, deletedValue)

	require.True(t, m.GetOption("x").IsAbsent())
}

func TestConcurrentMap_Range(t *testing.T) {
	t.Parallel()

	m := NewConcurrentMap[string, int]()
	for i := 0; i < 10; i++ {
		m.Set(strconv.Itoa(i), i)
	}

	visited := 0
	m.Range(func(key string, value int) bool {
		visited++
		return visited < 3
	})
	require.Equal(t, 3, visited)
}

func TestNewConcurrentMapWithCapacity(t *testing.T) {
	t.Parallel()

	m := NewConcurrentMapWithCapacity[string, int](8)
	m.Set("a", 1)

	value, ok := m.Get("a")
	require.True(t, ok)
	require.Equal(t, 1, value)
}
