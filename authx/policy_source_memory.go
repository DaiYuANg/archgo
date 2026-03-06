package authx

import (
	"context"
	"sync"
)

// InMemoryPolicySource is a runtime-mutable policy source.
type InMemoryPolicySource struct {
	mu       sync.RWMutex
	snapshot PolicySnapshot
}

// NewInMemoryPolicySource creates in-memory policy source with initial snapshot.
func NewInMemoryPolicySource(initial PolicySnapshot) *InMemoryPolicySource {
	return &InMemoryPolicySource{
		snapshot: initial.clone(),
	}
}

// ReplaceSnapshot atomically replaces in-memory policy snapshot.
func (s *InMemoryPolicySource) ReplaceSnapshot(snapshot PolicySnapshot) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot = snapshot.clone()
}

// LoadPolicies returns current policy snapshot.
func (s *InMemoryPolicySource) LoadPolicies(ctx context.Context) (PolicySnapshot, error) {
	_ = ctx
	if s == nil {
		return PolicySnapshot{}, ErrInvalidPolicy
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshot.clone(), nil
}
