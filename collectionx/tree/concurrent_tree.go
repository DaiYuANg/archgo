package tree

import "sync"

// ConcurrentTree is a goroutine-safe tree wrapper.
// Read APIs that return nodes use snapshots to avoid exposing mutable internals.
// Zero value is ready to use.
type ConcurrentTree[K comparable, V any] struct {
	mu   sync.RWMutex
	tree *Tree[K, V]
}

// NewConcurrentTree creates an empty concurrent tree.
func NewConcurrentTree[K comparable, V any]() *ConcurrentTree[K, V] {
	return &ConcurrentTree[K, V]{
		tree: NewTree[K, V](),
	}
}

// BuildConcurrent constructs a concurrent tree from entries.
func BuildConcurrent[K comparable, V any](entries []Entry[K, V]) (*ConcurrentTree[K, V], error) {
	tree, err := Build(entries)
	if err != nil {
		return nil, err
	}
	return &ConcurrentTree[K, V]{tree: tree}, nil
}

// AddRoot inserts one root node.
func (t *ConcurrentTree[K, V]) AddRoot(id K, value V) error {
	if t == nil {
		return ErrNodeNotFound
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ensureInitLocked()
	return t.tree.AddRoot(id, value)
}

// AddChild inserts one child node under parentID.
func (t *ConcurrentTree[K, V]) AddChild(parentID K, id K, value V) error {
	if t == nil {
		return ErrNodeNotFound
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ensureInitLocked()
	return t.tree.AddChild(parentID, id, value)
}

// Move moves node id under newParentID.
func (t *ConcurrentTree[K, V]) Move(id K, newParentID K) error {
	if t == nil {
		return ErrNodeNotFound
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ensureInitLocked()
	return t.tree.Move(id, newParentID)
}

// Remove deletes one node and its whole subtree.
func (t *ConcurrentTree[K, V]) Remove(id K) bool {
	if t == nil {
		return false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ensureInitLocked()
	return t.tree.Remove(id)
}

// SetValue updates node value by id.
func (t *ConcurrentTree[K, V]) SetValue(id K, value V) bool {
	if t == nil {
		return false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ensureInitLocked()
	return t.tree.SetValue(id, value)
}

// Get returns node by id from snapshot.
func (t *ConcurrentTree[K, V]) Get(id K) (*Node[K, V], bool) {
	return t.Snapshot().Get(id)
}

// Has reports whether id exists.
func (t *ConcurrentTree[K, V]) Has(id K) bool {
	if t == nil {
		return false
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.tree == nil {
		return false
	}
	return t.tree.Has(id)
}

// Parent returns parent node by child id from snapshot.
func (t *ConcurrentTree[K, V]) Parent(id K) (*Node[K, V], bool) {
	return t.Snapshot().Parent(id)
}

// Children returns children snapshot by node id.
func (t *ConcurrentTree[K, V]) Children(id K) []*Node[K, V] {
	return t.Snapshot().Children(id)
}

// Roots returns root nodes snapshot.
func (t *ConcurrentTree[K, V]) Roots() []*Node[K, V] {
	return t.Snapshot().Roots()
}

// Ancestors returns parent chain from direct parent to top root.
func (t *ConcurrentTree[K, V]) Ancestors(id K) []*Node[K, V] {
	return t.Snapshot().Ancestors(id)
}

// Descendants returns all descendants in DFS pre-order.
func (t *ConcurrentTree[K, V]) Descendants(id K) []*Node[K, V] {
	return t.Snapshot().Descendants(id)
}

// RangeDFS iterates all nodes in DFS pre-order until fn returns false.
func (t *ConcurrentTree[K, V]) RangeDFS(fn func(node *Node[K, V]) bool) {
	if t == nil || fn == nil {
		return
	}
	t.Snapshot().RangeDFS(fn)
}

// Len returns total node count.
func (t *ConcurrentTree[K, V]) Len() int {
	if t == nil {
		return 0
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.tree == nil {
		return 0
	}
	return t.tree.Len()
}

// IsEmpty reports whether tree has no nodes.
func (t *ConcurrentTree[K, V]) IsEmpty() bool {
	return t.Len() == 0
}

// Clear removes all nodes.
func (t *ConcurrentTree[K, V]) Clear() {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ensureInitLocked()
	t.tree.Clear()
}

// Snapshot returns an immutable-style copy in a normal Tree.
func (t *ConcurrentTree[K, V]) Snapshot() *Tree[K, V] {
	if t == nil {
		return NewTree[K, V]()
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.tree == nil {
		return NewTree[K, V]()
	}
	return t.tree.Clone()
}

func (t *ConcurrentTree[K, V]) ensureInitLocked() {
	if t.tree == nil {
		t.tree = NewTree[K, V]()
	}
}
