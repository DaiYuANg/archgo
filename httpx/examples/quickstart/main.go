package main

import (
	"context"
	"fmt"
	"log"

	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/httpx/adapter"
	"github.com/DaiYuANg/arcgo/httpx/adapter/std"
	"github.com/DaiYuANg/arcgo/pkg/randomport"
)

type healthOutput struct {
	Body struct {
		Status string `json:"status"`
	} `json:"body"`
}

type createUserInput struct {
	Body struct {
		Name  string `json:"name" validate:"required,min=2,max=64"`
		Email string `json:"email" validate:"required,email"`
	} `json:"body"`
}

type createUserOutput struct {
	Body struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"body"`
}

type getUserInput struct {
	ID int `path:"id"`
}

type getUserOutput struct {
	Body struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"body"`
}

func main() {
	adapter := std.New(adapter.HumaOptions{
		Title:       "httpx quickstart",
		Version:     "1.0.0",
		Description: "Typed HTTP quickstart example",
		DocsPath:    "/docs",
		OpenAPIPath: "/openapi.json",
	})

	server := httpx.NewServer(
		httpx.WithAdapter(adapter),
		httpx.WithBasePath("/api"),
		httpx.WithValidation(),
		httpx.WithPrintRoutes(true),
	)

	httpx.MustGet(server, "/health", func(ctx context.Context, in *struct{}) (*healthOutput, error) {
		out := &healthOutput{}
		out.Body.Status = "ok"
		return out, nil
	})

	v1 := server.Group("/v1")

	httpx.MustGroupPost(v1, "/users", func(ctx context.Context, in *createUserInput) (*createUserOutput, error) {
		out := &createUserOutput{}
		out.Body.ID = 1001
		out.Body.Name = in.Body.Name
		out.Body.Email = in.Body.Email
		return out, nil
	})

	httpx.MustGroupGet(v1, "/users/{id}", func(ctx context.Context, in *getUserInput) (*getUserOutput, error) {
		out := &getUserOutput{}
		out.Body.ID = in.ID
		out.Body.Name = "demo-user"
		return out, nil
	})

	port := randomport.MustFind()
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Server starting on %s\n", addr)
	fmt.Printf("OpenAPI JSON: http://localhost%s/openapi.json\n", addr)
	fmt.Printf("Swagger UI:   http://localhost%s/docs\n", addr)

	if err := server.ListenAndServe(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
