package authx

import (
	"context"
	"fmt"
	"sync"
)

// PolicySourceChain loads policies from multiple sources and merges them.
// Sources are loaded in order, and the resulting snapshots are merged
// using the configured PolicyMerger.
//
// Use cases:
//   - Combine policies from multiple sources (file + database + remote)
//   - Layered policy configuration (base policies + environment-specific overrides)
//   - Fallback chain (try primary source, fall back to secondary)
type PolicySourceChain struct {
	mu      sync.RWMutex
	sources []PolicySource
	merger  PolicyMerger
	name    string
	logger  interface {
		Debug(string, ...any)
		Info(string, ...any)
		Warn(string, ...any)
		Error(string, ...any)
	}
}

// PolicySourceChainConfig configures a policy source chain.
type PolicySourceChainConfig struct {
	// Sources is the list of policy sources to chain.
	// Sources are loaded in order during LoadPolicies.
	Sources []PolicySource
	// Merger is used to merge snapshots from multiple sources.
	// If nil, uses DefaultPolicyMerger.
	Merger PolicyMerger
	// Name is the optional name for this chain.
	// Defaults to "chain" if empty.
	Name string
}

// NewPolicySourceChain creates a new policy source chain.
func NewPolicySourceChain(cfg PolicySourceChainConfig) *PolicySourceChain {
	name := cfg.Name
	if name == "" {
		name = "chain"
	}

	merger := cfg.Merger
	if merger == nil {
		merger = NewDefaultPolicyMerger()
	}

	// Filter out nil sources
	sources := make([]PolicySource, 0, len(cfg.Sources))
	for _, src := range cfg.Sources {
		if src != nil {
			sources = append(sources, src)
		}
	}

	return &PolicySourceChain{
		sources: sources,
		merger:  merger,
		name:    name,
	}
}

// LoadPolicies loads from all sources and merges the results.
// Sources are loaded in order, and snapshots are merged using the configured merger.
func (c *PolicySourceChain) LoadPolicies(ctx context.Context) (PolicySnapshot, error) {
	c.mu.RLock()
	sources := c.sources
	merger := c.merger
	c.mu.RUnlock()

	if len(sources) == 0 {
		return PolicySnapshot{
			Permissions:  make([]PermissionRule, 0),
			RoleBindings: make([]RoleBinding, 0),
		}, nil
	}

	// Load from all sources
	snapshots := make([]PolicySnapshot, 0, len(sources))
	for _, src := range sources {
		if src == nil {
			continue
		}
		snapshot, err := src.LoadPolicies(ctx)
		if err != nil {
			return PolicySnapshot{}, fmt.Errorf("load from source %q: %w", src.Name(), err)
		}
		snapshots = append(snapshots, snapshot)
	}

	// Merge all snapshots
	return merger.Merge(ctx, snapshots...)
}

// AddSource appends a source to the chain.
// Returns the new source count.
func (c *PolicySourceChain) AddSource(src PolicySource) int {
	if src == nil {
		return len(c.sources)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.sources = append(c.sources, src)
	return len(c.sources)
}

// PrependSource inserts a source at the beginning of the chain.
// Higher priority sources should be prepended.
func (c *PolicySourceChain) PrependSource(src PolicySource) {
	if src == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.sources = append([]PolicySource{src}, c.sources...)
}

// RemoveSource removes a source by name.
// Returns true if a source was removed.
func (c *PolicySourceChain) RemoveSource(name string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, src := range c.sources {
		if src != nil && src.Name() == name {
			c.sources = append(c.sources[:i], c.sources[i+1:]...)
			return true
		}
	}
	return false
}

// Sources returns a copy of the source list.
func (c *PolicySourceChain) Sources() []PolicySource {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]PolicySource, len(c.sources))
	copy(result, c.sources)
	return result
}

// SourceCount returns the number of sources in the chain.
func (c *PolicySourceChain) SourceCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.sources)
}

// SetMerger replaces the policy merger.
func (c *PolicySourceChain) SetMerger(merger PolicyMerger) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if merger != nil {
		c.merger = merger
	}
}

