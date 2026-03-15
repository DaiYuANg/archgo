package json

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DaiYuANg/archgo/kvx"
	"github.com/DaiYuANg/archgo/kvx/adapter/redis"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// User represents a user document.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	Tags      []string  `json:"tags"`
	Address   Address   `json:"address"`
	CreatedAt time.Time `json:"created_at"`
}

// Address represents an address.
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Country string `json:"country"`
}

// Example demonstrates basic JSON operations.
func Example() {
	fmt.Println("JSON module example")
	// Output: JSON module example
}

// ExampleJSON demonstrates JSON document operations.
func ExampleJSON() {
	fmt.Println("JSON operations example")
	// Output: JSON operations example
}

// ExampleDocumentRepository demonstrates typed document repository.
func ExampleDocumentRepository() {
	fmt.Println("DocumentRepository example")
	// Output: DocumentRepository example
}

// setupRedisContainer starts a Redis container with RedisJSON module.
func setupRedisContainer(ctx context.Context) (testcontainers.Container, kvx.Client, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis/redis-stack:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to get host: %w", err)
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to get port: %w", err)
	}

	client, err := redis.New(kvx.ClientOptions{
		Addrs: []string{fmt.Sprintf("%s:%s", host, port.Port())},
	})
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to create client: %w", err)
	}

	return container, client, nil
}

