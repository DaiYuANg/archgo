package authx

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/DaiYuANg/arcgo/observability"
)

// IdentityProvider is business-side identity loading contract.
type IdentityProvider interface {
	LoadByPrincipal(ctx context.Context, principal string) (UserDetails, error)
}

// Manager is the high-level facade for authentication and authorization.
type Manager struct {
	flow           *AuthFlow
	userProviders  *identityProviderChain
	policySources  *PolicySourceChain
	logger         *slog.Logger
	obs            observability.Observability
	diagnostics    *DiagnosticsTracker
	eventPublisher *EventPublisher

	authorizer    atomic.Pointer[CasbinAuthorizer]
	policyVersion atomic.Int64
}

const (
	metricAuthenticateTotal      = "authx_authenticate_total"
	metricAuthenticateDurationMS = "authx_authenticate_duration_ms"
	metricAuthorizeTotal         = "authx_authorize_total"
	metricAuthorizeDurationMS    = "authx_authorize_duration_ms"
	metricPolicyReloadTotal      = "authx_policy_reload_total"
	metricPolicyReloadDurationMS = "authx_policy_reload_duration_ms"
)

func (m *Manager) loggerSafe() *slog.Logger {
	if m == nil {
		return normalizeLogger(nil).With("component", "authx.manager")
	}
	return normalizeLogger(m.logger)
}

func (m *Manager) observabilitySafe() observability.Observability {
	if m == nil {
		return observability.Nop()
	}
	return observability.Normalize(m.obs, m.logger)
}

// NewManager creates manager with spring-security-style option chain.
func NewManager(opts ...ManagerOption) (*Manager, error) {
	cfg := managerConfig{}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}

	if len(cfg.providers) == 0 {
		cfg.providers = []IdentityProvider{NewInMemoryIdentityProvider()}
	}
	if len(cfg.sources) == 0 {
		cfg.sources = []PolicySource{NewMemoryPolicySource(MemoryPolicySourceConfig{
			Name: "default-memory",
		})}
	}

	obs := observability.Normalize(cfg.observability, cfg.logger)
	logger := normalizeLogger(obs.Logger()).With("component", "authx.manager")

	userProviders, err := newIdentityProviderChain(cfg.providers...)
	if err != nil {
		return nil, err
	}
	userProviders.SetLogger(logger.With("node", "provider-chain"))

	passwordAuthenticator, err := NewAuthbossPasswordAuthenticator(userProviders)
	if err != nil {
		return nil, err
	}
	passwordAuthenticator.SetLogger(logger.With("node", "authenticator"))

	flow, err := NewAuthFlow(passwordAuthenticator)
	if err != nil {
		return nil, err
	}

	policySources := NewPolicySourceChain(PolicySourceChainConfig{
		Sources: cfg.sources,
		Name:    "manager-policy-chain",
	})
	policySources.SetLogger(logger.With("node", "policy-source-chain"))

	authorizer, err := NewCasbinAuthorizer()
	if err != nil {
		return nil, err
	}
	authorizer.SetLogger(logger.With("node", "authorizer"))

	manager := &Manager{
		flow:           flow,
		userProviders:  userProviders,
		policySources:  policySources,
		logger:         logger,
		obs:            obs,
		diagnostics:    NewDiagnosticsTracker(),
		eventPublisher: cfg.eventPublisher,
	}

	// Create default event publisher if not provided
	if manager.eventPublisher == nil {
		var publisherOpts []EventPublisherOption
		if cfg.logger != nil {
			publisherOpts = append(publisherOpts, WithEventPublisherLogger(cfg.logger))
		}
		if cfg.observability != nil {
			publisherOpts = append(publisherOpts, WithEventPublisherObservability(cfg.observability))
		}
		manager.eventPublisher = NewEventPublisher(publisherOpts...)
	}
	manager.authorizer.Store(authorizer)
	logger.Info("manager initialized", "providers", len(cfg.providers), "sources", len(cfg.sources))

	return manager, nil
}

