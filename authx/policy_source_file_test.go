package authx

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilePolicySource_InitialLoad(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "policies.json")

	policyData := &JSONPolicyFile{
		Permissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
			AllowPermission("bob", "/api/admin", "write"),
		},
		RoleBindings: []RoleBinding{
			NewRoleBinding("charlie", "admin"),
		},
	}
	err := policyData.WriteToFile(tmpFile)
	assert.NoError(t, err)

	src, err := NewFilePolicySource(FilePolicySourceConfig{
		Path: tmpFile,
		Name: "test-file",
	})
	assert.NoError(t, err)
	assert.Equal(t, "test-file", src.Name())
	assert.Equal(t, tmpFile, src.Path())
	assert.Equal(t, int64(1), src.Version())

	ctx := context.Background()
	snapshot, err := src.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot.Permissions, 2)
	assert.Len(t, snapshot.RoleBindings, 1)
}

func TestFilePolicySource_FileNotFound(t *testing.T) {
	_, err := NewFilePolicySource(FilePolicySourceConfig{
		Path: "/nonexistent/path/policies.json",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy file not found")
}

func TestFilePolicySource_EmptyPath(t *testing.T) {
	_, err := NewFilePolicySource(FilePolicySourceConfig{
		Path: "",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file path is required")
}

func TestFilePolicySource_Reload(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "policies.json")

	// Initial policy
	policyData := &JSONPolicyFile{
		Permissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
		},
	}
	err := policyData.WriteToFile(tmpFile)
	assert.NoError(t, err)

	src, err := NewFilePolicySource(FilePolicySourceConfig{
		Path: tmpFile,
	})
	assert.NoError(t, err)

	ctx := context.Background()
	snapshot1, err := src.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot1.Permissions, 1)

	// Modify file
	policyData = &JSONPolicyFile{
		Permissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
			AllowPermission("alice", "/api/users", "write"),
			AllowPermission("bob", "/api/admin", "read"),
		},
	}
	err = policyData.WriteToFile(tmpFile)
	assert.NoError(t, err)

	// Manual reload
	err = src.Reload(ctx)
	assert.NoError(t, err)

	snapshot2, err := src.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, snapshot2.Permissions, 3)
	assert.Equal(t, int64(2), src.Version())
}

func TestFilePolicySource_ReloadHook(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "policies.json")

	policyData := &JSONPolicyFile{
		Permissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
		},
	}
	err := policyData.WriteToFile(tmpFile)
	assert.NoError(t, err)

	hookCalled := int32(0)
	src, err := NewFilePolicySource(FilePolicySourceConfig{
		Path: tmpFile,
		ReloadHook: func(ctx context.Context, snapshot PolicySnapshot) (PolicySnapshot, error) {
			atomic.AddInt32(&hookCalled, 1)
			// Add extra permission via hook
			snapshot.Permissions = append(snapshot.Permissions, AllowPermission("system", "/api/health", "read"))
			return snapshot, nil
		},
	})
	assert.NoError(t, err)

	ctx := context.Background()
	snapshot, err := src.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&hookCalled))
	assert.Len(t, snapshot.Permissions, 2)
	assert.Equal(t, "system", snapshot.Permissions[1].Subject)
}

func TestFilePolicySource_InvalidJSON(t *testing.T) {
	// Create temp file with invalid JSON
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "policies.json")

	err := os.WriteFile(tmpFile, []byte(`{invalid json}`), 0o644)
	assert.NoError(t, err)

	_, err = NewFilePolicySource(FilePolicySourceConfig{
		Path: tmpFile,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode policy file")
}

func TestFilePolicySource_LastError(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "policies.json")

	policyData := &JSONPolicyFile{
		Permissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
		},
	}
	err := policyData.WriteToFile(tmpFile)
	assert.NoError(t, err)

	src, err := NewFilePolicySource(FilePolicySourceConfig{
		Path: tmpFile,
	})
	assert.NoError(t, err)

	// Initial load should succeed
	assert.NoError(t, src.LastError())

	// Delete the file to cause an error
	err = os.Remove(tmpFile)
	assert.NoError(t, err)

	// Reload should fail
	err = src.Reload(context.Background())
	assert.Error(t, err)
	assert.Error(t, src.LastError())
}

func TestReadPolicyFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "policies.json")

	expected := &JSONPolicyFile{
		Permissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
			DenyPermission("bob", "/api/admin", "write"),
		},
		RoleBindings: []RoleBinding{
			NewRoleBinding("charlie", "admin"),
			NewRoleBinding("diana", "viewer"),
		},
	}

	err := expected.WriteToFile(tmpFile)
	assert.NoError(t, err)

	actual, err := ReadPolicyFile(tmpFile)
	assert.NoError(t, err)
	assert.Equal(t, len(expected.Permissions), len(actual.Permissions))
	assert.Equal(t, len(expected.RoleBindings), len(actual.RoleBindings))
}

func TestFilePolicySource_NoChangeSkip(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "policies.json")

	policyData := &JSONPolicyFile{
		Permissions: []PermissionRule{
			AllowPermission("alice", "/api/users", "read"),
		},
	}
	err := policyData.WriteToFile(tmpFile)
	assert.NoError(t, err)

	src, err := NewFilePolicySource(FilePolicySourceConfig{
		Path: tmpFile,
	})
	assert.NoError(t, err)

	ctx := context.Background()
	_, err = src.LoadPolicies(ctx)
	assert.NoError(t, err)

	initialVersion := src.Version()

	// LoadPolicies without file change should not increment version
	_, err = src.LoadPolicies(ctx)
	assert.NoError(t, err)
	assert.Equal(t, initialVersion, src.Version())
}
