package tree

import "testing"

func BenchmarkTreeDescendants(b *testing.B) {
	tr := NewTree[int, int]()
	_ = tr.AddRoot(0, 0)
	for i := 1; i <= 10_000; i++ {
		_ = tr.AddChild(0, i, i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tr.Descendants(0)
	}
}

func BenchmarkConcurrentTreeDescendants(b *testing.B) {
	tr := NewConcurrentTree[int, int]()
	_ = tr.AddRoot(0, 0)
	for i := 1; i <= 10_000; i++ {
		_ = tr.AddChild(0, i, i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tr.Descendants(0)
	}
}
