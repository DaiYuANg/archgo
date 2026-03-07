package authx

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FilePolicySource loads policies from a file and supports hot reloading
// via fsnotify.
//
// Supported formats:
//   - JSON (default)
//
// File format example:
//
//	{
//	  "permissions": [
//	    {"subject": "alice", "resource": "/api/users", "action": "read", "allowed": true},
//	    {"subject": "bob", "resource": "/api/admin", "action": "write", "allowed": false}
//	  ],
//	  "role_bindings": [
//	    {"subject": "charlie", "role": "admin"},
//	    {"subject": "diana", "role": "viewer"}
//	  ]
//	}
type FilePolicySource struct {
	mu         sync.RWMutex
	path       string
	snapshot   PolicySnapshot
	version    int64
	name       string
	lastMod    time.Time
	reloadHook func(context.Context, PolicySnapshot) (PolicySnapshot, error)
	watcher    *fsnotify.Watcher
	stopCh     chan struct{}
	lastError  error
	logger     interface {
		Debug(string, ...any)
		Info(string, ...any)
		Warn(string, ...any)
		Error(string, ...any)
	}
}

// FilePolicySourceConfig configures a file policy source.
type FilePolicySourceConfig struct {
	// Path is the path to the policy file.
	Path string
	// Name is the optional name for this policy source.
	// Defaults to "file:<path>" if empty.
	Name string
	// ReloadHook is called after each successful reload.
	// It can transform the loaded snapshot before it's stored.
	ReloadHook func(context.Context, PolicySnapshot) (PolicySnapshot, error)
	// Logger for debug/info/warn/error logging.
	// If nil, no logging is performed.
	Logger interface {
		Debug(string, ...any)
		Info(string, ...any)
		Warn(string, ...any)
		Error(string, ...any)
	}
}

// NewFilePolicySource creates a new file policy source.
// It loads the initial snapshot from the file.
func NewFilePolicySource(cfg FilePolicySourceConfig) (*FilePolicySource, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("%w: file path is required", ErrInvalidPolicy)
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("%w: resolve file path: %v", ErrInvalidPolicy, err)
	}

	// Check file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: policy file not found: %s", ErrInvalidPolicy, absPath)
	}

	name := cfg.Name
	if name == "" {
		name = "file:" + absPath
	}

	s := &FilePolicySource{
		path:       absPath,
		version:    0,
		name:       name,
		reloadHook: cfg.ReloadHook,
		logger:     cfg.Logger,
		stopCh:     make(chan struct{}),
	}

	// Load initial snapshot
	if err := s.loadFromFile(context.Background()); err != nil {
		return nil, err
	}

	return s, nil
}

// LoadPolicies returns the current cached snapshot.
// The snapshot is refreshed automatically via fsnotify.
func (s *FilePolicySource) LoadPolicies(ctx context.Context) (PolicySnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.lastError != nil {
		return PolicySnapshot{}, s.lastError
	}

	return PolicySnapshot{
		Permissions:  slicesClone(s.snapshot.Permissions),
		RoleBindings: slicesClone(s.snapshot.RoleBindings),
	}, nil
}

// StartWatching starts watching the file for changes.
// Call this after creating the source to enable hot reloading.
func (s *FilePolicySource) StartWatching() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.watcher != nil {
		return nil // Already watching
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create fsnotify watcher: %w", err)
	}

	// Watch the directory containing the file
	dir := filepath.Dir(s.path)
	if err := watcher.Add(dir); err != nil {
		_ = watcher.Close()
		return fmt.Errorf("add directory to watcher: %w", err)
	}

	s.watcher = watcher

	go s.watchLoop()

	if s.logger != nil {
		s.logger.Info("file policy source started watching", "path", s.path)
	}

	return nil
}

// StopWatching stops watching the file for changes.
func (s *FilePolicySource) StopWatching() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.watcher == nil {
		return nil
	}

	close(s.stopCh)

	if err := s.watcher.Close(); err != nil {
		return fmt.Errorf("close watcher: %w", err)
	}

	s.watcher = nil

	if s.logger != nil {
		s.logger.Info("file policy source stopped watching", "path", s.path)
	}

	return nil
}

// Reload manually reloads the policy from the file.
// This bypasses the modification time check and forces a reload.
func (s *FilePolicySource) Reload(ctx context.Context) error {
	return s.loadFromFileForce(ctx)
}

// Version returns the current policy version.
func (s *FilePolicySource) Version() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.version
}

// LastError returns the last error encountered during reload.
func (s *FilePolicySource) LastError() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastError
}

// Name returns the policy source name.
func (s *FilePolicySource) Name() string {
	return s.name
}

