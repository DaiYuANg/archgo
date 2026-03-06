package set

import "github.com/samber/lo"

// Set is a generic hash set based on map[T]struct{}.
// Zero value is ready to use.
type Set[T comparable] struct {
	items map[T]struct{}
}

// NewSet creates a new set and fills it with optional items.
func NewSet[T comparable](items ...T) *Set[T] {
	s := &Set[T]{}
	s.Add(items...)
	return s
}

// Add inserts one or more items.
func (s *Set[T]) Add(items ...T) {
	if s == nil || len(items) == 0 {
		return
	}
	s.ensureInit()
	for _, item := range items {
		s.items[item] = struct{}{}
	}
}

// Remove deletes an item and reports whether it existed.
func (s *Set[T]) Remove(item T) bool {
	if s == nil || s.items == nil {
		return false
	}
	_, existed := s.items[item]
	if existed {
		delete(s.items, item)
	}
	return existed
}

// Contains reports whether item exists.
func (s *Set[T]) Contains(item T) bool {
	if s == nil || s.items == nil {
		return false
	}
	_, ok := s.items[item]
	return ok
}

// Len returns total item count.
func (s *Set[T]) Len() int {
	if s == nil {
		return 0
	}
	return len(s.items)
}

// IsEmpty reports whether the set has no items.
func (s *Set[T]) IsEmpty() bool {
	return s.Len() == 0
}

// Clear removes all items.
func (s *Set[T]) Clear() {
	if s == nil {
		return
	}
	clear(s.items)
}

// Values returns all items as a slice.
func (s *Set[T]) Values() []T {
	if s == nil || len(s.items) == 0 {
		return nil
	}
	return lo.Keys(s.items)
}

// Range iterates all items until fn returns false.
func (s *Set[T]) Range(fn func(item T) bool) {
	if s == nil || fn == nil {
		return
	}
	for item := range s.items {
		if !fn(item) {
			return
		}
	}
}

// Clone returns a shallow copy.
func (s *Set[T]) Clone() *Set[T] {
	out := &Set[T]{}
	if s == nil || len(s.items) == 0 {
		return out
	}
	out.items = lo.Assign(map[T]struct{}{}, s.items)
	return out
}

// Union returns a new set that contains items from both sets.
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	out := s.Clone()
	if other == nil || len(other.items) == 0 {
		return out
	}
	out.ensureInit()
	for item := range other.items {
		out.items[item] = struct{}{}
	}
	return out
}

// Intersect returns a new set that contains shared items.
func (s *Set[T]) Intersect(other *Set[T]) *Set[T] {
	out := &Set[T]{}
	if s == nil || other == nil || len(s.items) == 0 || len(other.items) == 0 {
		return out
	}

	left := s.items
	right := other.items
	if len(left) > len(right) {
		left, right = right, left
	}

	out.items = make(map[T]struct{})
	for item := range left {
		if _, ok := right[item]; ok {
			out.items[item] = struct{}{}
		}
	}
	return out
}

// Difference returns a new set with items in s but not in other.
func (s *Set[T]) Difference(other *Set[T]) *Set[T] {
	out := &Set[T]{}
	if s == nil || len(s.items) == 0 {
		return out
	}
	out.items = make(map[T]struct{}, len(s.items))

	if other == nil || len(other.items) == 0 {
		for item := range s.items {
			out.items[item] = struct{}{}
		}
		return out
	}

	for item := range s.items {
		if _, ok := other.items[item]; !ok {
			out.items[item] = struct{}{}
		}
	}
	return out
}

func (s *Set[T]) ensureInit() {
	if s.items == nil {
		s.items = make(map[T]struct{})
	}
}
