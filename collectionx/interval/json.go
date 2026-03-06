package interval

import (
	"encoding/json"

	common "github.com/DaiYuANg/arcgo/collectionx/internal"
)

// ToJSON serializes normalized ranges to JSON.
func (s *RangeSet[T]) ToJSON() ([]byte, error) {
	return json.Marshal(s.Ranges())
}

// MarshalJSON implements json.Marshaler.
func (s *RangeSet[T]) MarshalJSON() ([]byte, error) {
	return s.ToJSON()
}

// String implements fmt.Stringer.
func (s *RangeSet[T]) String() string {
	data, err := s.ToJSON()
	return common.JSONResultString(data, err, "[]")
}

// ToJSON serializes range-map entries to JSON.
func (m *RangeMap[T, V]) ToJSON() ([]byte, error) {
	return json.Marshal(m.Entries())
}

// MarshalJSON implements json.Marshaler.
func (m *RangeMap[T, V]) MarshalJSON() ([]byte, error) {
	return m.ToJSON()
}

// String implements fmt.Stringer.
func (m *RangeMap[T, V]) String() string {
	data, err := m.ToJSON()
	return common.JSONResultString(data, err, "[]")
}