// Name returns the chain name.
func (c *PolicySourceChain) Name() string {
	return c.name
}

// SetLogger sets the logger for the policy source chain.
func (c *PolicySourceChain) SetLogger(logger interface {
	Debug(string, ...any)
	Info(string, ...any)
	Warn(string, ...any)
	Error(string, ...any)
}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger = logger
}

// Ensure PolicySourceChain implements PolicySource.
var _ PolicySource = (*PolicySourceChain)(nil)

// FallbackPolicySource tries sources in order until one succeeds.
// This is useful for primary/backup source configurations.
type FallbackPolicySource struct {
	mu      sync.RWMutex
	sources []PolicySource
	name    string
}

// FallbackPolicySourceConfig configures a fallback policy source.
type FallbackPolicySourceConfig struct {
	// Sources is the ordered list of fallback sources.
	// The first successful source is used.
	Sources []PolicySource
	// Name is the optional name for this fallback source.
	// Defaults to "fallback" if empty.
	Name string
}

// NewFallbackPolicySource creates a new fallback policy source.
func NewFallbackPolicySource(cfg FallbackPolicySourceConfig) *FallbackPolicySource {
	name := cfg.Name
	if name == "" {
		name = "fallback"
	}

	// Filter out nil sources
	sources := make([]PolicySource, 0, len(cfg.Sources))
	for _, src := range cfg.Sources {
		if src != nil {
			sources = append(sources, src)
		}
	}

	return &FallbackPolicySource{
		sources: sources,
		name:    name,
	}
}

// LoadPolicies tries each source in order until one succeeds.
// Returns an error if all sources fail.
func (s *FallbackPolicySource) LoadPolicies(ctx context.Context) (PolicySnapshot, error) {
	s.mu.RLock()
	sources := s.sources
	s.mu.RUnlock()

	if len(sources) == 0 {
		return PolicySnapshot{
			Permissions:  make([]PermissionRule, 0),
			RoleBindings: make([]RoleBinding, 0),
		}, nil
	}

	var lastErr error
	for _, src := range sources {
		if src == nil {
			continue
		}
		snapshot, err := src.LoadPolicies(ctx)
		if err == nil {
			return snapshot, nil
		}
		lastErr = err
	}

	return PolicySnapshot{}, fmt.Errorf("%w: all fallback sources failed, last error: %v", ErrProviderUnavailable, lastErr)
}

// AddSource appends a fallback source.
func (s *FallbackPolicySource) AddSource(src PolicySource) {
	if src == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sources = append(s.sources, src)
}

// Name returns the fallback source name.
func (s *FallbackPolicySource) Name() string {
	return s.name
}

// Ensure FallbackPolicySource implements PolicySource.
var _ PolicySource = (*FallbackPolicySource)(nil)

// ConditionalPolicySource selects a source based on a condition function.
type ConditionalPolicySource struct {
	mu        sync.RWMutex
	condition func(context.Context) PolicySource
	sources   []PolicySource // All possible sources for introspection
	name      string
}

// ConditionalPolicySourceConfig configures a conditional policy source.
type ConditionalPolicySourceConfig struct {
	// Condition returns the selected source based on context.
	// It is called during each LoadPolicies call.
	Condition func(context.Context) PolicySource
	// Sources is the list of all possible sources (for introspection).
	Sources []PolicySource
	// Name is the optional name for this source.
	// Defaults to "conditional" if empty.
	Name string
}

// NewConditionalPolicySource creates a new conditional policy source.
func NewConditionalPolicySource(cfg ConditionalPolicySourceConfig) *ConditionalPolicySource {
	name := cfg.Name
	if name == "" {
		name = "conditional"
	}

	return &ConditionalPolicySource{
		condition: cfg.Condition,
		sources:   cfg.Sources,
		name:      name,
	}
}

