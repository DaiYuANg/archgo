package mapping

import (
	"sync"

	"github.com/samber/lo"
	"github.com/samber/mo"
)

// ConcurrentMap is a goroutine-safe strongly-typed map.
// Zero value is ready to use.
type ConcurrentMap[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]V
}

// NewConcurrentMap creates an empty concurrent map.
func NewConcurrentMap[K comparable, V any]() *ConcurrentMap[K, V] {
	return &ConcurrentMap[K, V]{
		items: make(map[K]V),
	}
}

// Set puts a key-value pair.
func (m *ConcurrentMap[K, V]) Set(key K, value V) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureInitLocked(1)
	m.items[key] = value
}

// SetAll copies all entries from source.
func (m *ConcurrentMap[K, V]) SetAll(source map[K]V) {
	if m == nil || len(source) == 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureInitLocked(len(source))
	m.items = lo.Assign(m.items, source)
}

// Get returns the value for key.
func (m *ConcurrentMap[K, V]) Get(key K) (V, bool) {
	var zero V
	if m == nil {
		return zero, false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.items == nil {
		return zero, false
	}
	value, ok := m.items[key]
	return value, ok
}

// GetOption returns value for key as mo.Option.
func (m *ConcurrentMap[K, V]) GetOption(key K) mo.Option[V] {
	value, ok := m.Get(key)
	if !ok {
		return mo.None[V]()
	}
	return mo.Some(value)
}

// GetOrDefault returns value for key or fallback when key does not exist.
func (m *ConcurrentMap[K, V]) GetOrDefault(key K, fallback V) V {
	value, ok := m.Get(key)
	if !ok {
		return fallback
	}
	return value
}

// GetOrStore returns existing value if key exists; otherwise stores and returns value.
// loaded is true when key already exists.
func (m *ConcurrentMap[K, V]) GetOrStore(key K, value V) (actual V, loaded bool) {
	if m == nil {
		return value, false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureInitLocked(1)

	if existing, ok := m.items[key]; ok {
		return existing, true
	}
	m.items[key] = value
	return value, false
}

// Delete removes key and reports whether it existed.
func (m *ConcurrentMap[K, V]) Delete(key K) bool {
	if m == nil {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.items == nil {
		return false
	}
	_, existed := m.items[key]
	if existed {
		delete(m.items, key)
	}
	return existed
}

// LoadAndDelete removes key and returns previous value.
func (m *ConcurrentMap[K, V]) LoadAndDelete(key K) (V, bool) {
	var zero V
	if m == nil {
		return zero, false
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.items == nil {
		return zero, false
	}
	value, ok := m.items[key]
	if ok {
		delete(m.items, key)
	}
	return value, ok
}

// LoadAndDeleteOption removes key and returns previous value as mo.Option.
func (m *ConcurrentMap[K, V]) LoadAndDeleteOption(key K) mo.Option[V] {
	value, ok := m.LoadAndDelete(key)
	if !ok {
		return mo.None[V]()
	}
	return mo.Some(value)
}

// Len returns total entry count.
func (m *ConcurrentMap[K, V]) Len() int {
	if m == nil {
		return 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.items)
}

// IsEmpty reports whether map has no entries.
func (m *ConcurrentMap[K, V]) IsEmpty() bool {
	return m.Len() == 0
}

// Clear removes all entries.
func (m *ConcurrentMap[K, V]) Clear() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	clear(m.items)
}

// Keys returns a snapshot of keys.
func (m *ConcurrentMap[K, V]) Keys() []K {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.items) == 0 {
		return nil
	}
	return lo.Keys(m.items)
}

// Values returns a snapshot of values.
func (m *ConcurrentMap[K, V]) Values() []V {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.items) == 0 {
		return nil
	}
	return lo.Values(m.items)
}

// All returns a copied built-in map.
func (m *ConcurrentMap[K, V]) All() map[K]V {
	if m == nil {
		return map[K]V{}
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.items) == 0 {
		return map[K]V{}
	}
	return lo.Assign(map[K]V{}, m.items)
}

// Range iterates a stable snapshot until fn returns false.
func (m *ConcurrentMap[K, V]) Range(fn func(key K, value V) bool) {
	if m == nil || fn == nil {
		return
	}
	for key, value := range m.All() {
		if !fn(key, value) {
			return
		}
	}
}

func (m *ConcurrentMap[K, V]) ensureInitLocked(capacity int) {
	if m.items != nil {
		return
	}
	if capacity < 0 {
		capacity = 0
	}
	m.items = make(map[K]V, capacity)
}