// Authenticate authenticates credential and stores security context in returned context.
func (m *Manager) Authenticate(ctx context.Context, credential Credential) (context.Context, Authentication, error) {
	if m == nil || m.flow == nil {
		return ctx, Authentication{}, fmt.Errorf("%w: manager is not configured", ErrInvalidAuthenticator)
	}
	obs := m.observabilitySafe()
	credentialKind := ""
	if credential != nil {
		credentialKind = credential.Kind()
	}
	ctx, span := obs.StartSpan(ctx, "authx.authenticate",
		observability.String("auth.credential.kind", credentialKind),
	)
	defer span.End()

	start := time.Now()
	result := "success"
	defer func() {
		obs.AddCounter(ctx, metricAuthenticateTotal, 1,
			observability.String("result", result),
			observability.String("credential_kind", credentialKind),
		)
		obs.RecordHistogram(ctx, metricAuthenticateDurationMS, float64(time.Since(start).Milliseconds()),
			observability.String("result", result),
			observability.String("credential_kind", credentialKind),
		)
	}()

	logger := m.loggerSafe()
	logger.Debug("authenticate started", "credential_kind", credentialKind)

	identity, err := m.flow.Authenticate(ctx, credential)
	if err != nil {
		result = "error"
		span.RecordError(err)
		logger.Warn("authenticate failed", "credential_kind", credentialKind, "error", err.Error())
		return ctx, Authentication{}, err
	}

	authentication := NewAuthentication(identity, m.policyVersion.Load())
	securityContext := NewSecurityContext(authentication)
	logger.Info("authenticate succeeded", "principal_id", identity.ID(), "policy_version", authentication.PolicyVersion())

	return WithSecurityContext(ctx, securityContext), authentication, nil
}

// AuthenticatePassword authenticates username/password with password credential.
func (m *Manager) AuthenticatePassword(ctx context.Context, username, password string) (context.Context, Authentication, error) {
	return m.Authenticate(ctx, PasswordCredential{
		Username: username,
		Password: password,
	})
}

// Can authorizes current authenticated user in context.
func (m *Manager) Can(ctx context.Context, action, resource string) (bool, error) {
	if m == nil {
		return false, fmt.Errorf("%w: manager is nil", ErrInvalidAuthorizer)
	}

	obs := m.observabilitySafe()
	ctx, span := obs.StartSpan(ctx, "authx.authorize",
		observability.String("auth.action", action),
		observability.String("auth.resource", resource),
	)
	defer span.End()

	start := time.Now()
	result := "allow"
	defer func() {
		obs.AddCounter(ctx, metricAuthorizeTotal, 1,
			observability.String("result", result),
			observability.String("action", action),
		)
		obs.RecordHistogram(ctx, metricAuthorizeDurationMS, float64(time.Since(start).Milliseconds()),
			observability.String("result", result),
			observability.String("action", action),
		)
	}()

	logger := m.loggerSafe()
	authentication, err := RequireAuthentication(ctx)
	if err != nil {
		result = "error"
		span.RecordError(err)
		logger.Warn("authorize failed: no authentication in context", "action", action, "resource", resource)
		return false, err
	}

	authorizer := m.authorizer.Load()
	if authorizer == nil {
		result = "error"
		missingAuthorizerErr := fmt.Errorf("%w: manager authorizer is nil", ErrInvalidAuthorizer)
		span.RecordError(missingAuthorizerErr)
		return false, fmt.Errorf("%w: manager authorizer is nil", ErrInvalidAuthorizer)
	}

	decision, err := authorizer.Authorize(ctx, authentication.Identity(), NewRequest(action, resource, nil))
	if err != nil {
		result = "error"
		span.RecordError(err)
		logger.Warn("authorize failed", "principal_id", authentication.Identity().ID(), "action", action, "resource", resource, "error", err.Error())
		return false, err
	}
	if decision.Allowed {
		result = "allow"
	} else {
		result = "deny"
	}
	logger.Debug("authorize finished", "principal_id", authentication.Identity().ID(), "action", action, "resource", resource, "allowed", decision.Allowed)

	return decision.Allowed, nil
}

