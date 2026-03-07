package authx

import (
	"context"
	"slices"
)

// PermissionRule is an AuthX-owned authorization policy rule.
type PermissionRule struct {
	Subject  string
	Resource string
	Action   string
	Allowed  bool
}

// RoleBinding links a subject to a role.
type RoleBinding struct {
	Subject string
	Role    string
}

// AllowPermission creates an allow policy rule.
func AllowPermission(subject, resource, action string) PermissionRule {
	return PermissionRule{
		Subject:  subject,
		Resource: resource,
		Action:   action,
		Allowed:  true,
	}
}

// DenyPermission creates a deny policy rule.
func DenyPermission(subject, resource, action string) PermissionRule {
	return PermissionRule{
		Subject:  subject,
		Resource: resource,
		Action:   action,
		Allowed:  false,
	}
}

// NewRoleBinding creates a role binding.
func NewRoleBinding(subject, role string) RoleBinding {
	return RoleBinding{
		Subject: subject,
		Role:    role,
	}
}

// PolicySnapshot is a full authorization rules snapshot for hot reload.
type PolicySnapshot struct {
	Permissions  []PermissionRule
	RoleBindings []RoleBinding
}

// PolicySource loads latest authorization snapshot from external source.
type PolicySource interface {
	// LoadPolicies loads the latest policy snapshot.
	LoadPolicies(ctx context.Context) (PolicySnapshot, error)
	// Name returns the policy source name for identification.
	Name() string
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
