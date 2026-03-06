package mapping

import (
	"slices"
	"sync"

	"github.com/samber/lo"
	"github.com/samber/mo"
)

// ConcurrentMultiMap is a goroutine-safe multimap.
// Zero value is ready to use.
type ConcurrentMultiMap[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K][]V
}

// NewConcurrentMultiMap creates an empty concurrent multimap.
func NewConcurrentMultiMap[K comparable, V any]() *ConcurrentMultiMap[K, V] {
	return &ConcurrentMultiMap[K, V]{
		items: make(map[K][]V),
	}
}

// Put appends one value for key.
func (m *ConcurrentMultiMap[K, V]) Put(key K, value V) {
	m.PutAll(key, value)
}

// PutAll appends values for key.
func (m *ConcurrentMultiMap[K, V]) PutAll(key K, values ...V) {
	if m == nil || len(values) == 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureInitLocked()
	m.items[key] = append(m.items[key], values...)
}

// Set replaces all values for key.
// Passing no values removes the key.
func (m *ConcurrentMultiMap[K, V]) Set(key K, values ...V) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureInitLocked()
	if len(values) == 0 {
		delete(m.items, key)
		return
	}
	m.items[key] = slices.Clone(values)
}

// Get returns a copy of values for key.
func (m *ConcurrentMultiMap[K, V]) Get(key K) []V {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	values, ok := m.items[key]
	if !ok || len(values) == 0 {
		return nil
	}
	return slices.Clone(values)
}

// GetOption returns values for key as mo.Option.
func (m *ConcurrentMultiMap[K, V]) GetOption(key K) mo.Option[[]V] {
	values := m.Get(key)
	if len(values) == 0 {
		return mo.None[[]V]()
	}
	return mo.Some(values)
}

// Delete removes all values for key.
func (m *ConcurrentMultiMap[K, V]) Delete(key K) bool {
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

// DeleteValueIf removes values matching predicate under key and returns removed count.
func (m *ConcurrentMultiMap[K, V]) DeleteValueIf(key K, predicate func(value V) bool) int {
	if m == nil || predicate == nil {
		return 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.items == nil {
		return 0
	}
	values, ok := m.items[key]
	if !ok || len(values) == 0 {
		return 0
	}

	next := lo.Filter(values, func(item V, _ int) bool {
		return !predicate(item)
	})
	removed := len(values) - len(next)
	if len(next) == 0 {
		delete(m.items, key)
	} else {
		m.items[key] = next
	}
	return removed
}

// ContainsKey reports whether key exists.
func (m *ConcurrentMultiMap[K, V]) ContainsKey(key K) bool {
	if m == nil {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.items == nil {
		return false
	}
	_, ok := m.items[key]
	return ok
}

// Len returns key count.
func (m *ConcurrentMultiMap[K, V]) Len() int {
	if m == nil {
		return 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.items)
}

// ValueCount returns total stored value count.
func (m *ConcurrentMultiMap[K, V]) ValueCount() int {
	if m == nil {
		return 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	return lo.SumBy(lo.Values(m.items), func(values []V) int {
		return len(values)
	})
}

// IsEmpty reports whether map has no keys.
func (m *ConcurrentMultiMap[K, V]) IsEmpty() bool {
	return m.Len() == 0
}

// Clear removes all entries.
func (m *ConcurrentMultiMap[K, V]) Clear() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	clear(m.items)
}

// Keys returns all keys.
func (m *ConcurrentMultiMap[K, V]) Keys() []K {
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

// All returns a deep-copied built-in map.
func (m *ConcurrentMultiMap[K, V]) All() map[K][]V {
	if m == nil {
		return map[K][]V{}
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.items) == 0 {
		return map[K][]V{}
	}
	out := make(map[K][]V, len(m.items))
	for key, values := range m.items {
		out[key] = slices.Clone(values)
	}
	return out
}

// Snapshot returns an immutable-style copy in a normal MultiMap.
func (m *ConcurrentMultiMap[K, V]) Snapshot() *MultiMap[K, V] {
	return NewMultiMapFromAll(m.All())
}

// Range iterates key-values snapshots until fn returns false.
func (m *ConcurrentMultiMap[K, V]) Range(fn func(key K, values []V) bool) {
	if m == nil || fn == nil {
		return
	}
	for key, values := range m.All() {
		if !fn(key, values) {
			return
		}
	}
}

func (m *ConcurrentMultiMap[K, V]) ensureInitLocked() {
	if m.items == nil {
		m.items = make(map[K][]V)
	}
}

// NewMultiMapFromAll creates a multimap from a built-in deep map.
func NewMultiMapFromAll[K comparable, V any](source map[K][]V) *MultiMap[K, V] {
	out := NewMultiMap[K, V]()
	for key, values := range source {
		out.Set(key, values...)
	}
	return out
}