// SetIdentityProviders replaces current provider chain.
func (m *Manager) SetIdentityProviders(providers ...IdentityProvider) error {
	if m == nil || m.userProviders == nil {
		return fmt.Errorf("%w: manager is not configured", ErrInvalidAuthenticator)
	}
	logger := m.loggerSafe()
	if err := m.userProviders.Set(providers...); err != nil {
		return err
	}
	logger.Info("identity providers replaced", "providers", len(providers))
	return nil
}

// AddIdentityProvider appends one provider to current provider chain.
func (m *Manager) AddIdentityProvider(provider IdentityProvider) error {
	if m == nil || m.userProviders == nil {
		return fmt.Errorf("%w: manager is not configured", ErrInvalidAuthenticator)
	}
	logger := m.loggerSafe()
	if err := m.userProviders.Add(provider); err != nil {
		return err
	}
	logger.Info("identity provider added")
	return nil
}

// SetPolicySources replaces policy source chain.
func (m *Manager) SetPolicySources(sources ...PolicySource) error {
	if m == nil || m.policySources == nil {
		return fmt.Errorf("%w: manager is not configured", ErrInvalidPolicy)
	}
	logger := m.loggerSafe()

	// Clear existing sources and add new ones
	m.policySources.mu.Lock()
	m.policySources.sources = make([]PolicySource, 0, len(sources))
	for _, src := range sources {
		if src != nil {
			m.policySources.sources = append(m.policySources.sources, src)
		}
	}
	m.policySources.mu.Unlock()

	logger.Info("policy sources replaced", "sources", len(sources))
	return nil
}

// AddPolicySource appends one source to policy source chain.
func (m *Manager) AddPolicySource(source PolicySource) error {
	if m == nil || m.policySources == nil {
		return fmt.Errorf("%w: manager is not configured", ErrInvalidPolicy)
	}
	logger := m.loggerSafe()
	m.policySources.AddSource(source)
	logger.Info("policy source added")
	return nil
}

// LoadPolicies loads and merges snapshots from all configured sources.
func (m *Manager) LoadPolicies(ctx context.Context) (int64, error) {
	if m == nil {
		return 0, fmt.Errorf("%w: manager is nil", ErrInvalidAuthorizer)
	}
	obs := m.observabilitySafe()
	ctx, span := obs.StartSpan(ctx, "authx.policy.load")
	defer span.End()

	start := time.Now()
	result := "success"
	defer func() {
		obs.AddCounter(ctx, metricPolicyReloadTotal, 1,
			observability.String("operation", "load"),
			observability.String("result", result),
		)
		obs.RecordHistogram(ctx, metricPolicyReloadDurationMS, float64(time.Since(start).Milliseconds()),
			observability.String("operation", "load"),
			observability.String("result", result),
		)
	}()

	logger := m.loggerSafe()

	sources := m.policySources.Sources()
	if len(sources) == 0 {
		result = "error"
		emptySourcesErr := fmt.Errorf("%w: policy sources are empty", ErrInvalidPolicy)
		span.RecordError(emptySourcesErr)
		return 0, fmt.Errorf("%w: policy sources are empty", ErrInvalidPolicy)
	}
	logger.Info("load policies started", "sources", len(sources))

	merged := PolicySnapshot{
		Permissions:  make([]PermissionRule, 0),
		RoleBindings: make([]RoleBinding, 0),
	}

	for _, source := range sources {
		snapshot, err := source.LoadPolicies(ctx)
		if err != nil {
			result = "error"
			span.RecordError(err)
			logger.Error("load policies failed", "error", err.Error())
			return 0, err
		}

		merged.Permissions = append(merged.Permissions, snapshot.Permissions...)
		merged.RoleBindings = append(merged.RoleBindings, snapshot.RoleBindings...)
	}
	version, err := m.ReplacePolicies(ctx, merged)
	if err != nil {
		result = "error"
		span.RecordError(err)
		return 0, err
	}
	logger.Info("load policies succeeded", "version", version, "permission_rules", len(merged.Permissions), "role_bindings", len(merged.RoleBindings))
	return version, nil
}

