package authx

import (
	"context"
	"time"
)

// Diagnostics represents the runtime diagnostic information of authx manager.
type Diagnostics struct {
	// AuthorizerType is the type of the configured authorizer.
	AuthorizerType string
	// AuthenticatorType is the type of the configured authenticator.
	AuthenticatorType string
	// PolicyVersion is the current hot policy version.
	PolicyVersion int64
	// PolicySources is the list of configured policy source names.
	PolicySources []string
	// IdentityProviders is the list of configured identity provider names.
	IdentityProviders []string
	// LastReloadTime is the time of the last policy reload.
	LastReloadTime time.Time
	// LastReloadResult is the result of the last policy reload (success/failed).
	LastReloadResult string
	// LastReloadError is the error message of the last failed reload.
	LastReloadError string
	// TotalPermissions is the total number of loaded permission rules.
	TotalPermissions int
	// TotalRoleBindings is the total number of loaded role bindings.
	TotalRoleBindings int
}

// DiagnosticsProvider provides diagnostic information.
type DiagnosticsProvider interface {
	Diagnostics(ctx context.Context) Diagnostics
}

// DiagnosticsOption configures diagnostics tracking.
type DiagnosticsOption func(*diagnosticsConfig)

type diagnosticsConfig struct {
	trackReloadHistory bool
	maxHistorySize     int
}

// WithReloadHistory enables reload history tracking.
func WithReloadHistory(maxSize int) DiagnosticsOption {
	return func(cfg *diagnosticsConfig) {
		cfg.trackReloadHistory = true
		cfg.maxHistorySize = maxSize
	}
}

// ReloadHistoryEntry represents a single reload history entry.
type ReloadHistoryEntry struct {
	Timestamp time.Time
	Version   int64
	Result    string
	Error     string
}

// DiagnosticsTracker tracks diagnostic information for manager.
type DiagnosticsTracker struct {
	config            diagnosticsConfig
	policyVersion     int64
	lastReloadTime    time.Time
	lastReloadResult  string
	lastReloadError   string
	totalPermissions  int
	totalRoleBindings int
	history           []ReloadHistoryEntry
}

// NewDiagnosticsTracker creates a new diagnostics tracker.
func NewDiagnosticsTracker(opts ...DiagnosticsOption) *DiagnosticsTracker {
	cfg := diagnosticsConfig{
		trackReloadHistory: false,
		maxHistorySize:     10,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	return &DiagnosticsTracker{
		config:  cfg,
		history: make([]ReloadHistoryEntry, 0),
	}
}

// RecordReloadSuccess records a successful policy reload.
func (t *DiagnosticsTracker) RecordReloadSuccess(version int64, permissions, roleBindings int) {
	if t == nil {
		return
	}
	t.policyVersion = version
	t.lastReloadTime = time.Now()
	t.lastReloadResult = "success"
	t.lastReloadError = ""
	t.totalPermissions = permissions
	t.totalRoleBindings = roleBindings

	if t.config.trackReloadHistory {
		entry := ReloadHistoryEntry{
			Timestamp: t.lastReloadTime,
			Version:   version,
			Result:    "success",
		}
		t.history = append(t.history, entry)
		if len(t.history) > t.config.maxHistorySize {
			t.history = t.history[1:]
		}
	}
}

// RecordReloadFailure records a failed policy reload.
func (t *DiagnosticsTracker) RecordReloadFailure(err error) {
	if t == nil {
		return
	}
	t.lastReloadTime = time.Now()
	t.lastReloadResult = "failed"
	t.lastReloadError = err.Error()

	if t.config.trackReloadHistory {
		entry := ReloadHistoryEntry{
			Timestamp: t.lastReloadTime,
			Version:   t.policyVersion,
			Result:    "failed",
			Error:     err.Error(),
		}
		t.history = append(t.history, entry)
		if len(t.history) > t.config.maxHistorySize {
			t.history = t.history[1:]
		}
	}
}

// PolicyVersion returns the current policy version.
func (t *DiagnosticsTracker) PolicyVersion() int64 {
	if t == nil {
		return 0
	}
	return t.policyVersion
}

// ReloadHistory returns the reload history.
func (t *DiagnosticsTracker) ReloadHistory() []ReloadHistoryEntry {
	if t == nil {
		return nil
	}
	result := make([]ReloadHistoryEntry, len(t.history))
	copy(result, t.history)
	return result
}

// BuildDiagnostics builds diagnostics with provided component info.
func (t *DiagnosticsTracker) BuildDiagnostics(authorizerType, authenticatorType string, policySources, identityProviders []string) Diagnostics {
	if t == nil {
		return Diagnostics{}
	}
	return Diagnostics{
		AuthorizerType:    authorizerType,
		AuthenticatorType: authenticatorType,
		PolicyVersion:     t.policyVersion,
		PolicySources:     policySources,
		IdentityProviders: identityProviders,
		LastReloadTime:    t.lastReloadTime,
		LastReloadResult:  t.lastReloadResult,
		LastReloadError:   t.lastReloadError,
		TotalPermissions:  t.totalPermissions,
		TotalRoleBindings: t.totalRoleBindings,
	}
}