// Path returns the policy file path.
func (s *FilePolicySource) Path() string {
	return s.path
}

func (s *FilePolicySource) watchLoop() {
	for {
		select {
		case <-s.stopCh:
			return
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			// Check if the event is for our file
			if filepath.Clean(event.Name) == s.path {
				// On Windows, file writes often trigger Remove and Create events
				// On Unix, typically Write events
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
					// Small delay to ensure file is fully written
					time.Sleep(100 * time.Millisecond)
					if err := s.loadFromFile(context.Background()); err != nil {
						if s.logger != nil {
							s.logger.Error("file policy source reload failed", "path", s.path, "error", err)
						}
					} else if s.logger != nil {
						s.logger.Debug("file policy source reloaded", "path", s.path, "version", s.version)
					}
				}
			}
		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			if s.logger != nil {
				s.logger.Error("file policy source watcher error", "error", err)
			}
		}
	}
}

func (s *FilePolicySource) loadFromFile(ctx context.Context) error {
	_ = ctx

	file, err := os.Open(s.path)
	if err != nil {
		s.mu.Lock()
		s.lastError = err
		s.mu.Unlock()
		return fmt.Errorf("open policy file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Get file info for modification time
	info, err := file.Stat()
	if err != nil {
		s.mu.Lock()
		s.lastError = err
		s.mu.Unlock()
		return fmt.Errorf("stat policy file: %w", err)
	}

	// Skip if file hasn't changed
	s.mu.RLock()
	lastMod := s.lastMod
	s.mu.RUnlock()

	if !info.ModTime().After(lastMod) && s.version > 0 {
		return nil // File unchanged
	}

	var data filePolicyData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		s.mu.Lock()
		s.lastError = err
		s.mu.Unlock()
		return fmt.Errorf("decode policy file: %w", err)
	}

	snapshot := PolicySnapshot(data)

	// Apply reload hook if present
	if s.reloadHook != nil {
		snapshot, err = s.reloadHook(context.Background(), snapshot)
		if err != nil {
			s.mu.Lock()
			s.lastError = err
			s.mu.Unlock()
			return fmt.Errorf("reload hook failed: %w", err)
		}
	}

	s.mu.Lock()
	s.snapshot = snapshot
	s.version++
	s.lastMod = info.ModTime()
	s.lastError = nil
	s.mu.Unlock()

	return nil
}

// loadFromFileForce forces a reload without checking modification time.
func (s *FilePolicySource) loadFromFileForce(ctx context.Context) error {
	_ = ctx

	file, err := os.Open(s.path)
	if err != nil {
		s.mu.Lock()
		s.lastError = err
		s.mu.Unlock()
		return fmt.Errorf("open policy file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Get file info for modification time
	info, err := file.Stat()
	if err != nil {
		s.mu.Lock()
		s.lastError = err
		s.mu.Unlock()
		return fmt.Errorf("stat policy file: %w", err)
	}

	var data filePolicyData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		s.mu.Lock()
		s.lastError = err
		s.mu.Unlock()
		return fmt.Errorf("decode policy file: %w", err)
	}

	snapshot := PolicySnapshot(data)

	// Apply reload hook if present
	if s.reloadHook != nil {
		snapshot, err = s.reloadHook(context.Background(), snapshot)
		if err != nil {
			s.mu.Lock()
			s.lastError = err
			s.mu.Unlock()
			return fmt.Errorf("reload hook failed: %w", err)
		}
	}

	s.mu.Lock()
	s.snapshot = snapshot
	s.version++
	s.lastMod = info.ModTime()
	s.lastError = nil
	s.mu.Unlock()

	return nil
}

// filePolicyData is the JSON structure for file-based policies.
type filePolicyData struct {
	Permissions  []PermissionRule `json:"permissions"`
	RoleBindings []RoleBinding    `json:"role_bindings"`
}

// Ensure FilePolicySource implements PolicySource.
var _ PolicySource = (*FilePolicySource)(nil)

// JSONPolicyFile is a helper for creating policy JSON files programmatically.
type JSONPolicyFile struct {
	Permissions  []PermissionRule `json:"permissions"`
	RoleBindings []RoleBinding    `json:"role_bindings"`
}

// WriteToFile writes the policy data to a JSON file.
func (p *JSONPolicyFile) WriteToFile(path string) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// ReadFromFile reads policy data from a JSON file.
func ReadPolicyFile(path string) (*JSONPolicyFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var policy JSONPolicyFile
	if err := json.Unmarshal(data, &policy); err != nil {
		return nil, err
	}

	return &policy, nil
}
