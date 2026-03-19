package httpx

import (
	"strings"
	"sync"
)

type routeMatcher struct {
	mu   sync.RWMutex
	root *routeMatcherNode
}

type routeMatcherNode struct {
	staticChildren map[string]*routeMatcherNode
	paramChild     *routeMatcherNode
	routes         []routeMatchEntry
	minSeq         uint64
}

type routeMatchEntry struct {
	seq   uint64
	route RouteInfo
}

func newRouteMatcher() *routeMatcher {
	return &routeMatcher{
		root: &routeMatcherNode{},
	}
}

func (m *routeMatcher) Add(path string, route RouteInfo, seq uint64) {
	if m == nil || seq == 0 {
		return
	}

	segments := splitRouteSegments(path)

	m.mu.Lock()
	defer m.mu.Unlock()

	node := m.ensureRootLocked()
	node.recordMinSeq(seq)

	for _, segment := range segments {
		if isPathParameterSegment(segment) {
			if node.paramChild == nil {
				node.paramChild = &routeMatcherNode{}
			}
			node = node.paramChild
		} else {
			if node.staticChildren == nil {
				node.staticChildren = map[string]*routeMatcherNode{}
			}
			if node.staticChildren[segment] == nil {
				node.staticChildren[segment] = &routeMatcherNode{}
			}
			node = node.staticChildren[segment]
		}
		node.recordMinSeq(seq)
	}

	node.routes = append(node.routes, routeMatchEntry{
		seq:   seq,
		route: route,
	})
}

func (m *routeMatcher) Match(path string) (RouteInfo, bool) {
	if m == nil {
		return RouteInfo{}, false
	}

	segments := splitRouteSegments(path)

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.root == nil {
		return RouteInfo{}, false
	}

	matched, ok := m.root.match(segments, 0)
	if !ok {
		return RouteInfo{}, false
	}
	return matched.route, true
}

func (m *routeMatcher) ensureRootLocked() *routeMatcherNode {
	if m.root == nil {
		m.root = &routeMatcherNode{}
	}
	return m.root
}

func (n *routeMatcherNode) match(segments []string, index int) (routeMatchEntry, bool) {
	if n == nil {
		return routeMatchEntry{}, false
	}

	if index == len(segments) {
		if len(n.routes) == 0 {
			return routeMatchEntry{}, false
		}
		return n.routes[0], true
	}

	segment := segments[index]
	staticChild := n.staticChildren[segment]
	paramChild := n.paramChild

	first, second := orderedRouteChildren(staticChild, paramChild)
	if matched, ok := matchRouteChild(first, segments, index+1); ok {
		if second == nil || second.minSeq == 0 || second.minSeq >= matched.seq {
			return matched, true
		}
		if alternative, ok := matchRouteChild(second, segments, index+1); ok && alternative.seq < matched.seq {
			return alternative, true
		}
		return matched, true
	}

	return matchRouteChild(second, segments, index+1)
}

func (n *routeMatcherNode) recordMinSeq(seq uint64) {
	if n == nil || seq == 0 {
		return
	}
	if n.minSeq == 0 || seq < n.minSeq {
		n.minSeq = seq
	}
}

func orderedRouteChildren(left, right *routeMatcherNode) (*routeMatcherNode, *routeMatcherNode) {
	switch {
	case left == nil:
		return right, nil
	case right == nil:
		return left, nil
	case left.minSeq == 0:
		return right, left
	case right.minSeq == 0:
		return left, right
	case left.minSeq <= right.minSeq:
		return left, right
	default:
		return right, left
	}
}

func matchRouteChild(node *routeMatcherNode, segments []string, index int) (routeMatchEntry, bool) {
	if node == nil {
		return routeMatchEntry{}, false
	}
	return node.match(segments, index)
}

func splitRouteSegments(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func isPathParameterSegment(segment string) bool {
	return strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}")
}
