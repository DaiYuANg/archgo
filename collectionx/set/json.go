package set

import (
	"encoding/json"

	common "github.com/DaiYuANg/arcgo/collectionx/internal"
)

// ToJSON serializes set values to JSON.
func (s *Set[T]) ToJSON() ([]byte, error) {
	return json.Marshal(s.Values())
}

// MarshalJSON implements json.Marshaler.
func (s *Set[T]) MarshalJSON() ([]byte, error) {
	return s.ToJSON()
}

// String implements fmt.Stringer.
func (s *Set[T]) String() string {
	data, err := s.ToJSON()
	return common.JSONResultString(data, err, "[]")
}

// ToJSON serializes concurrent set values to JSON.
func (s *ConcurrentSet[T]) ToJSON() ([]byte, error) {
	return json.Marshal(s.Values())
}

// MarshalJSON implements json.Marshaler.
func (s *ConcurrentSet[T]) MarshalJSON() ([]byte, error) {
	return s.ToJSON()
}

// String implements fmt.Stringer.
func (s *ConcurrentSet[T]) String() string {
	data, err := s.ToJSON()
	return common.JSONResultString(data, err, "[]")
}

// ToJSON serializes multiset counts to JSON.
func (s *MultiSet[T]) ToJSON() ([]byte, error) {
	return json.Marshal(s.AllCounts())
}

// MarshalJSON implements json.Marshaler.
func (s *MultiSet[T]) MarshalJSON() ([]byte, error) {
	return s.ToJSON()
}

// String implements fmt.Stringer.
func (s *MultiSet[T]) String() string {
	data, err := s.ToJSON()
	return common.JSONResultString(data, err, "{}")
}

// ToJSON serializes ordered set values to JSON.
func (s *OrderedSet[T]) ToJSON() ([]byte, error) {
	return json.Marshal(s.Values())
}

// MarshalJSON implements json.Marshaler.
func (s *OrderedSet[T]) MarshalJSON() ([]byte, error) {
	return s.ToJSON()
}

// String implements fmt.Stringer.
func (s *OrderedSet[T]) String() string {
	data, err := s.ToJSON()
	return common.JSONResultString(data, err, "[]")
}
