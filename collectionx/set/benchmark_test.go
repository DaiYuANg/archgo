package set

import "testing"

const benchSetKeySpace = 1 << 12

func BenchmarkSetContains(b *testing.B) {
	s := NewSet[int]()
	for i := 0; i < benchSetKeySpace; i++ {
		s.Add(i)
	}

	mask := benchSetKeySpace - 1
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Contains(i & mask)
	}
}

func BenchmarkSetAddRemove(b *testing.B) {
	s := NewSet[int]()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Add(i)
		s.Remove(i)
	}
}

func BenchmarkConcurrentSetContainsParallel(b *testing.B) {
	s := NewConcurrentSet[int]()
	for i := 0; i < benchSetKeySpace; i++ {
		s.Add(i)
	}

	mask := benchSetKeySpace - 1
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = s.Contains(i & mask)
			i++
		}
	})
}