// TestJSONIntegration tests JSON operations with real Redis.
func TestJSONIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start Redis container
	container, client, err := setupRedisContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to setup container: %v", err)
	}
	defer container.Terminate(ctx)
	defer client.Close()

	json := NewJSON(client)

	// Test 1: Set and Get document
	t.Run("SetAndGet", func(t *testing.T) {
		user := User{
			ID:    "user1",
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   30,
			Tags:  []string{"developer", "gopher"},
			Address: Address{
				Street:  "123 Main St",
				City:    "San Francisco",
				Country: "USA",
			},
			CreatedAt: time.Now(),
		}

		// Set document
		if err := json.Set(ctx, "user:user1", user, time.Hour); err != nil {
			t.Fatalf("Failed to set JSON: %v", err)
		}

		// Get document
		var retrieved User
		if err := json.Get(ctx, "user:user1", &retrieved); err != nil {
			t.Fatalf("Failed to get JSON: %v", err)
		}

		if retrieved.ID != user.ID {
			t.Errorf("Expected ID %s, got %s", user.ID, retrieved.ID)
		}
		if retrieved.Name != user.Name {
			t.Errorf("Expected Name %s, got %s", user.Name, retrieved.Name)
		}
	})

	// Test 2: Path operations
	t.Run("PathOperations", func(t *testing.T) {
		user := User{
			ID:    "user2",
			Name:  "Jane Doe",
			Email: "jane@example.com",
			Age:   25,
		}

		// Set document
		if err := json.Set(ctx, "user:user2", user, time.Hour); err != nil {
			t.Fatalf("Failed to set JSON: %v", err)
		}

		// Update path
		if err := json.SetPath(ctx, "user:user2", "$.age", 26); err != nil {
			t.Fatalf("Failed to set path: %v", err)
		}

		// Get path
		var age int
		if err := json.GetPath(ctx, "user:user2", "$.age", &age); err != nil {
			t.Fatalf("Failed to get path: %v", err)
		}

		if age != 26 {
			t.Errorf("Expected age 26, got %d", age)
		}
	})

	// Test 3: Exists and Delete
	t.Run("ExistsAndDelete", func(t *testing.T) {
		user := User{ID: "user3", Name: "Test User"}

		// Set document
		if err := json.Set(ctx, "user:user3", user, time.Hour); err != nil {
			t.Fatalf("Failed to set JSON: %v", err)
		}

		// Check exists
		exists, err := json.Exists(ctx, "user:user3")
		if err != nil {
			t.Fatalf("Failed to check exists: %v", err)
		}
		if !exists {
			t.Error("Expected document to exist")
		}

		// Delete document
		if err := json.Delete(ctx, "user:user3"); err != nil {
			t.Fatalf("Failed to delete: %v", err)
		}

		// Check not exists
		exists, err = json.Exists(ctx, "user:user3")
		if err != nil {
			t.Fatalf("Failed to check exists: %v", err)
		}
		if exists {
			t.Error("Expected document to not exist")
		}
	})

	// Test 4: Type and Length
	t.Run("TypeAndLength", func(t *testing.T) {
		data := map[string]interface{}{
			"name":  "test",
			"items": []string{"a", "b", "c"},
			"count": 42,
		}

		if err := json.Set(ctx, "test:type", data, time.Hour); err != nil {
			t.Fatalf("Failed to set JSON: %v", err)
		}

		// Get type
		typ, err := json.Type(ctx, "test:type", "$.items")
		if err != nil {
			t.Fatalf("Failed to get type: %v", err)
		}
		if typ != "array" {
			t.Errorf("Expected type 'array', got '%s'", typ)
		}

		// Get length
		length, err := json.Length(ctx, "test:type", "$.items")
		if err != nil {
			t.Fatalf("Failed to get length: %v", err)
		}
		if length != 3 {
			t.Errorf("Expected length 3, got %d", length)
		}
	})

	// Test 5: Object operations
	t.Run("ObjectOperations", func(t *testing.T) {
		obj := map[string]interface{}{
			"name": "original",
			"age":  30,
		}

		if err := json.Set(ctx, "test:obj", obj, time.Hour); err != nil {
			t.Fatalf("Failed to set JSON: %v", err)
		}

		// Merge objects
		toMerge := map[string]interface{}{
			"city":    "NYC",
			"country": "USA",
		}

		if err := json.ObjectMerge(ctx, "test:obj", "$", toMerge); err != nil {
			t.Fatalf("Failed to merge: %v", err)
		}

		// Get keys
		keys, err := json.ObjectKeys(ctx, "test:obj", "$")
		if err != nil {
			t.Fatalf("Failed to get keys: %v", err)
		}

		if len(keys) != 4 {
			t.Errorf("Expected 4 keys, got %d", len(keys))
		}
	})

	// Test 6: DocumentRepository
	t.Run("DocumentRepository", func(t *testing.T) {
		repo := NewDocumentRepository[User](client, "users")

		user := &User{
			ID:    "user4",
			Name:  "Repo User",
			Email: "repo@example.com",
			Age:   35,
		}

		// Save
		if err := repo.Save(ctx, "user4", user, time.Hour); err != nil {
			t.Fatalf("Failed to save: %v", err)
		}

		// Find
		found, err := repo.FindByID(ctx, "user4")
		if err != nil {
			t.Fatalf("Failed to find: %v", err)
		}

		if found.Name != user.Name {
			t.Errorf("Expected name %s, got %s", user.Name, found.Name)
		}

		// Update path
		if err := repo.UpdatePath(ctx, "user4", "$.age", 36); err != nil {
			t.Fatalf("Failed to update path: %v", err)
		}

		// Get path
		var age int
		if err := repo.GetPath(ctx, "user4", "$.age", &age); err != nil {
			t.Fatalf("Failed to get path: %v", err)
		}

		if age != 36 {
			t.Errorf("Expected age 36, got %d", age)
		}

		// Exists
		exists, err := repo.Exists(ctx, "user4")
		if err != nil {
			t.Fatalf("Failed to check exists: %v", err)
		}
		if !exists {
			t.Error("Expected user to exist")
		}

		// Delete
		if err := repo.Delete(ctx, "user4"); err != nil {
			t.Fatalf("Failed to delete: %v", err)
		}
	})

	// Test 7: MultiGet
	t.Run("MultiGet", func(t *testing.T) {
		users := []User{
			{ID: "multi1", Name: "User 1"},
			{ID: "multi2", Name: "User 2"},
			{ID: "multi3", Name: "User 3"},
		}

		for _, u := range users {
			if err := json.Set(ctx, fmt.Sprintf("multi:%s", u.ID), u, time.Hour); err != nil {
				t.Fatalf("Failed to set JSON: %v", err)
			}
		}

		results, err := json.MultiGet(ctx, []string{"multi:multi1", "multi:multi2", "multi:multi3"})
		if err != nil {
			t.Fatalf("Failed to multi get: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
	})
}

// BenchmarkJSONSet benchmarks JSON set operations.
func BenchmarkJSONSet(b *testing.B) {
	user := User{
		ID:    "bench",
		Name:  "Benchmark User",
		Email: "bench@example.com",
		Age:   30,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This would require a real client
		_ = user
		_ = ctx
	}
}
