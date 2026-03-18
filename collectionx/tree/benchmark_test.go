package tree

import "testing"

const (
	benchTreeNodes     = 10_000
	benchTreeBranching = 4
	benchTreeLeafID    = benchTreeNodes
)

func buildBenchTree(tb testing.TB) *Tree[int, int] {
	tb.Helper()
	tr := NewTree[int, int]()
	if err := tr.AddRoot(0, 0); err != nil {
		tb.Fatalf("AddRoot() error = %v", err)
	}
	for i := 1; i <= benchTreeNodes; i++ {
		parentID := (i - 1) / benchTreeBranching
		if err := tr.AddChild(parentID, i, i); err != nil {
			tb.Fatalf("AddChild(%d, %d) error = %v", parentID, i, err)
		}
	}
	return tr
}

func buildBenchConcurrentTree(tb testing.TB) *ConcurrentTree[int, int] {
	tb.Helper()
	tr := NewConcurrentTree[int, int]()
	if err := tr.AddRoot(0, 0); err != nil {
		tb.Fatalf("AddRoot() error = %v", err)
	}
	for i := 1; i <= benchTreeNodes; i++ {
		parentID := (i - 1) / benchTreeBranching
		if err := tr.AddChild(parentID, i, i); err != nil {
			tb.Fatalf("AddChild(%d, %d) error = %v", parentID, i, err)
		}
	}
	return tr
}

func BenchmarkTreeGet(b *testing.B) {
	tr := buildBenchTree(b)
	mask := benchTreeNodes - 1

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tr.Get((i & mask) + 1)
	}
}

func BenchmarkTreeChildren(b *testing.B) {
	tr := buildBenchTree(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tr.Children(0)
	}
}

func BenchmarkTreeAncestors(b *testing.B) {
	tr := buildBenchTree(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tr.Ancestors(benchTreeLeafID)
	}
}

func BenchmarkTreeDescendants(b *testing.B) {
	tr := buildBenchTree(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tr.Descendants(0)
	}
}

func BenchmarkTreeClone(b *testing.B) {
	tr := buildBenchTree(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clone := tr.Clone()
		if clone.Len() != tr.Len() {
			b.Fatalf("unexpected clone length: %d", clone.Len())
		}
	}
}

func BenchmarkConcurrentTreeGetParallel(b *testing.B) {
	tr := buildBenchConcurrentTree(b)
	mask := benchTreeNodes - 1

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _ = tr.Get((i & mask) + 1)
			i++
		}
	})
}

func BenchmarkConcurrentTreeDescendants(b *testing.B) {
	tr := buildBenchConcurrentTree(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tr.Descendants(0)
	}
}
