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

// CasbinModelType defines the type of Casbin model to use.
type CasbinModelType string

const (
	// CasbinModelExact uses exact matching for resources.
	CasbinModelExact CasbinModelType = "exact"
	// CasbinModelPrefix uses prefix matching (e.g., /api/admin/*).
	CasbinModelPrefix CasbinModelType = "prefix"
	// CasbinModelGlob uses glob pattern matching (e.g., /api/users/*/orders/*).
	CasbinModelGlob CasbinModelType = "glob"
	// CasbinModelKeyMatch uses Casbin's keyMatch (e.g., /api/:resource).
	CasbinModelKeyMatch CasbinModelType = "key_match"
	// CasbinModelKeyMatch2 uses Casbin's keyMatch2 with regex support.
	CasbinModelKeyMatch2 CasbinModelType = "key_match2"
)

// CasbinAuthorizer authorizes requests with internal casbin engine.
type CasbinAuthorizer struct {
	enforcer  *casbin.Enforcer
	modelType CasbinModelType
	logger    *slog.Logger
}

// CasbinAuthorizerOption configures a CasbinAuthorizer.
type CasbinAuthorizerOption func(*casbinAuthorizerConfig)

type casbinAuthorizerConfig struct {
	modelType   CasbinModelType
	customModel string
	logger      *slog.Logger
}

// WithCasbinModelType sets the Casbin model type.
func WithCasbinModelType(modelType CasbinModelType) CasbinAuthorizerOption {
	return func(cfg *casbinAuthorizerConfig) {
		cfg.modelType = modelType
	}
}

// WithCasbinCustomModel sets a custom Casbin model string.
func WithCasbinCustomModel(model string) CasbinAuthorizerOption {
	return func(cfg *casbinAuthorizerConfig) {
		cfg.customModel = model
	}
}

// WithCasbinLogger sets the logger for the authorizer.
func WithCasbinLogger(logger *slog.Logger) CasbinAuthorizerOption {
	return func(cfg *casbinAuthorizerConfig) {
		cfg.logger = logger
	}
}

// NewCasbinAuthorizer creates a casbin authorizer with configurable model.
func NewCasbinAuthorizer(opts ...CasbinAuthorizerOption) (*CasbinAuthorizer, error) {
	cfg := casbinAuthorizerConfig{
		modelType: CasbinModelExact,
		logger:    normalizeLogger(nil).With("component", "authx.authorizer", "name", "casbin"),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	var m model.Model
	var err error

	if cfg.customModel != "" {
		m, err = model.NewModelFromString(cfg.customModel)
	} else {
		m, err = model.NewModelFromString(buildCasbinModel(cfg.modelType))
	}

	if err != nil {
		return nil, fmt.Errorf("%w: build casbin model: %v", ErrInvalidAuthorizer, err)
	}

	enforcer, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("%w: build casbin enforcer: %v", ErrInvalidAuthorizer, err)
	}

	return &CasbinAuthorizer{
		enforcer:  enforcer,
		modelType: cfg.modelType,
		logger:    cfg.logger,
	}, nil
}

// SetLogger sets slog logger for this authorizer.
func (a *CasbinAuthorizer) SetLogger(logger *slog.Logger) {
	if a == nil {
		return
	}
	a.logger = normalizeLogger(logger).With("component", "authx.authorizer", "name", "casbin", "model", a.modelType)
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
	a.logger.Debug("load permissions started", "rules", len(rules), "model", a.modelType)

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
	a.logger.Info("load permissions succeeded", "rules", len(policies), "model", a.modelType)
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
	a.logger.Debug("enforce finished", "subject", subject, "action", request.Action, "resource", request.Resource, "allowed", allowed, "model", a.modelType)

	return lo.Ternary(
		allowed,
		Allow("allowed by loaded policy"),
		Deny("denied by loaded policy"),
	), nil
}

// ModelType returns the current Casbin model type.
func (a *CasbinAuthorizer) ModelType() CasbinModelType {
	if a == nil {
		return CasbinModelExact
	}
	return a.modelType
}

// buildCasbinModel returns the Casbin model string for the specified type.
func buildCasbinModel(modelType CasbinModelType) string {
	switch modelType {
	case CasbinModelPrefix:
		return `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, eft

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = (r.sub == p.sub || g(r.sub, p.sub)) && keyMatch(r.obj, p.obj) && r.act == p.act
`
	case CasbinModelGlob:
		return `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, eft

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = (r.sub == p.sub || g(r.sub, p.sub)) && globMatch(r.obj, p.obj) && r.act == p.act
`
	case CasbinModelKeyMatch:
		return `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, eft

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = (r.sub == p.sub || g(r.sub, p.sub)) && keyMatch(r.obj, p.obj) && r.act == p.act
`
	case CasbinModelKeyMatch2:
		return `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, eft

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = (r.sub == p.sub || g(r.sub, p.sub)) && keyMatch2(r.obj, p.obj) && r.act == p.act
`
	default:
		// Exact match (default)
		return `
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
	}
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
