package set

import (
	"slices"

	"github.com/samber/lo"
)

// OrderedSet keeps insertion order of unique items.
// Zero value is ready to use.
type OrderedSet[T comparable] struct {
	order []T
	items map[T]struct{}
	index map[T]int
}

// NewOrderedSet creates an ordered set with optional items.
func NewOrderedSet[T comparable](items ...T) *OrderedSet[T] {
	s := &OrderedSet[T]{}
	s.Add(items...)
	return s
}

// Add inserts one or more items.
func (s *OrderedSet[T]) Add(items ...T) {
	if s == nil || len(items) == 0 {
		return
	}
	s.ensureInit()

	lo.ForEach(items, func(item T, _ int) {
		if _, exists := s.items[item]; exists {
			return
		}
		s.order = append(s.order, item)
		s.items[item] = struct{}{}
		s.index[item] = len(s.order) - 1
	})
}

// Remove deletes item and reports whether it existed.
func (s *OrderedSet[T]) Remove(item T) bool {
	if s == nil || s.items == nil {
		return false
	}
	pos, ok := s.index[item]
	if !ok {
		return false
	}

	delete(s.items, item)
	delete(s.index, item)

	copy(s.order[pos:], s.order[pos+1:])
	s.order = s.order[:len(s.order)-1]
	for i := pos; i < len(s.order); i++ {
		s.index[s.order[i]] = i
	}
	return true
}

// Contains reports whether item exists.
func (s *OrderedSet[T]) Contains(item T) bool {
	if s == nil || s.items == nil {
		return false
	}
	_, ok := s.items[item]
	return ok
}

// Len returns item count.
func (s *OrderedSet[T]) Len() int {
	if s == nil {
		return 0
	}
	return len(s.order)
}

// IsEmpty reports whether set has no items.
func (s *OrderedSet[T]) IsEmpty() bool {
	return s.Len() == 0
}

// Clear removes all items.
func (s *OrderedSet[T]) Clear() {
	if s == nil {
		return
	}
	s.order = nil
	clear(s.items)
	clear(s.index)
}

// Values returns items in insertion order.
func (s *OrderedSet[T]) Values() []T {
	if s == nil || len(s.order) == 0 {
		return nil
	}
	return slices.Clone(s.order)
}

// At returns item at insertion index.
func (s *OrderedSet[T]) At(pos int) (T, bool) {
	var zero T
	if s == nil || pos < 0 || pos >= len(s.order) {
		return zero, false
	}
	return s.order[pos], true
}

// Range iterates items in insertion order until fn returns false.
func (s *OrderedSet[T]) Range(fn func(item T) bool) {
	if s == nil || fn == nil {
		return
	}
	for _, item := range s.order {
		if !fn(item) {
			return
		}
	}
}

// Clone returns a shallow copy.
func (s *OrderedSet[T]) Clone() *OrderedSet[T] {
	out := &OrderedSet[T]{}
	if s == nil {
		return out
	}
	out.order = slices.Clone(s.order)
	out.items = lo.Assign(map[T]struct{}{}, s.items)
	out.index = lo.Assign(map[T]int{}, s.index)
	return out
}

func (s *OrderedSet[T]) ensureInit() {
	if s.order == nil {
		s.order = make([]T, 0)
	}
	if s.items == nil {
		s.items = make(map[T]struct{})
	}
	if s.index == nil {
		s.index = make(map[T]int)
	}
}
