package set

import (
	"sync"

	"github.com/samber/lo"
)

// ConcurrentSet is a goroutine-safe set.
// Zero value is ready to use.
type ConcurrentSet[T comparable] struct {
	mu    sync.RWMutex
	items map[T]struct{}
}

// NewConcurrentSet creates a new concurrent set.
func NewConcurrentSet[T comparable](items ...T) *ConcurrentSet[T] {
	s := &ConcurrentSet[T]{}
	s.Add(items...)
	return s
}

// Add inserts one or more items.
func (s *ConcurrentSet[T]) Add(items ...T) {
	if s == nil || len(items) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensureInitLocked(len(items))
	for _, item := range items {
		s.items[item] = struct{}{}
	}
}

// AddIfAbsent inserts one item only when it does not exist.
// Returns true when inserted, false when it already exists.
func (s *ConcurrentSet[T]) AddIfAbsent(item T) bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensureInitLocked(1)
	if _, exists := s.items[item]; exists {
		return false
	}
	s.items[item] = struct{}{}
	return true
}

// Remove deletes an item and reports whether it existed.
func (s *ConcurrentSet[T]) Remove(item T) bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.items == nil {
		return false
	}
	_, existed := s.items[item]
	if existed {
		delete(s.items, item)
	}
	return existed
}

// Contains reports whether item exists.
func (s *ConcurrentSet[T]) Contains(item T) bool {
	if s == nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.items == nil {
		return false
	}
	_, ok := s.items[item]
	return ok
}

// Len returns total item count.
func (s *ConcurrentSet[T]) Len() int {
	if s == nil {
		return 0
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// IsEmpty reports whether set has no items.
func (s *ConcurrentSet[T]) IsEmpty() bool {
	return s.Len() == 0
}

// Clear removes all items.
func (s *ConcurrentSet[T]) Clear() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	clear(s.items)
}

// Values returns a snapshot of all items.
func (s *ConcurrentSet[T]) Values() []T {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.items) == 0 {
		return nil
	}
	return lo.Keys(s.items)
}

// Range iterates a stable snapshot until fn returns false.
func (s *ConcurrentSet[T]) Range(fn func(item T) bool) {
	if s == nil || fn == nil {
		return
	}
	for _, item := range s.Values() {
		if !fn(item) {
			return
		}
	}
}

// Snapshot returns an immutable-style copy in a normal Set.
func (s *ConcurrentSet[T]) Snapshot() *Set[T] {
	out := &Set[T]{}
	if s == nil {
		return out
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.items) == 0 {
		return out
	}
	out.items = lo.Assign(map[T]struct{}{}, s.items)
	return out
}

func (s *ConcurrentSet[T]) ensureInitLocked(capacity int) {
	if s.items != nil {
		return
	}
	if capacity < 0 {
		capacity = 0
	}
	s.items = make(map[T]struct{}, capacity)
}
