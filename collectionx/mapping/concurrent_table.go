package mapping

import (
	"sync"

	"github.com/samber/lo"
	"github.com/samber/mo"
)

// ConcurrentTable is a goroutine-safe 2D key-value structure.
// Zero value is ready to use.
type ConcurrentTable[R comparable, C comparable, V any] struct {
	mu   sync.RWMutex
	data map[R]map[C]V
}

// NewConcurrentTable creates an empty concurrent table.
func NewConcurrentTable[R comparable, C comparable, V any]() *ConcurrentTable[R, C, V] {
	return &ConcurrentTable[R, C, V]{
		data: make(map[R]map[C]V),
	}
}

// Put sets value at (rowKey, columnKey).
func (t *ConcurrentTable[R, C, V]) Put(rowKey R, columnKey C, value V) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ensureInitLocked()

	row := t.ensureRowLocked(rowKey)
	row[columnKey] = value
}

// Get returns value at (rowKey, columnKey).
func (t *ConcurrentTable[R, C, V]) Get(rowKey R, columnKey C) (V, bool) {
	var zero V
	if t == nil {
		return zero, false
	}
	t.mu.RLock()
	defer t.mu.RUnlock()

	row, ok := t.data[rowKey]
	if !ok {
		return zero, false
	}
	value, ok := row[columnKey]
	return value, ok
}

// GetOption returns value at (rowKey, columnKey) as mo.Option.
func (t *ConcurrentTable[R, C, V]) GetOption(rowKey R, columnKey C) mo.Option[V] {
	value, ok := t.Get(rowKey, columnKey)
	if !ok {
		return mo.None[V]()
	}
	return mo.Some(value)
}

// SetRow replaces one entire row.
// Empty rowValues removes the row.
func (t *ConcurrentTable[R, C, V]) SetRow(rowKey R, rowValues map[C]V) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ensureInitLocked()

	if len(rowValues) == 0 {
		delete(t.data, rowKey)
		return
	}
	t.data[rowKey] = lo.Assign(map[C]V{}, rowValues)
}

// Row returns one row as a copied map.
func (t *ConcurrentTable[R, C, V]) Row(rowKey R) map[C]V {
	if t == nil {
		return map[C]V{}
	}
	t.mu.RLock()
	defer t.mu.RUnlock()

	row, ok := t.data[rowKey]
	if !ok || len(row) == 0 {
		return map[C]V{}
	}
	return lo.Assign(map[C]V{}, row)
}

// Column returns one column as a copied map[row]value.
func (t *ConcurrentTable[R, C, V]) Column(columnKey C) map[R]V {
	if t == nil {
		return map[R]V{}
	}
	t.mu.RLock()
	defer t.mu.RUnlock()

	out := make(map[R]V)
	for rowKey, row := range t.data {
		if value, ok := row[columnKey]; ok {
			out[rowKey] = value
		}
	}
	return out
}

// Delete removes one cell and reports whether it existed.
func (t *ConcurrentTable[R, C, V]) Delete(rowKey R, columnKey C) bool {
	if t == nil {
		return false
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	row, ok := t.data[rowKey]
	if !ok {
		return false
	}
	_, existed := row[columnKey]
	if !existed {
		return false
	}

	delete(row, columnKey)
	if len(row) == 0 {
		delete(t.data, rowKey)
	}
	return true
}

// DeleteRow removes one row and reports whether it existed.
func (t *ConcurrentTable[R, C, V]) DeleteRow(rowKey R) bool {
	if t == nil {
		return false
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	_, existed := t.data[rowKey]
	if existed {
		delete(t.data, rowKey)
	}
	return existed
}

// DeleteColumn removes one column from all rows and returns removed cell count.
func (t *ConcurrentTable[R, C, V]) DeleteColumn(columnKey C) int {
	if t == nil {
		return 0
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	removed := 0
	for rowKey, row := range t.data {
		if _, ok := row[columnKey]; ok {
			delete(row, columnKey)
			removed++
		}
		if len(row) == 0 {
			delete(t.data, rowKey)
		}
	}
	return removed
}

// Has reports whether cell exists.
func (t *ConcurrentTable[R, C, V]) Has(rowKey R, columnKey C) bool {
	_, ok := t.Get(rowKey, columnKey)
	return ok
}

// RowCount returns total row count.
func (t *ConcurrentTable[R, C, V]) RowCount() int {
	if t == nil {
		return 0
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.data)
}

// Len returns total cell count.
func (t *ConcurrentTable[R, C, V]) Len() int {
	if t == nil {
		return 0
	}
	t.mu.RLock()
	defer t.mu.RUnlock()

	return lo.SumBy(lo.Values(t.data), func(row map[C]V) int {
		return len(row)
	})
}

// IsEmpty reports whether table has no cells.
func (t *ConcurrentTable[R, C, V]) IsEmpty() bool {
	return t.Len() == 0
}

// Clear removes all cells.
func (t *ConcurrentTable[R, C, V]) Clear() {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	clear(t.data)
}

// RowKeys returns all row keys.
func (t *ConcurrentTable[R, C, V]) RowKeys() []R {
	if t == nil {
		return nil
	}
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.data) == 0 {
		return nil
	}
	return lo.Keys(t.data)
}

// ColumnKeys returns all unique column keys.
func (t *ConcurrentTable[R, C, V]) ColumnKeys() []C {
	if t == nil {
		return nil
	}
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.data) == 0 {
		return nil
	}
	set := make(map[C]struct{})
	for _, row := range t.data {
		for columnKey := range row {
			set[columnKey] = struct{}{}
		}
	}
	return lo.Keys(set)
}

// All returns a deep-copied built-in map.
func (t *ConcurrentTable[R, C, V]) All() map[R]map[C]V {
	if t == nil {
		return map[R]map[C]V{}
	}
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.data) == 0 {
		return map[R]map[C]V{}
	}
	out := make(map[R]map[C]V, len(t.data))
	for rowKey, row := range t.data {
		out[rowKey] = lo.Assign(map[C]V{}, row)
	}
	return out
}

// Snapshot returns an immutable-style copy in a normal Table.
func (t *ConcurrentTable[R, C, V]) Snapshot() *Table[R, C, V] {
	out := NewTable[R, C, V]()
	for rowKey, row := range t.All() {
		out.SetRow(rowKey, row)
	}
	return out
}

// Range iterates all cells snapshots until fn returns false.
func (t *ConcurrentTable[R, C, V]) Range(fn func(rowKey R, columnKey C, value V) bool) {
	if t == nil || fn == nil {
		return
	}
	for rowKey, row := range t.All() {
		for columnKey, value := range row {
			if !fn(rowKey, columnKey, value) {
				return
			}
		}
	}
}

func (t *ConcurrentTable[R, C, V]) ensureInitLocked() {
	if t.data == nil {
		t.data = make(map[R]map[C]V)
	}
}

func (t *ConcurrentTable[R, C, V]) ensureRowLocked(rowKey R) map[C]V {
	row, ok := t.data[rowKey]
	if !ok {
		row = make(map[C]V)
		t.data[rowKey] = row
	}
	return row
}
