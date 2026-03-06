package mapping

import (
	"encoding/json"

	common "github.com/DaiYuANg/arcgo/collectionx/internal"
)

// ToJSON serializes map entries to JSON.
func (m *Map[K, V]) ToJSON() ([]byte, error) {
	return json.Marshal(m.All())
}

// MarshalJSON implements json.Marshaler.
func (m *Map[K, V]) MarshalJSON() ([]byte, error) {
	return m.ToJSON()
}

// String implements fmt.Stringer.
func (m *Map[K, V]) String() string {
	data, err := m.ToJSON()
	return common.JSONResultString(data, err, "{}")
}

// ToJSON serializes concurrent map entries to JSON.
func (m *ConcurrentMap[K, V]) ToJSON() ([]byte, error) {
	return json.Marshal(m.All())
}

// MarshalJSON implements json.Marshaler.
func (m *ConcurrentMap[K, V]) MarshalJSON() ([]byte, error) {
	return m.ToJSON()
}

// String implements fmt.Stringer.
func (m *ConcurrentMap[K, V]) String() string {
	data, err := m.ToJSON()
	return common.JSONResultString(data, err, "{}")
}

// ToJSON serializes bidirectional map entries to JSON.
func (m *BiMap[K, V]) ToJSON() ([]byte, error) {
	return json.Marshal(m.All())
}

// MarshalJSON implements json.Marshaler.
func (m *BiMap[K, V]) MarshalJSON() ([]byte, error) {
	return m.ToJSON()
}

// String implements fmt.Stringer.
func (m *BiMap[K, V]) String() string {
	data, err := m.ToJSON()
	return common.JSONResultString(data, err, "{}")
}

// ToJSON serializes ordered map entries to JSON.
func (m *OrderedMap[K, V]) ToJSON() ([]byte, error) {
	return json.Marshal(m.All())
}

// MarshalJSON implements json.Marshaler.
func (m *OrderedMap[K, V]) MarshalJSON() ([]byte, error) {
	return m.ToJSON()
}

// String implements fmt.Stringer.
func (m *OrderedMap[K, V]) String() string {
	data, err := m.ToJSON()
	return common.JSONResultString(data, err, "{}")
}

// ToJSON serializes multimap entries to JSON.
func (m *MultiMap[K, V]) ToJSON() ([]byte, error) {
	return json.Marshal(m.All())
}

// MarshalJSON implements json.Marshaler.
func (m *MultiMap[K, V]) MarshalJSON() ([]byte, error) {
	return m.ToJSON()
}

// String implements fmt.Stringer.
func (m *MultiMap[K, V]) String() string {
	data, err := m.ToJSON()
	return common.JSONResultString(data, err, "{}")
}

// ToJSON serializes concurrent multimap entries to JSON.
func (m *ConcurrentMultiMap[K, V]) ToJSON() ([]byte, error) {
	return json.Marshal(m.All())
}

// MarshalJSON implements json.Marshaler.
func (m *ConcurrentMultiMap[K, V]) MarshalJSON() ([]byte, error) {
	return m.ToJSON()
}

// String implements fmt.Stringer.
func (m *ConcurrentMultiMap[K, V]) String() string {
	data, err := m.ToJSON()
	return common.JSONResultString(data, err, "{}")
}

// ToJSON serializes table cells to JSON.
func (t *Table[R, C, V]) ToJSON() ([]byte, error) {
	return json.Marshal(t.All())
}

// MarshalJSON implements json.Marshaler.
func (t *Table[R, C, V]) MarshalJSON() ([]byte, error) {
	return t.ToJSON()
}

// String implements fmt.Stringer.
func (t *Table[R, C, V]) String() string {
	data, err := t.ToJSON()
	return common.JSONResultString(data, err, "{}")
}

// ToJSON serializes concurrent table cells to JSON.
func (t *ConcurrentTable[R, C, V]) ToJSON() ([]byte, error) {
	return json.Marshal(t.All())
}

// MarshalJSON implements json.Marshaler.
func (t *ConcurrentTable[R, C, V]) MarshalJSON() ([]byte, error) {
	return t.ToJSON()
}

// String implements fmt.Stringer.
func (t *ConcurrentTable[R, C, V]) String() string {
	data, err := t.ToJSON()
	return common.JSONResultString(data, err, "{}")
}
