package authx

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	casbin "github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/samber/lo"
)

const defaultCasbinModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, eft

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = (r.sub == p.sub || g(r.sub, p.sub)) && r.obj == p.obj && r.act == p.act
`

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

// CasbinAuthorizer authorizes requests with internal casbin engine.
type CasbinAuthorizer struct {
	enforcer *casbin.Enforcer
	logger   *slog.Logger
}

// NewCasbinAuthorizer creates a casbin authorizer with AuthX default model.
func NewCasbinAuthorizer() (*CasbinAuthorizer, error) {
	m, err := model.NewModelFromString(defaultCasbinModel)
	if err != nil {
		return nil, fmt.Errorf("%w: build casbin model: %v", ErrInvalidAuthorizer, err)
	}

	enforcer, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("%w: build casbin enforcer: %v", ErrInvalidAuthorizer, err)
	}

	return &CasbinAuthorizer{
		enforcer: enforcer,
		logger:   normalizeLogger(nil).With("component", "authx.authorizer", "name", "casbin"),
	}, nil
}

// SetLogger sets slog logger for this authorizer.
func (a *CasbinAuthorizer) SetLogger(logger *slog.Logger) {
	if a == nil {
		return
	}
	a.logger = normalizeLogger(logger).With("component", "authx.authorizer", "name", "casbin")
}

// LoadPermissions loads AuthX permission rules into authorization engine.
func (a *CasbinAuthorizer) LoadPermissions(ctx context.Context, rules ...PermissionRule) error {
	_ = ctx
	if a == nil || a.enforcer == nil {
		return fmt.Errorf("%w: authorizer is not configured", ErrInvalidAuthorizer)
	}
	if len(rules) == 0 {
		return nil
	}
	a.logger.Debug("load permissions started", "rules", len(rules))

	policies := make([][]string, 0, len(rules))
	for _, rule := range rules {
		normalized := normalizePermissionRule(rule)
		if err := validatePermissionRule(normalized); err != nil {
			return err
		}

		effect := lo.Ternary(normalized.Allowed, "allow", "deny")
		policies = append(policies, []string{normalized.Subject, normalized.Resource, normalized.Action, effect})
	}

	if _, err := a.enforcer.AddPolicies(policies); err != nil {
		a.logger.Error("load permissions failed", "error", err.Error())
		return err
	}
	a.logger.Info("load permissions succeeded", "rules", len(policies))
	return nil
}

// LoadRoleBindings loads subject-role bindings into authorization engine.
func (a *CasbinAuthorizer) LoadRoleBindings(ctx context.Context, bindings ...RoleBinding) error {
	_ = ctx
	if a == nil || a.enforcer == nil {
		return fmt.Errorf("%w: authorizer is not configured", ErrInvalidAuthorizer)
	}
	if len(bindings) == 0 {
		return nil
	}
	a.logger.Debug("load role bindings started", "bindings", len(bindings))

	groupings := make([][]string, 0, len(bindings))
	for _, binding := range bindings {
		normalized := normalizeRoleBinding(binding)
		if err := validateRoleBinding(normalized); err != nil {
			return err
		}
		groupings = append(groupings, []string{normalized.Subject, normalized.Role})
	}

	if _, err := a.enforcer.AddGroupingPolicies(groupings); err != nil {
		a.logger.Error("load role bindings failed", "error", err.Error())
		return err
	}
	a.logger.Info("load role bindings succeeded", "bindings", len(groupings))
	return nil
}

// ResetPolicies clears all loaded authorization policies and role bindings.
func (a *CasbinAuthorizer) ResetPolicies(ctx context.Context) error {
	_ = ctx
	if a == nil || a.enforcer == nil {
		return fmt.Errorf("%w: authorizer is not configured", ErrInvalidAuthorizer)
	}
	a.enforcer.ClearPolicy()
	a.logger.Info("policies cleared")
	return nil
}

// Authorize checks permission with internal policy engine.
func (a *CasbinAuthorizer) Authorize(ctx context.Context, identity Identity, request Request) (Decision, error) {
	_ = ctx
	if a == nil || a.enforcer == nil {
		return Decision{}, fmt.Errorf("%w: authorizer is not configured", ErrInvalidAuthorizer)
	}
	if identity == nil || !identity.IsAuthenticated() {
		return Decision{}, ErrUnauthorized
	}
	if err := request.Validate(); err != nil {
		return Decision{}, err
	}

	subject := strings.TrimSpace(identity.ID())
	if subject == "" {
		subject = strings.TrimSpace(identity.Name())
	}
	if subject == "" {
		return Decision{}, ErrUnauthorized
	}

	allowed, err := a.enforcer.Enforce(subject, request.Resource, request.Action)
	if err != nil {
		a.logger.Error("enforce failed", "subject", subject, "action", request.Action, "resource", request.Resource, "error", err.Error())
		return Decision{}, err
	}
	a.logger.Debug("enforce finished", "subject", subject, "action", request.Action, "resource", request.Resource, "allowed", allowed)

	return lo.Ternary(
		allowed,
		Allow("allowed by loaded policy"),
		Deny("denied by loaded policy"),
	), nil
}

func normalizePermissionRule(rule PermissionRule) PermissionRule {
	return PermissionRule{
		Subject:  strings.TrimSpace(rule.Subject),
		Resource: strings.TrimSpace(rule.Resource),
		Action:   strings.TrimSpace(rule.Action),
		Allowed:  rule.Allowed,
	}
}

func validatePermissionRule(rule PermissionRule) error {
	switch {
	case rule.Subject == "":
		return fmt.Errorf("%w: subject is required", ErrInvalidPolicy)
	case rule.Resource == "":
		return fmt.Errorf("%w: resource is required", ErrInvalidPolicy)
	case rule.Action == "":
		return fmt.Errorf("%w: action is required", ErrInvalidPolicy)
	default:
		return nil
	}
}

func normalizeRoleBinding(binding RoleBinding) RoleBinding {
	return RoleBinding{
		Subject: strings.TrimSpace(binding.Subject),
		Role:    strings.TrimSpace(binding.Role),
	}
}

func validateRoleBinding(binding RoleBinding) error {
	switch {
	case binding.Subject == "":
		return fmt.Errorf("%w: subject is required", ErrInvalidPolicy)
	case binding.Role == "":
		return fmt.Errorf("%w: role is required", ErrInvalidPolicy)
	default:
		return nil
	}
}