// LoadPolicies selects a source using the condition function and loads from it.
func (s *ConditionalPolicySource) LoadPolicies(ctx context.Context) (PolicySnapshot, error) {
	s.mu.RLock()
	condition := s.condition
	s.mu.RUnlock()

	if condition == nil {
		return PolicySnapshot{}, fmt.Errorf("%w: no condition function configured", ErrInvalidPolicy)
	}

	source := condition(ctx)
	if source == nil {
		return PolicySnapshot{}, fmt.Errorf("%w: condition returned nil source", ErrProviderUnavailable)
	}

	return source.LoadPolicies(ctx)
}

// Name returns the conditional source name.
func (s *ConditionalPolicySource) Name() string {
	return s.name
}

// Sources returns all possible sources.
func (s *ConditionalPolicySource) Sources() []PolicySource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]PolicySource, len(s.sources))
	copy(result, s.sources)
	return result
}

// Ensure ConditionalPolicySource implements PolicySource.
var _ PolicySource = (*ConditionalPolicySource)(nil)

// CachedPolicySource wraps another source and caches the result.
// Cache is invalidated based on the configured TTL.
type CachedPolicySource struct {
	mu      sync.RWMutex
	wrapped PolicySource
	cache   PolicySnapshot
	name    string
}

// CachedPolicySourceConfig configures a cached policy source.
type CachedPolicySourceConfig struct {
	// Wrapped is the underlying source to cache.
	Wrapped PolicySource
	// Name is the optional name for this cached source.
	// Defaults to "cached:<wrapped_name>" if empty.
	Name string
}

// NewCachedPolicySource creates a new cached policy source.
// Note: This is a simplified implementation. For production use,
// consider adding proper TTL-based invalidation with time.Duration.
func NewCachedPolicySource(cfg CachedPolicySourceConfig) *CachedPolicySource {
	wrapped := cfg.Wrapped
	if wrapped == nil {
		panic("wrapped source is required")
	}

	name := cfg.Name
	if name == "" {
		name = "cached:" + wrapped.Name()
	}

	return &CachedPolicySource{
		wrapped: wrapped,
		name:    name,
	}
}

// LoadPolicies returns cached result or loads from wrapped source.
// Note: Current implementation caches indefinitely.
// TODO: Add proper TTL-based invalidation.
func (s *CachedPolicySource) LoadPolicies(ctx context.Context) (PolicySnapshot, error) {
	s.mu.RLock()
	if s.cache.Permissions != nil || s.cache.RoleBindings != nil {
		cache := s.cache
		s.mu.RUnlock()
		return PolicySnapshot{
			Permissions:  slicesClone(cache.Permissions),
			RoleBindings: slicesClone(cache.RoleBindings),
		}, nil
	}
	s.mu.RUnlock()

	// Cache miss, load from wrapped source
	snapshot, err := s.wrapped.LoadPolicies(ctx)
	if err != nil {
		return PolicySnapshot{}, err
	}

	s.mu.Lock()
	s.cache = PolicySnapshot{
		Permissions:  slicesClone(snapshot.Permissions),
		RoleBindings: slicesClone(snapshot.RoleBindings),
	}
	s.mu.Unlock()

	return snapshot, nil
}

// Invalidate clears the cache.
func (s *CachedPolicySource) Invalidate() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache = PolicySnapshot{}
}

// Refresh forces a refresh from the wrapped source.
func (s *CachedPolicySource) Refresh(ctx context.Context) (PolicySnapshot, error) {
	snapshot, err := s.wrapped.LoadPolicies(ctx)
	if err != nil {
		return PolicySnapshot{}, err
	}

	s.mu.Lock()
	s.cache = PolicySnapshot{
		Permissions:  slicesClone(snapshot.Permissions),
		RoleBindings: slicesClone(snapshot.RoleBindings),
	}
	s.mu.Unlock()

	return snapshot, nil
}

// Name returns the cached source name.
func (s *CachedPolicySource) Name() string {
	return s.name
}

// Wrapped returns the underlying wrapped source.
func (s *CachedPolicySource) Wrapped() PolicySource {
	return s.wrapped
}

// Ensure CachedPolicySource implements PolicySource.
var _ PolicySource = (*CachedPolicySource)(nil)
