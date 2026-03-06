package interval

import "testing"

const benchRangeSetSize = 1 << 10

func BenchmarkRangeSetContains(b *testing.B) {
	s := NewRangeSet[int]()
	for i := 0; i < benchRangeSetSize; i++ {
		start := i * 2
		s.Add(start, start+1)
	}

	mask := (benchRangeSetSize * 2) - 1
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Contains(i & mask)
	}
}

func BenchmarkRangeMapPutGet(b *testing.B) {
	m := NewRangeMap[int, int]()
	sizeMask := benchRangeSetSize - 1

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slot := i & sizeMask
		start := slot * 2
		end := start + 2
		m.Put(start, end, i)
		_, _ = m.Get(start)
	}
}
