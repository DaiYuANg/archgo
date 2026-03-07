package authx

import (
	"context"
	"sync"
)

// MemoryPolicySource is an in-memory policy source that holds
// a snapshot of permissions and role bindings.
//
// Use cases:
//   - Bootstrap policies during application startup
//   - Testing and development environments
//   - Static policies that rarely change
//   - Caching layer for remote policy sources
type MemoryPolicySource struct {
	mu       sync.RWMutex
	snapshot PolicySnapshot
	version  int64
	name     string
}

// MemoryPolicySourceConfig configures a memory policy source.
type MemoryPolicySourceConfig struct {
	// Name is the optional name for this policy source.
	// Defaults to "memory" if empty.
	Name string
	// InitialPermissions are the initial permission rules.
	InitialPermissions []PermissionRule
	// InitialRoleBindings are the initial role bindings.
	InitialRoleBindings []RoleBinding
}

// NewMemoryPolicySource creates a new in-memory policy source.
func NewMemoryPolicySource(cfg MemoryPolicySourceConfig) *MemoryPolicySource {
	name := cfg.Name
	if name == "" {
		name = "memory"
	}

	return &MemoryPolicySource{
		snapshot: PolicySnapshot{
			Permissions:  slicesClone(cfg.InitialPermissions),
			RoleBindings: slicesClone(cfg.InitialRoleBindings),
		},
		version: 1,
		name:    name,
	}
}

// LoadPolicies returns the current in-memory snapshot.
func (s *MemoryPolicySource) LoadPolicies(ctx context.Context) (PolicySnapshot, error) {
	_ = ctx

	s.mu.RLock()
	defer s.mu.RUnlock()

	return PolicySnapshot{
		Permissions:  slicesClone(s.snapshot.Permissions),
		RoleBindings: slicesClone(s.snapshot.RoleBindings),
	}, nil
}

// Version returns the current policy version.
func (s *MemoryPolicySource) Version() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.version
}

// Update replaces the in-memory snapshot and increments the version.
// Returns the new version number.
func (s *MemoryPolicySource) Update(permissions []PermissionRule, roleBindings []RoleBinding) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.snapshot.Permissions = slicesClone(permissions)
	s.snapshot.RoleBindings = slicesClone(roleBindings)
	s.version++

	return s.version
}

// UpdateSnapshot replaces the in-memory snapshot with a new snapshot.
// Returns the new version number.
func (s *MemoryPolicySource) UpdateSnapshot(snapshot PolicySnapshot) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.snapshot.Permissions = slicesClone(snapshot.Permissions)
	s.snapshot.RoleBindings = slicesClone(snapshot.RoleBindings)
	s.version++

	return s.version
}

// Clear removes all policies and increments the version.
func (s *MemoryPolicySource) Clear() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.snapshot.Permissions = nil
	s.snapshot.RoleBindings = nil
	s.version++

	return s.version
}

// AddPermission adds a single permission rule and increments the version.
func (s *MemoryPolicySource) AddPermission(perm PermissionRule) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.snapshot.Permissions = append(s.snapshot.Permissions, perm)
	s.version++

	return s.version
}

// RemovePermission removes permission rules matching the given predicate.
// Returns the new version number.
func (s *MemoryPolicySource) RemovePermission(match func(PermissionRule) bool) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	remaining := make([]PermissionRule, 0, len(s.snapshot.Permissions))
	for _, perm := range s.snapshot.Permissions {
		if !match(perm) {
			remaining = append(remaining, perm)
		}
	}

	if len(remaining) == len(s.snapshot.Permissions) {
		return s.version
	}

	s.snapshot.Permissions = remaining
	s.version++

	return s.version
}

// AddRoleBinding adds a single role binding and increments the version.
func (s *MemoryPolicySource) AddRoleBinding(binding RoleBinding) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.snapshot.RoleBindings = append(s.snapshot.RoleBindings, binding)
	s.version++

	return s.version
}

// RemoveRoleBinding removes role bindings matching the given predicate.
// Returns the new version number.
func (s *MemoryPolicySource) RemoveRoleBinding(match func(RoleBinding) bool) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	remaining := make([]RoleBinding, 0, len(s.snapshot.RoleBindings))
	for _, binding := range s.snapshot.RoleBindings {
		if !match(binding) {
			remaining = append(remaining, binding)
		}
	}

	if len(remaining) == len(s.snapshot.RoleBindings) {
		return s.version
	}

	s.snapshot.RoleBindings = remaining
	s.version++

	return s.version
}

