package collectionx

import "github.com/DaiYuANg/arcgo/collectionx/tree"

type Tree[K comparable, V any] = tree.Tree[K, V]

func NewTree[K comparable, V any]() *Tree[K, V] {
	return tree.NewTree[K, V]()
}

type ConcurrentTree[K comparable, V any] = tree.ConcurrentTree[K, V]

func NewConcurrentTree[K comparable, V any]() *ConcurrentTree[K, V] {
	return tree.NewConcurrentTree[K, V]()
}

type TreeNode[K comparable, V any] = tree.Node[K, V]

type TreeEntry[K comparable, V any] = tree.Entry[K, V]

func NewRootTreeEntry[K comparable, V any](id K, value V) TreeEntry[K, V] {
	return tree.RootEntry(id, value)
}

func NewChildTreeEntry[K comparable, V any](id K, parentID K, value V) TreeEntry[K, V] {
	return tree.ChildEntry(id, parentID, value)
}

func BuildTree[K comparable, V any](entries []TreeEntry[K, V]) (*Tree[K, V], error) {
	return tree.Build(entries)
}

func BuildConcurrentTree[K comparable, V any](entries []TreeEntry[K, V]) (*ConcurrentTree[K, V], error) {
	return tree.BuildConcurrent(entries)
}

var (
	ErrTreeNodeAlreadyExists = tree.ErrNodeAlreadyExists
	ErrTreeNodeNotFound      = tree.ErrNodeNotFound
	ErrTreeParentNotFound    = tree.ErrParentNotFound
	ErrTreeCycleDetected     = tree.ErrCycleDetected
)
