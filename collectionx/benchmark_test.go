package collectionx

import (
	"strconv"
	"testing"
)

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

func BenchmarkRootOrderedMapSetGet(b *testing.B) {
	m := NewOrderedMap[string, int]()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key-" + strconv.Itoa(i&1023)
		m.Set(key, i)
		_, _ = m.Get(key)
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

func BenchmarkRootMultiSetCount(b *testing.B) {
	s := NewMultiSet[int]()
	for i := 0; i < 1024; i++ {
		s.AddN(i, 4)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Count(i & 1023)
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

func BenchmarkRootTrieGet(b *testing.B) {
	t := NewTrie[int]()
	for i := 0; i < 1024; i++ {
		t.Put("user/"+strconv.Itoa(i), i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = t.Get("user/" + strconv.Itoa(i&1023))
	}
}

func BenchmarkRootRangeSetContains(b *testing.B) {
	rs := NewRangeSet[int]()
	for i := 0; i < 1024; i++ {
		start := i * 4
		rs.Add(start, start+2)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rs.Contains((i & 1023) * 4)
	}
}
