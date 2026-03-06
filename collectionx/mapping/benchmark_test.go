package mapping

import "testing"

const (
	benchMapKeySpace = 1 << 12
	benchTableDim    = 1 << 6
)

func BenchmarkMapSetGet(b *testing.B) {
	m := NewMap[int, int]()
	mask := benchMapKeySpace - 1

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := i & mask
		m.Set(k, i)
		_, _ = m.Get(k)
	}
}

func BenchmarkConcurrentMapGetParallel(b *testing.B) {
	m := NewConcurrentMap[int, int]()
	for i := 0; i < benchMapKeySpace; i++ {
		m.Set(i, i)
	}

	mask := benchMapKeySpace - 1
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _ = m.Get(i & mask)
			i++
		}
	})
}

func BenchmarkMultiMapPutGet(b *testing.B) {
	m := NewMultiMap[int, int]()
	mask := benchMapKeySpace - 1

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := i & mask
		m.Put(k, i)
		_ = m.Get(k)
	}
}

func BenchmarkTablePutGet(b *testing.B) {
	t := NewTable[int, int, int]()
	rowMask := benchTableDim - 1
	colMask := benchTableDim - 1

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		row := i & rowMask
		col := (i >> 6) & colMask
		t.Put(row, col, i)
		_, _ = t.Get(row, col)
	}
}

func BenchmarkConcurrentTableGetParallel(b *testing.B) {
	t := NewConcurrentTable[int, int, int]()
	for row := 0; row < benchTableDim; row++ {
		for col := 0; col < benchTableDim; col++ {
			t.Put(row, col, row+col)
		}
	}

	rowMask := benchTableDim - 1
	colMask := benchTableDim - 1
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			row := i & rowMask
			col := (i >> 6) & colMask
			_, _ = t.Get(row, col)
			i++
		}
	})
}
