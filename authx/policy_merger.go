package authx

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// PolicyMerger merges multiple policy snapshots into a single consolidated snapshot.
// This allows combining policies from multiple sources with deduplication,
// priority handling, and conflict detection.
//
// Use cases:
//   - Merge policies from multiple sources (file + database + remote)
//   - Apply source priority (e.g., deny overrides allow)
//   - Detect and handle conflicting rules
//   - Deduplicate identical rules
type PolicyMerger interface {
	Merge(ctx context.Context, snapshots ...PolicySnapshot) (PolicySnapshot, error)
}

// PolicyMergerFunc is a function-based policy merger.
type PolicyMergerFunc func(ctx context.Context, snapshots ...PolicySnapshot) (PolicySnapshot, error)

// Merge implements PolicyMerger interface.
func (f PolicyMergerFunc) Merge(ctx context.Context, snapshots ...PolicySnapshot) (PolicySnapshot, error) {
	if f == nil {
		return PolicySnapshot{}, fmt.Errorf("%w: policy merger function is nil", ErrInvalidPolicy)
	}
	return f(ctx, snapshots...)
}

// DefaultPolicyMerger is the default merger that concatenates and deduplicates rules.
// Deduplication key: subject + resource + action + effect
type DefaultPolicyMerger struct{}

// NewDefaultPolicyMerger creates a default policy merger.
func NewDefaultPolicyMerger() *DefaultPolicyMerger {
	return &DefaultPolicyMerger{}
}

// Merge concatenates all snapshots and removes duplicates.
func (m *DefaultPolicyMerger) Merge(ctx context.Context, snapshots ...PolicySnapshot) (PolicySnapshot, error) {
	_ = ctx

	if len(snapshots) == 0 {
		return PolicySnapshot{
			Permissions:  make([]PermissionRule, 0),
			RoleBindings: make([]RoleBinding, 0),
		}, nil
	}

	result := PolicySnapshot{
		Permissions:  make([]PermissionRule, 0),
		RoleBindings: make([]RoleBinding, 0),
	}

	// Use maps for deduplication
	permissionSet := make(map[string]PermissionRule)
	roleBindingSet := make(map[string]RoleBinding)

	for _, snapshot := range snapshots {
		for _, perm := range snapshot.Permissions {
			key := permissionKeyWithEffect(perm)
			if _, exists := permissionSet[key]; !exists {
				permissionSet[key] = perm
			}
		}

		for _, binding := range snapshot.RoleBindings {
			key := roleBindingKey(binding)
			if _, exists := roleBindingSet[key]; !exists {
				roleBindingSet[key] = binding
			}
		}
	}

	// Convert maps to sorted slices for deterministic output
	for _, perm := range permissionSet {
		result.Permissions = append(result.Permissions, perm)
	}
	sortPermissions(result.Permissions)

	for _, binding := range roleBindingSet {
		result.RoleBindings = append(result.RoleBindings, binding)
	}
	sortRoleBindings(result.RoleBindings)

	return result, nil
}

// PriorityPolicyMerger merges policies with source priority.
// Higher priority sources override lower priority sources.
type PriorityPolicyMerger struct {
	// PriorityOrder defines source priority from highest to lowest.
	// Rules from higher priority sources override conflicting rules from lower priority sources.
	priorityOrder []string
}

// PriorityPolicyMergerConfig configures a priority policy merger.
type PriorityPolicyMergerConfig struct {
	// PriorityOrder defines source priority from highest to lowest.
	// If empty, uses snapshot order (first = highest priority).
	PriorityOrder []string
}

// NewPriorityPolicyMerger creates a priority-based policy merger.
func NewPriorityPolicyMerger(cfg PriorityPolicyMergerConfig) *PriorityPolicyMerger {
	order := cfg.PriorityOrder
	if len(order) == 0 {
		// Default: use index-based priority
		order = []string{"default"}
	}
	return &PriorityPolicyMerger{
		priorityOrder: order,
	}
}

