package interval

import "testing"

const benchRangeSetSize = 1 << 10

func BenchmarkRangeContains(b *testing.B) {
	r := Range[int]{Start: 128, End: 2048}
	mask := 4095

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Contains(i & mask)
	}
}

func BenchmarkRangeMerge(b *testing.B) {
	left := Range[int]{Start: 100, End: 200}
	right := Range[int]{Start: 150, End: 250}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		merged, ok := left.Merge(right)
		if !ok || merged.Start != 100 || merged.End != 250 {
			b.Fatalf("unexpected merge result: %+v ok=%v", merged, ok)
		}
	}
}

func BenchmarkRangeSetAdd(b *testing.B) {
	s := NewRangeSet[int]()
	sizeMask := benchRangeSetSize - 1

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slot := i & sizeMask
		start := slot * 4
		s.Add(start, start+2)
	}
}

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

func BenchmarkRangeMapGet(b *testing.B) {
	m := NewRangeMap[int, int]()
	for i := 0; i < benchRangeSetSize; i++ {
		start := i * 4
		m.Put(start, start+3, i)
	}
	mask := benchRangeSetSize - 1

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		point := (i & mask) * 4
		_, _ = m.Get(point)
	}
}
