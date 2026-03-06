package tree

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTree_AddAndRelationships(t *testing.T) {
	tr := NewTree[int, string]()

	require.NoError(t, tr.AddRoot(1, "root"))
	require.NoError(t, tr.AddChild(1, 2, "child-a"))
	require.NoError(t, tr.AddChild(2, 3, "child-b"))

	n3, ok := tr.Get(3)
	require.True(t, ok)
	require.Equal(t, "child-b", n3.Value())

	parent, ok := tr.Parent(3)
	require.True(t, ok)
	require.Equal(t, 2, parent.ID())

	ancestors := tr.Ancestors(3)
	require.Equal(t, []int{2, 1}, nodeIDs(ancestors))

	descendants := tr.Descendants(1)
	require.Equal(t, []int{2, 3}, nodeIDs(descendants))
	require.Equal(t, []int{2}, nodeIDs(tr.Children(1)))
	require.Equal(t, 3, tr.Len())
}

func TestTree_MoveAndCycleDetection(t *testing.T) {
	tr := NewTree[int, string]()

	require.NoError(t, tr.AddRoot(1, "root-a"))
	require.NoError(t, tr.AddRoot(2, "root-b"))
	require.NoError(t, tr.AddChild(1, 3, "child"))

	require.NoError(t, tr.Move(3, 2))

	parent, ok := tr.Parent(3)
	require.True(t, ok)
	require.Equal(t, 2, parent.ID())
	require.Equal(t, []int{1, 2}, nodeIDs(tr.Roots()))

	require.ErrorIs(t, tr.Move(2, 3), ErrCycleDetected)
}

func TestTree_RemoveSubtree(t *testing.T) {
	tr := NewTree[int, string]()

	require.NoError(t, tr.AddRoot(1, "r1"))
	require.NoError(t, tr.AddChild(1, 2, "c1"))
	require.NoError(t, tr.AddChild(2, 3, "c2"))
	require.NoError(t, tr.AddRoot(4, "r2"))

	require.True(t, tr.Remove(2))
	require.False(t, tr.Has(2))
	require.False(t, tr.Has(3))
	require.True(t, tr.Has(1))
	require.True(t, tr.Has(4))
	require.Equal(t, []int{1, 4}, nodeIDs(tr.Roots()))
	require.Equal(t, 2, tr.Len())
}

func TestTree_CloneIsolation(t *testing.T) {
	tr := NewTree[int, string]()
	require.NoError(t, tr.AddRoot(1, "root"))
	require.NoError(t, tr.AddChild(1, 2, "child"))

	cloned := tr.Clone()
	require.Equal(t, tr.Len(), cloned.Len())
	require.True(t, cloned.SetValue(2, "cloned-only"))

	origNode, _ := tr.Get(2)
	clonedNode, _ := cloned.Get(2)
	require.Equal(t, "child", origNode.Value())
	require.Equal(t, "cloned-only", clonedNode.Value())
}

func TestBuild(t *testing.T) {
	entries := []Entry[int, string]{
		ChildEntry(2, 1, "child-a"),
		RootEntry(1, "root"),
		ChildEntry(3, 2, "child-b"),
	}

	tr, err := Build(entries)
	require.NoError(t, err)
	require.Equal(t, 3, tr.Len())
	require.Equal(t, []int{1}, nodeIDs(tr.Roots()))
	require.Equal(t, []int{2, 3}, nodeIDs(tr.Descendants(1)))
}

func TestBuild_WithInvalidParent(t *testing.T) {
	entries := []Entry[int, string]{
		ChildEntry(1, 9, "orphan"),
	}

	_, err := Build(entries)
	require.ErrorIs(t, err, ErrParentNotFound)
}

func TestBuild_WithCycle(t *testing.T) {
	entries := []Entry[int, string]{
		ChildEntry(1, 2, "a"),
		ChildEntry(2, 1, "b"),
	}

	_, err := Build(entries)
	require.ErrorIs(t, err, ErrCycleDetected)
}

func TestBuild_WithDuplicateNode(t *testing.T) {
	entries := []Entry[int, string]{
		RootEntry(1, "a"),
		RootEntry(1, "b"),
	}

	_, err := Build(entries)
	require.ErrorIs(t, err, ErrNodeAlreadyExists)
}

func nodeIDs(nodes []*Node[int, string]) []int {
	if len(nodes) == 0 {
		return nil
	}
	out := make([]int, 0, len(nodes))
	for _, node := range nodes {
		out = append(out, node.ID())
	}
	return out
}