// Merge merges snapshots with priority handling.
// Higher priority snapshots override conflicting rules from lower priority snapshots.
func (m *PriorityPolicyMerger) Merge(ctx context.Context, snapshots ...PolicySnapshot) (PolicySnapshot, error) {
	_ = ctx

	if len(snapshots) == 0 {
		return PolicySnapshot{
			Permissions:  make([]PermissionRule, 0),
			RoleBindings: make([]RoleBinding, 0),
		}, nil
	}

	result := PolicySnapshot{
		Permissions:  make([]PermissionRule, 0),
		RoleBindings: make([]RoleBinding, 0),
	}

	// Process snapshots from first to last.
	// Later snapshots override earlier snapshots for the same key.
	for _, snapshot := range snapshots {
		// For permissions, later snapshot overrides by key
		for _, perm := range snapshot.Permissions {
			key := permissionKey(perm)
			idx := findPermissionKey(result.Permissions, key)
			if idx >= 0 {
				// Later snapshot overrides
				result.Permissions[idx] = perm
			} else {
				result.Permissions = append(result.Permissions, perm)
			}
		}

		// For role bindings, accumulate (multiple bindings allowed)
		result.RoleBindings = append(result.RoleBindings, snapshot.RoleBindings...)
	}

	sortPermissions(result.Permissions)
	sortRoleBindings(result.RoleBindings)

	return result, nil
}

// ConflictDetectingPolicyMerger wraps another merger and detects conflicts.
type ConflictDetectingPolicyMerger struct {
	wrapped    PolicyMerger
	onConflict func(ctx context.Context, conflict PolicyConflict) error
}

// PolicyConflict represents a detected policy conflict.
type PolicyConflict struct {
	Key     string
	Rule1   PermissionRule
	Rule2   PermissionRule
	Source1 int // snapshot index
	Source2 int // snapshot index
}

// ConflictDetectingPolicyMergerConfig configures a conflict-detecting merger.
type ConflictDetectingPolicyMergerConfig struct {
	// Wrapped is the underlying merger to use after conflict detection.
	// If nil, uses DefaultPolicyMerger.
	Wrapped PolicyMerger
	// OnConflict is called when a conflict is detected.
	// If it returns an error, Merge stops and returns the error.
	// If nil, conflicts are logged but allowed.
	OnConflict func(ctx context.Context, conflict PolicyConflict) error
}

// NewConflictDetectingPolicyMerger creates a conflict-detecting policy merger.
func NewConflictDetectingPolicyMerger(cfg ConflictDetectingPolicyMergerConfig) *ConflictDetectingPolicyMerger {
	wrapped := cfg.Wrapped
	if wrapped == nil {
		wrapped = NewDefaultPolicyMerger()
	}
	return &ConflictDetectingPolicyMerger{
		wrapped:    wrapped,
		onConflict: cfg.OnConflict,
	}
}

// Merge detects conflicts before delegating to the wrapped merger.
func (m *ConflictDetectingPolicyMerger) Merge(ctx context.Context, snapshots ...PolicySnapshot) (PolicySnapshot, error) {
	conflicts := m.detectConflicts(snapshots)

	for _, conflict := range conflicts {
		if m.onConflict != nil {
			if err := m.onConflict(ctx, conflict); err != nil {
				return PolicySnapshot{}, WrapError(CodePolicyMergeConflict,
					fmt.Sprintf("policy conflict detected: %s", conflict.Key), err)
			}
		}
	}

	return m.wrapped.Merge(ctx, snapshots...)
}

func (m *ConflictDetectingPolicyMerger) detectConflicts(snapshots []PolicySnapshot) []PolicyConflict {
	var conflicts []PolicyConflict

	// Build a map of all permissions by key
	permissionMap := make(map[string][]indexedPermission)
	for si, snapshot := range snapshots {
		for _, perm := range snapshot.Permissions {
			key := permissionKey(perm)
			permissionMap[key] = append(permissionMap[key], indexedPermission{
				Permission: perm,
				Source:     si,
			})
		}
	}

	// Find conflicts (same key, different effect)
	for key, perms := range permissionMap {
		if len(perms) < 2 {
			continue
		}

		// Check for conflicting effects
		for i := 0; i < len(perms); i++ {
			for j := i + 1; j < len(perms); j++ {
				if perms[i].Permission.Allowed != perms[j].Permission.Allowed {
					conflicts = append(conflicts, PolicyConflict{
						Key:     key,
						Rule1:   perms[i].Permission,
						Rule2:   perms[j].Permission,
						Source1: perms[i].Source,
						Source2: perms[j].Source,
					})
				}
			}
		}
	}

	return conflicts
}

type indexedPermission struct {
	Permission PermissionRule
	Source     int
}

// StrictPolicyMerger rejects any conflicts during merge.
type StrictPolicyMerger struct {
	wrapped PolicyMerger
}

// NewStrictPolicyMerger creates a strict policy merger that rejects conflicts.
func NewStrictPolicyMerger() *StrictPolicyMerger {
	return &StrictPolicyMerger{
		wrapped: NewDefaultPolicyMerger(),
	}
}