// Snapshot returns a copy of the current snapshot.
func (s *MemoryPolicySource) Snapshot() PolicySnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return PolicySnapshot{
		Permissions:  slicesClone(s.snapshot.Permissions),
		RoleBindings: slicesClone(s.snapshot.RoleBindings),
	}
}

// Name returns the policy source name.
func (s *MemoryPolicySource) Name() string {
	return s.name
}

// slicesClone creates a shallow clone of a slice.
func slicesClone[T any](s []T) []T {
	if s == nil {
		return nil
	}
	result := make([]T, len(s))
	copy(result, s)
	return result
}

// Ensure MemoryPolicySource implements PolicySource.
var _ PolicySource = (*MemoryPolicySource)(nil)

// StaticPolicySource is a read-only policy source that always returns
// the same snapshot. Useful for testing and fixed configurations.
type StaticPolicySource struct {
	snapshot PolicySnapshot
	name     string
}

// StaticPolicySourceConfig configures a static policy source.
type StaticPolicySourceConfig struct {
	// Name is the optional name for this policy source.
	// Defaults to "static" if empty.
	Name string
	// Permissions are the fixed permission rules.
	Permissions []PermissionRule
	// RoleBindings are the fixed role bindings.
	RoleBindings []RoleBinding
}

// NewStaticPolicySource creates a new read-only static policy source.
func NewStaticPolicySource(cfg StaticPolicySourceConfig) *StaticPolicySource {
	name := cfg.Name
	if name == "" {
		name = "static"
	}

	return &StaticPolicySource{
		snapshot: PolicySnapshot{
			Permissions:  slicesClone(cfg.Permissions),
			RoleBindings: slicesClone(cfg.RoleBindings),
		},
		name: name,
	}
}

// LoadPolicies always returns the fixed snapshot.
func (s *StaticPolicySource) LoadPolicies(ctx context.Context) (PolicySnapshot, error) {
	return PolicySnapshot{
		Permissions:  slicesClone(s.snapshot.Permissions),
		RoleBindings: slicesClone(s.snapshot.RoleBindings),
	}, nil
}

// Name returns the policy source name.
func (s *StaticPolicySource) Name() string {
	return s.name
}

// Ensure StaticPolicySource implements PolicySource.
var _ PolicySource = (*StaticPolicySource)(nil)

// MutablePolicySource is a policy source that allows external mutation
// through a function callback.
type MutablePolicySource struct {
	mu       sync.RWMutex
	snapshot PolicySnapshot
	version  int64
	name     string
	mutator  func(context.Context, PolicySnapshot) (PolicySnapshot, error)
}

// MutablePolicySourceConfig configures a mutable policy source.
type MutablePolicySourceConfig struct {
	// Name is the optional name for this policy source.
	// Defaults to "mutable" if empty.
	Name string
	// InitialSnapshot is the initial policy snapshot.
	InitialSnapshot PolicySnapshot
	// Mutator is called during LoadPolicies to transform the snapshot.
	// If nil, the source behaves like a MemoryPolicySource.
	Mutator func(context.Context, PolicySnapshot) (PolicySnapshot, error)
}

// NewMutablePolicySource creates a new mutable policy source.
func NewMutablePolicySource(cfg MutablePolicySourceConfig) *MutablePolicySource {
	name := cfg.Name
	if name == "" {
		name = "mutable"
	}

	return &MutablePolicySource{
		snapshot: PolicySnapshot{
			Permissions:  slicesClone(cfg.InitialSnapshot.Permissions),
			RoleBindings: slicesClone(cfg.InitialSnapshot.RoleBindings),
		},
		version: 1,
		name:    name,
		mutator: cfg.Mutator,
	}
}

// LoadPolicies loads the current snapshot and optionally applies the mutator.
func (s *MutablePolicySource) LoadPolicies(ctx context.Context) (PolicySnapshot, error) {
	s.mu.RLock()
	current := PolicySnapshot{
		Permissions:  slicesClone(s.snapshot.Permissions),
		RoleBindings: slicesClone(s.snapshot.RoleBindings),
	}
	s.mu.RUnlock()

	if s.mutator == nil {
		return current, nil
	}

	return s.mutator(ctx, current)
}

// Update replaces the base snapshot and increments the version.
// The mutator (if any) will be applied on next LoadPolicies call.
func (s *MutablePolicySource) Update(permissions []PermissionRule, roleBindings []RoleBinding) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.snapshot.Permissions = slicesClone(permissions)
	s.snapshot.RoleBindings = slicesClone(roleBindings)
	s.version++

	return s.version
}

// Name returns the policy source name.
func (s *MutablePolicySource) Name() string {
	return s.name
}

// Ensure MutablePolicySource implements PolicySource.
var _ PolicySource = (*MutablePolicySource)(nil)
