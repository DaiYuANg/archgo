// Package kvx_test provides examples for using the kvx package.
package kvx_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/DaiYuANg/archgo/kvx"
	"github.com/DaiYuANg/archgo/kvx/adapter/redis"
	"github.com/DaiYuANg/archgo/kvx/module/lock"
	"github.com/DaiYuANg/archgo/kvx/module/search"
	"github.com/DaiYuANg/archgo/kvx/repository"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// User represents a user entity.
type User struct {
	ID        string   `kvx:"id"`
	Name      string   `kvx:"name,index"`
	Email     string   `kvx:"email,index"`
	Age       int      `kvx:"age,index"`
	Tags      []string `kvx:"tags"`
	CreatedAt int64    `kvx:"created_at"`
}

// Product represents a product entity.
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
}

// setupRedisContainer starts a Redis container with all modules.
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

// Example demonstrates basic kvx usage.
func Example() {
	fmt.Println("kvx example")
	// Output: kvx example
}

// TestFullWorkflow demonstrates a complete workflow using all kvx features.
func TestFullWorkflow(t *testing.T) {
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

	// ====================
	// 1. Hash Repository
	// ====================
	t.Run("HashRepository", func(t *testing.T) {
		repo := repository.NewHashRepository[User](client, client, "user")

		// Create users
		users := []*User{
			{ID: "user1", Name: "Alice", Email: "alice@example.com", Age: 30, Tags: []string{"admin"}, CreatedAt: time.Now().Unix()},
			{ID: "user2", Name: "Bob", Email: "bob@example.com", Age: 25, Tags: []string{"user"}, CreatedAt: time.Now().Unix()},
			{ID: "user3", Name: "Charlie", Email: "charlie@example.com", Age: 35, Tags: []string{"user", "premium"}, CreatedAt: time.Now().Unix()},
		}

		// Save users
		for _, user := range users {
			if err := repo.Save(ctx, user); err != nil {
				t.Fatalf("Failed to save user: %v", err)
			}
		}

		// Find by ID
		found, err := repo.FindByID(ctx, "user1")
		if err != nil {
			t.Fatalf("Failed to find user: %v", err)
		}
		if found.Name != "Alice" {
			t.Errorf("Expected name Alice, got %s", found.Name)
		}

		// Find by field (using index)
		byEmail, err := repo.FindByField(ctx, "Email", "bob@example.com")
		if err != nil {
			t.Fatalf("Failed to find by email: %v", err)
		}
		if len(byEmail) != 1 || byEmail[0].Name != "Bob" {
			t.Errorf("Expected to find Bob by email")
		}

		// Find by multiple fields
		byFields, err := repo.FindByFields(ctx, map[string]string{
			"Age": "30",
		})
		if err != nil {
			t.Fatalf("Failed to find by fields: %v", err)
		}
		if len(byFields) != 1 || byFields[0].Name != "Alice" {
			t.Errorf("Expected to find Alice by age")
		}

		// Count
		count, err := repo.Count(ctx)
		if err != nil {
			t.Fatalf("Failed to count: %v", err)
		}
		if count != 3 {
			t.Errorf("Expected count 3, got %d", count)
		}

		// Delete
		if err := repo.Delete(ctx, "user3"); err != nil {
			t.Fatalf("Failed to delete: %v", err)
		}

		// Verify deletion
		exists, _ := repo.Exists(ctx, "user3")
		if exists {
			t.Error("Expected user3 to be deleted")
		}
	})

	// ====================
	// 2. Search Module
	// ====================
	t.Run("SearchModule", func(t *testing.T) {
		// Create search index
		schema := search.NewSchemaBuilder().
			TextField("name", true).
			TagField("category", true).
			NumericField("price", true).
			Build()

		index := search.NewIndex(client, "products", "product:", schema)

		// Drop if exists
		_ = index.Drop(ctx)

		// Create index
		if err := index.Create(ctx); err != nil {
			t.Fatalf("Failed to create index: %v", err)
		}
		defer index.Drop(ctx)

		// Add products
		products := []map[string][]byte{
			{"name": []byte("iPhone"), "category": []byte("electronics"), "price": []byte("999")},
			{"name": []byte("MacBook"), "category": []byte("electronics"), "price": []byte("1999")},
			{"name": []byte("Shoes"), "category": []byte("sports"), "price": []byte("120")},
		}

		for i, p := range products {
			key := fmt.Sprintf("product:%d", i+1)
			if err := client.HSet(ctx, key, p); err != nil {
				t.Fatalf("Failed to add product: %v", err)
			}
		}

		// Wait for indexing
		time.Sleep(1 * time.Second)

		// Search
		qb := search.NewQueryBuilder()
		query := qb.Tag("category", "electronics").Build()

		result, err := index.Search(ctx, query, &search.SearchOptions{Limit: 10})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if result.Total < 2 {
			t.Errorf("Expected at least 2 electronics, got %d", result.Total)
		}
	})

	// ====================
	// 3. Lock Module
	// ====================
	t.Run("LockModule", func(t *testing.T) {
		// Acquire lock
		lock := lock.New(client, "workflow:lock", lock.DefaultOptions())

		if err := lock.Acquire(ctx); err != nil {
			t.Fatalf("Failed to acquire lock: %v", err)
		}

		// Do some work
		time.Sleep(100 * time.Millisecond)

		// Release lock
		if err := lock.Release(ctx); err != nil {
			t.Fatalf("Failed to release lock: %v", err)
		}

		// Try to acquire released lock
		lock2 := lock.New(client, "workflow:lock", lock.DefaultOptions())
		if err := lock2.Acquire(ctx); err != nil {
			t.Fatalf("Should be able to acquire released lock: %v", err)
		}
		lock2.Release(ctx)
	})

	// ====================
	// 4. Distributed Counter with Lock
	// ====================
	t.Run("DistributedCounter", func(t *testing.T) {
		counterKey := "workflow:counter"

		// Initialize counter
		client.Set(ctx, counterKey, []byte("0"), 0)

		// Increment counter 10 times with lock
		for i := 0; i < 10; i++ {
			err := lock.WithLock(ctx, client, "workflow:counter:lock", lock.DefaultOptions(), func() error {
				// Get current value
				val, err := client.Get(ctx, counterKey)
				if err != nil {
					return err
				}

				var count int
				fmt.Sscanf(string(val), "%d", &count)
				count++

				// Set new value
				return client.Set(ctx, counterKey, []byte(fmt.Sprintf("%d", count)), 0)
			})

			if err != nil {
				t.Fatalf("Failed to increment counter: %v", err)
			}
		}

		// Verify counter
		val, _ := client.Get(ctx, counterKey)
		var finalCount int
		fmt.Sscanf(string(val), "%d", &finalCount)

		if finalCount != 10 {
			t.Errorf("Expected counter to be 10, got %d", finalCount)
		}
	})

	// ====================
	// 5. Pipeline Batch Operations
	// ====================
	t.Run("Pipeline", func(t *testing.T) {
		pipe := client.Pipeline()
		defer pipe.Close()

		// Queue multiple operations
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("pipeline:key:%d", i)
			value := fmt.Sprintf("value:%d", i)
			pipe.Enqueue("SET", []byte(key), []byte(value))
		}

		// Execute
		results, err := pipe.Exec(ctx)
		if err != nil {
			t.Fatalf("Pipeline execution failed: %v", err)
		}

		if len(results) != 10 {
			t.Errorf("Expected 10 results, got %d", len(results))
		}

		// Verify one value
		val, err := client.Get(ctx, "pipeline:key:5")
		if err != nil {
			t.Fatalf("Failed to get value: %v", err)
		}
		if string(val) != "value:5" {
			t.Errorf("Expected 'value:5', got '%s'", string(val))
		}
	})
}

// ExampleRepository shows repository usage.
func ExampleRepository() {
	// This example demonstrates repository usage
	fmt.Println("Repository example")
	// Output: Repository example
}

// ExampleSearch shows search usage.
func ExampleSearch() {
	// This example demonstrates search usage
	fmt.Println("Search example")
	// Output: Search example
}

// ExampleLock shows lock usage.
func ExampleLock() {
	// This example demonstrates lock usage
	fmt.Println("Lock example")
	// Output: Lock example
}

// Helper function for formatting
func fmt.Sprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