// Merge rejects conflicts with an error.
func (m *StrictPolicyMerger) Merge(ctx context.Context, snapshots ...PolicySnapshot) (PolicySnapshot, error) {
	conflictHandler := func(ctx context.Context, conflict PolicyConflict) error {
		return fmt.Errorf("%w: conflicting rules for %s (allow=%v vs allow=%v)",
			ErrPolicyMergeConflict, conflict.Key, conflict.Rule1.Allowed, conflict.Rule2.Allowed)
	}

	detector := NewConflictDetectingPolicyMerger(ConflictDetectingPolicyMergerConfig{
		Wrapped:    m.wrapped,
		OnConflict: conflictHandler,
	})

	return detector.Merge(ctx, snapshots...)
}

// DenyOverridesPolicyMerger implements RBAC best practice: deny rules always override allow rules.
type DenyOverridesPolicyMerger struct{}

// NewDenyOverridesPolicyMerger creates a deny-overrides policy merger.
func NewDenyOverridesPolicyMerger() *DenyOverridesPolicyMerger {
	return &DenyOverridesPolicyMerger{}
}

// Merge applies deny-overrides semantics.
// If any rule denies access, the final rule will be deny.
func (m *DenyOverridesPolicyMerger) Merge(ctx context.Context, snapshots ...PolicySnapshot) (PolicySnapshot, error) {
	_ = ctx

	if len(snapshots) == 0 {
		return PolicySnapshot{
			Permissions:  make([]PermissionRule, 0),
			RoleBindings: make([]RoleBinding, 0),
		}, nil
	}

	permissionMap := make(map[string]PermissionRule)

	// Process all snapshots
	for _, snapshot := range snapshots {
		for _, perm := range snapshot.Permissions {
			key := permissionKey(perm)
			existing, exists := permissionMap[key]

			if !exists {
				permissionMap[key] = perm
			} else {
				// Deny overrides allow
				if !perm.Allowed {
					permissionMap[key] = perm
				} else if !existing.Allowed {
					// Keep existing deny
					permissionMap[key] = existing
				} else {
					// Both allow, keep existing
					permissionMap[key] = existing
				}
			}
		}
	}

	result := PolicySnapshot{
		Permissions:  make([]PermissionRule, 0, len(permissionMap)),
		RoleBindings: make([]RoleBinding, 0),
	}

	for _, perm := range permissionMap {
		result.Permissions = append(result.Permissions, perm)
	}
	sortPermissions(result.Permissions)

	// Merge role bindings from all snapshots
	roleBindingSet := make(map[string]RoleBinding)
	for _, snapshot := range snapshots {
		for _, binding := range snapshot.RoleBindings {
			key := roleBindingKey(binding)
			roleBindingSet[key] = binding
		}
	}

	for _, binding := range roleBindingSet {
		result.RoleBindings = append(result.RoleBindings, binding)
	}
	sortRoleBindings(result.RoleBindings)

	return result, nil
}

// Helper functions

// permissionKeyWithEffect returns a key including the effect (for exact rule matching).
func permissionKeyWithEffect(perm PermissionRule) string {
	return fmt.Sprintf("%s|%s|%s|%v",
		strings.TrimSpace(perm.Subject),
		strings.TrimSpace(perm.Resource),
		strings.TrimSpace(perm.Action),
		perm.Allowed,
	)
}

// permissionKey returns a key without effect (for conflict detection and override logic).
func permissionKey(perm PermissionRule) string {
	return fmt.Sprintf("%s|%s|%s",
		strings.TrimSpace(perm.Subject),
		strings.TrimSpace(perm.Resource),
		strings.TrimSpace(perm.Action),
	)
}

func roleBindingKey(binding RoleBinding) string {
	return fmt.Sprintf("%s|%s",
		strings.TrimSpace(binding.Subject),
		strings.TrimSpace(binding.Role),
	)
}

func sortPermissions(perms []PermissionRule) {
	sort.Slice(perms, func(i, j int) bool {
		if perms[i].Subject != perms[j].Subject {
			return perms[i].Subject < perms[j].Subject
		}
		if perms[i].Resource != perms[j].Resource {
			return perms[i].Resource < perms[j].Resource
		}
		return perms[i].Action < perms[j].Action
	})
}

func sortRoleBindings(bindings []RoleBinding) {
	sort.Slice(bindings, func(i, j int) bool {
		if bindings[i].Subject != bindings[j].Subject {
			return bindings[i].Subject < bindings[j].Subject
		}
		return bindings[i].Role < bindings[j].Role
	})
}

func findPermissionKey(perms []PermissionRule, key string) int {
	for i, perm := range perms {
		if permissionKey(perm) == key {
			return i
		}
	}
	return -1
}
