package main

import (
	"context"
	"log"

	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/httpx/adapter/std"
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
	adapter := std.New()

	server := httpx.NewServer(
		httpx.WithAdapter(adapter),
		httpx.WithBasePath("/api"),
		httpx.WithOpenAPIInfo("httpx quickstart", "1.0.0", "Typed HTTP quickstart example"),
		httpx.WithOpenAPIDocs(true),
		httpx.WithValidation(),
		httpx.WithPrintRoutes(true),
	)

	if err := httpx.Get(server, "/health", func(ctx context.Context, in *struct{}) (*healthOutput, error) {
		out := &healthOutput{}
		out.Body.Status = "ok"
		return out, nil
	}); err != nil {
		log.Fatal(err)
	}

	v1 := server.Group("/v1")

	if err := httpx.GroupPost(v1, "/users", func(ctx context.Context, in *createUserInput) (*createUserOutput, error) {
		out := &createUserOutput{}
		out.Body.ID = 1001
		out.Body.Name = in.Body.Name
		out.Body.Email = in.Body.Email
		return out, nil
	}); err != nil {
		log.Fatal(err)
	}

	if err := httpx.GroupGet(v1, "/users/{id}", func(ctx context.Context, in *getUserInput) (*getUserOutput, error) {
		out := &getUserOutput{}
		out.Body.ID = in.ID
		out.Body.Name = "demo-user"
		return out, nil
	}); err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.ListenAndServe(":8080"))
}
