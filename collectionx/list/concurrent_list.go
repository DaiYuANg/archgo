package list

import (
	"sync"

	"github.com/samber/mo"
)

// ConcurrentList is a goroutine-safe strongly-typed list.
// Zero value is ready to use.
type ConcurrentList[T any] struct {
	mu   sync.RWMutex
	core *List[T]
}

// NewConcurrentList creates a list and copies optional items.
func NewConcurrentList[T any](items ...T) *ConcurrentList[T] {
	return NewConcurrentListWithCapacity(len(items), items...)
}

// NewConcurrentListWithCapacity creates a list with preallocated capacity and optional items.
func NewConcurrentListWithCapacity[T any](capacity int, items ...T) *ConcurrentList[T] {
	if capacity < len(items) {
		capacity = len(items)
	}
	if capacity <= 0 {
		return &ConcurrentList[T]{}
	}
	return &ConcurrentList[T]{
		core: NewListWithCapacity(capacity, items...),
	}
}

// Add appends one or more items.
func (l *ConcurrentList[T]) Add(items ...T) {
	if l == nil || len(items) == 0 {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ensureInitLocked()
	l.core.Add(items...)
}

// Merge appends all items from a normal list.
func (l *ConcurrentList[T]) Merge(other *List[T]) *ConcurrentList[T] {
	if l == nil {
		return nil
	}
	if other == nil {
		return l
	}
	l.Add(other.Values()...)
	return l
}

// MergeConcurrent appends all items from another concurrent list snapshot.
func (l *ConcurrentList[T]) MergeConcurrent(other *ConcurrentList[T]) *ConcurrentList[T] {
	if l == nil {
		return nil
	}
	if other == nil {
		return l
	}
	l.Add(other.Values()...)
	return l
}

// MergeSlice appends all items from a slice.
func (l *ConcurrentList[T]) MergeSlice(items []T) *ConcurrentList[T] {
	if l == nil {
		return nil
	}
	l.Add(items...)
	return l
}

// AddAt inserts one item at index. index == Len() is allowed.
func (l *ConcurrentList[T]) AddAt(index int, item T) bool {
	return l.AddAllAt(index, item)
}

// AddAllAt inserts items at index while preserving order.
func (l *ConcurrentList[T]) AddAllAt(index int, items ...T) bool {
	if l == nil {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ensureInitLocked()
	return l.core.AddAllAt(index, items...)
}

// Get returns item at index.
func (l *ConcurrentList[T]) Get(index int) (T, bool) {
	var zero T
	if l == nil {
		return zero, false
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.core == nil {
		return zero, false
	}
	return l.core.Get(index)
}

// GetOption returns item at index as mo.Option.
func (l *ConcurrentList[T]) GetOption(index int) mo.Option[T] {
	value, ok := l.Get(index)
	if !ok {
		return mo.None[T]()
	}
	return mo.Some(value)
}

// Set replaces item at index.
func (l *ConcurrentList[T]) Set(index int, item T) bool {
	if l == nil {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.core == nil {
		return false
	}
	return l.core.Set(index, item)
}

// SetAll applies mapper to each item and replaces all items in-place.
// Returns updated item count.
func (l *ConcurrentList[T]) SetAll(mapper func(item T) T) int {
	if l == nil || mapper == nil {
		return 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.core == nil {
		return 0
	}
	return l.core.SetAll(mapper)
}

// SetAllIndexed applies mapper(index, item) to each item and replaces all items in-place.
// Returns updated item count.
func (l *ConcurrentList[T]) SetAllIndexed(mapper func(index int, item T) T) int {
	if l == nil || mapper == nil {
		return 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.core == nil {
		return 0
	}
	return l.core.SetAllIndexed(mapper)
}

// RemoveAt removes and returns item at index.
func (l *ConcurrentList[T]) RemoveAt(index int) (T, bool) {
	var zero T
	if l == nil {
		return zero, false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.core == nil {
		return zero, false
	}
	return l.core.RemoveAt(index)
}

// RemoveAtOption removes item at index and returns it as mo.Option.
func (l *ConcurrentList[T]) RemoveAtOption(index int) mo.Option[T] {
	value, ok := l.RemoveAt(index)
	if !ok {
		return mo.None[T]()
	}
	return mo.Some(value)
}

// RemoveIf removes all items matched by predicate and returns removed count.
func (l *ConcurrentList[T]) RemoveIf(predicate func(item T) bool) int {
	if l == nil || predicate == nil {
		return 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.core == nil {
		return 0
	}
	return l.core.RemoveIf(predicate)
}

// Len returns item count.
func (l *ConcurrentList[T]) Len() int {
	if l == nil {
		return 0
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.core == nil {
		return 0
	}
	return l.core.Len()
}

// IsEmpty reports whether list has no items.
func (l *ConcurrentList[T]) IsEmpty() bool {
	return l.Len() == 0
}

// Clear removes all items.
func (l *ConcurrentList[T]) Clear() {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.core == nil {
		return
	}
	l.core.Clear()
}

// Values returns a snapshot of items.
func (l *ConcurrentList[T]) Values() []T {
	if l == nil {
		return nil
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.core == nil {
		return nil
	}
	return l.core.Values()
}

// Range iterates a stable snapshot from left to right until fn returns false.
func (l *ConcurrentList[T]) Range(fn func(index int, item T) bool) {
	if l == nil || fn == nil {
		return
	}
	for index, item := range l.Values() {
		if !fn(index, item) {
			return
		}
	}
}

// Snapshot returns an immutable-style copy in a normal List.
func (l *ConcurrentList[T]) Snapshot() *List[T] {
	if l == nil {
		return NewList[T]()
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.core == nil {
		return NewList[T]()
	}
	return l.core.Clone()
}

func (l *ConcurrentList[T]) ensureInitLocked() {
	if l.core == nil {
		l.core = NewList[T]()
	}
}
