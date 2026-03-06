package authx

import (
	"context"
	"slices"
)

// PolicySnapshot is a full authorization rules snapshot for hot reload.
type PolicySnapshot struct {
	Permissions  []PermissionRule
	RoleBindings []RoleBinding
}

// PolicySource loads latest authorization snapshot from external source.
type PolicySource interface {
	LoadPolicies(ctx context.Context) (PolicySnapshot, error)
}

// NewPolicySnapshot creates a policy snapshot with copied slices.
func NewPolicySnapshot(permissions []PermissionRule, roleBindings []RoleBinding) PolicySnapshot {
	return PolicySnapshot{
		Permissions:  slices.Clone(permissions),
		RoleBindings: slices.Clone(roleBindings),
	}
}

func (s PolicySnapshot) clone() PolicySnapshot {
	return NewPolicySnapshot(s.Permissions, s.RoleBindings)
}