// LoadPoliciesFrom loads latest policies from one source and applies them.
func (m *Manager) LoadPoliciesFrom(ctx context.Context, source PolicySource) (int64, error) {
	if m == nil {
		return 0, fmt.Errorf("%w: manager is nil", ErrInvalidAuthorizer)
	}
	if source == nil {
		return 0, fmt.Errorf("%w: policy source is nil", ErrInvalidPolicy)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	logger := m.loggerSafe()

	snapshot, err := source.LoadPolicies(ctx)
	if err != nil {
		logger.Error("load policies from source failed", "error", err.Error())
		return 0, err
	}

	version, err := m.ReplacePolicies(ctx, snapshot)
	if err != nil {
		return 0, err
	}

	if err := m.SetPolicySources(source); err != nil {
		return 0, err
	}
	logger.Info("load policies from source succeeded", "version", version, "permission_rules", len(snapshot.Permissions), "role_bindings", len(snapshot.RoleBindings))
	return version, nil
}

// ReplacePolicies replaces policies immediately and hot swaps authorizer atomically.
func (m *Manager) ReplacePolicies(ctx context.Context, snapshot PolicySnapshot) (int64, error) {
	if m == nil {
		return 0, fmt.Errorf("%w: manager is nil", ErrInvalidAuthorizer)
	}
	obs := m.observabilitySafe()
	ctx, span := obs.StartSpan(ctx, "authx.policy.replace")
	defer span.End()

	start := time.Now()
	result := "success"
	defer func() {
		obs.AddCounter(ctx, metricPolicyReloadTotal, 1,
			observability.String("operation", "replace"),
			observability.String("result", result),
		)
		obs.RecordHistogram(ctx, metricPolicyReloadDurationMS, float64(time.Since(start).Milliseconds()),
			observability.String("operation", "replace"),
			observability.String("result", result),
		)
	}()

	logger := m.loggerSafe()

	nextAuthorizer, err := NewCasbinAuthorizer()
	if err != nil {
		result = "error"
		span.RecordError(err)
		return 0, err
	}
	nextAuthorizer.SetLogger(logger.With("node", "authorizer"))

	copiedSnapshot := snapshot.clone()
	logger.Debug("replace policies started", "permission_rules", len(copiedSnapshot.Permissions), "role_bindings", len(copiedSnapshot.RoleBindings))
	if err := nextAuthorizer.LoadPermissions(ctx, copiedSnapshot.Permissions...); err != nil {
		result = "error"
		span.RecordError(err)
		m.diagnostics.RecordReloadFailure(err)
		logger.Error("replace policies failed: load permissions", "error", err.Error())
		return 0, err
	}
	if err := nextAuthorizer.LoadRoleBindings(ctx, copiedSnapshot.RoleBindings...); err != nil {
		result = "error"
		span.RecordError(err)
		m.diagnostics.RecordReloadFailure(err)
		logger.Error("replace policies failed: load role bindings", "error", err.Error())
		return 0, err
	}

	m.authorizer.Store(nextAuthorizer)
	version := m.policyVersion.Add(1)

	// Record diagnostics
	m.diagnostics.RecordReloadSuccess(version, len(copiedSnapshot.Permissions), len(copiedSnapshot.RoleBindings))

	logger.Info("replace policies succeeded", "version", version)
	return version, nil
}

// PolicyVersion returns current hot policy version.
func (m *Manager) PolicyVersion() int64 {
	if m == nil {
		return 0
	}
	return m.policyVersion.Load()
}

// EventPublisher returns the event publisher for publishing authx events.
// The returned publisher can be used to subscribe to events or publish custom events.
func (m *Manager) EventPublisher() *EventPublisher {
	if m == nil {
		return nil
	}
	return m.eventPublisher
}
