package tree

// Clone returns a deep copy preserving parent-children structure.
func (t *Tree[K, V]) Clone() *Tree[K, V] {
	cloned := NewTree[K, V]()
	if t == nil || t.nodes == nil || t.nodes.IsEmpty() {
		return cloned
	}

	type pair struct {
		source *Node[K, V]
		target *Node[K, V]
	}

	stack := make([]pair, 0)
	for _, root := range t.Roots() {
		rootClone := newNode(root.ID(), root.Value())
		cloned.roots.Add(rootClone)
		cloned.nodes.Set(rootClone.ID(), rootClone)
		stack = append(stack, pair{source: root, target: rootClone})
	}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		for _, sourceChild := range current.source.Children() {
			targetChild := newNode(sourceChild.ID(), sourceChild.Value())
			targetChild.parent = current.target
			current.target.children.Add(targetChild)
			cloned.nodes.Set(targetChild.ID(), targetChild)
			stack = append(stack, pair{source: sourceChild, target: targetChild})
		}
	}

	return cloned
}

func cloneSubtreeDetached[K comparable, V any](root *Node[K, V]) *Node[K, V] {
	if root == nil {
		return nil
	}

	rootClone := newNode(root.ID(), root.Value())
	type pair struct {
		source *Node[K, V]
		target *Node[K, V]
	}
	stack := []pair{{source: root, target: rootClone}}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		for _, sourceChild := range current.source.Children() {
			targetChild := newNode(sourceChild.ID(), sourceChild.Value())
			targetChild.parent = current.target
			current.target.children.Add(targetChild)
			stack = append(stack, pair{source: sourceChild, target: targetChild})
		}
	}

	return rootClone
}
