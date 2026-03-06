package collectionx

import "testing"

func BenchmarkRootMapSetGet(b *testing.B) {
	m := NewMap[string, int]()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.Set("key", i)
		value, ok := m.Get("key")
		if !ok || value != i {
			b.Fatalf("unexpected map value: ok=%v value=%d expect=%d", ok, value, i)
		}
	}
}

func BenchmarkRootSetContains(b *testing.B) {
	s := NewSet[int]()
	for i := 0; i < 1024; i++ {
		s.Add(i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if !s.Contains(i % 1024) {
			b.Fatal("expected value to exist in set")
		}
	}
}

func BenchmarkRootListAppendGet(b *testing.B) {
	l := NewList[int]()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Add(i)
		value, ok := l.Get(l.Len() - 1)
		if !ok || value != i {
			b.Fatalf("unexpected list value: ok=%v value=%d expect=%d", ok, value, i)
		}
	}
}
