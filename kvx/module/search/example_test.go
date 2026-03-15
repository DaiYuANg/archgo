package search

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/DaiYuANg/archgo/kvx"
	"github.com/DaiYuANg/archgo/kvx/adapter/redis"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Product represents a product document for search.
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Tags        []string `json:"tags"`
}

// Example demonstrates basic search operations.
func Example() {
	// This example shows how to use the search module
	// In real usage, you would connect to a Redis instance with RediSearch module

	fmt.Println("Search module example")
	// Output: Search module example
}

// ExampleQueryBuilder demonstrates building search queries.
func ExampleQueryBuilder() {
	// Build a query for products in the "electronics" category with price between 100 and 500
	qb := NewQueryBuilder()
	query := qb.
		Tag("category", "electronics").
		Range("price", 100, 500).
		Build()

	fmt.Println("Query:", query)
	// Output: Query: @category:{electronics} @price:[100 500]
}

// ExampleSchemaBuilder demonstrates building search index schemas.
func ExampleSchemaBuilder() {
	schema := NewSchemaBuilder().
		TextField("name", true).
		TextField("description", false).
		TagField("category", true).
		NumericField("price", true).
		Build()

	fmt.Printf("Schema has %d fields\n", len(schema))
	// Output: Schema has 4 fields
}

// SearchIntegrationTest contains integration tests using testcontainers.
// Run with: go test -tags=integration ./kvx/module/search/...
type SearchIntegrationTest struct {
	container testcontainers.Container
	client    kvx.Client
	search    *Search
	index     *Index
}

// setupRedisContainer starts a Redis container with RediSearch module.
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

// TestSearchIntegration tests search functionality with real Redis.
// This test requires Docker and is skipped by default.
func TestSearchIntegration(t *testing.T) {
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

	// Create search instance
	search := NewSearch(client)

	// Create index schema
	schema := NewSchemaBuilder().
		TextField("name", true).
		TextField("description", false).
		TagField("category", true).
		NumericField("price", true).
		Build()

	// Create index
	index := NewIndex(client, "products", "product:", schema)

	// Drop index if exists (ignore error)
	_ = index.Drop(ctx)

	// Create new index
	if err := index.Create(ctx); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// Add some test products using Hash operations
	products := []Product{
		{ID: "1", Name: "iPhone 15", Description: "Latest iPhone", Category: "electronics", Price: 999},
		{ID: "2", Name: "MacBook Pro", Description: "Professional laptop", Category: "electronics", Price: 1999},
		{ID: "3", Name: "Running Shoes", Description: "Comfortable running shoes", Category: "sports", Price: 120},
		{ID: "4", Name: "Coffee Maker", Description: "Automatic coffee maker", Category: "home", Price: 150},
	}

	for _, p := range products {
		key := fmt.Sprintf("product:%s", p.ID)
		values := map[string][]byte{
			"name":        []byte(p.Name),
			"description": []byte(p.Description),
			"category":    []byte(p.Category),
			"price":       []byte(fmt.Sprintf("%v", p.Price)),
		}
		if err := client.HSet(ctx, key, values); err != nil {
			t.Fatalf("Failed to add product: %v", err)
		}
	}

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// Test 1: Search all electronics
	t.Run("SearchByCategory", func(t *testing.T) {
		qb := NewQueryBuilder()
		query := qb.Tag("category", "electronics").Build()

		result, err := index.Search(ctx, query, &SearchOptions{Limit: 10})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if result.Total < 2 {
			t.Errorf("Expected at least 2 electronics, got %d", result.Total)
		}
	})

	// Test 2: Search by price range
	t.Run("SearchByPriceRange", func(t *testing.T) {
		qb := NewQueryBuilder()
		query := qb.Range("price", 100, 200).Build()

		result, err := index.Search(ctx, query, &SearchOptions{Limit: 10})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if result.Total < 2 {
			t.Errorf("Expected at least 2 products in price range, got %d", result.Total)
		}
	})

	// Test 3: Search with text
	t.Run("SearchByText", func(t *testing.T) {
		qb := NewQueryBuilder()
		query := qb.Text("name", "iPhone").Build()

		result, err := index.Search(ctx, query, &SearchOptions{Limit: 10})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if result.Total < 1 {
			t.Errorf("Expected at least 1 product matching 'iPhone', got %d", result.Total)
		}
	})

	// Test 4: Combined search
	t.Run("SearchCombined", func(t *testing.T) {
		qb := NewQueryBuilder()
		query := qb.
			Tag("category", "electronics").
			Range("price", 500, 1500).
			Build()

		result, err := index.Search(ctx, query, &SearchOptions{Limit: 10})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		// Should find iPhone 15 (999) but not MacBook Pro (1999)
		if result.Total != 1 {
			t.Errorf("Expected 1 product, got %d", result.Total)
		}
	})

	// Test 5: Search with sort
	t.Run("SearchWithSort", func(t *testing.T) {
		result, err := index.Search(ctx, "*", &SearchOptions{
			Limit:     10,
			SortBy:    "price",
			Ascending: false,
		})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if result.Total < 2 {
			t.Errorf("Expected at least 2 products, got %d", result.Total)
		}
	})

	// Cleanup
	if err := index.Drop(ctx); err != nil {
		log.Printf("Failed to drop index: %v", err)
	}
}

// ExampleSearchableRepository demonstrates using SearchableRepository.
func ExampleSearchableRepository() {
	// This example shows how to use SearchableRepository
	// In real usage, you would connect to a Redis instance

	fmt.Println("SearchableRepository example")
	// Output: SearchableRepository example
}

// BenchmarkQueryBuilder benchmarks query building.
func BenchmarkQueryBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		qb := NewQueryBuilder()
		_ = qb.
			Tag("category", "electronics").
			Range("price", 100, 500).
			Text("name", "phone").
			Build()
	}
}
