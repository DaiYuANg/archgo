package main

import (
	"context"
	"fmt"
	"log"

	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/httpx/adapter/std"
	"github.com/DaiYuANg/arcgo/pkg/randomport"
	"github.com/danielgtaylor/huma/v2"
)

type healthOutput struct {
	Body struct {
		Status string `json:"status"`
	} `json:"body"`
}

type tenantInput struct {
	ID string `path:"id"`
}

type tenantOutput struct {
	Body struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"body"`
}

func main() {
	server := httpx.NewServer(
		httpx.WithAdapter(std.New()),
		httpx.WithBasePath("/api"),
		httpx.WithOpenAPIInfo("httpx organization example", "1.0.0", "Docs, security, and group defaults"),
		httpx.WithDocs(httpx.DocsOptions{
			Enabled:     true,
			DocsPath:    "/reference",
			OpenAPIPath: "/spec",
			SchemasPath: "/schemas",
			Renderer:    httpx.DocsRendererScalar,
		}),
		httpx.WithSecurity(httpx.SecurityOptions{
			Schemes: map[string]*huma.SecurityScheme{
				"bearerAuth": {
					Type:   "http",
					Scheme: "bearer",
				},
			},
			Requirements: []map[string][]string{
				{"bearerAuth": {}},
			},
		}),
	)

	server.RegisterGlobalHeader(&huma.Param{
		Name:        "X-Request-Id",
		In:          "header",
		Description: "request correlation id",
		Schema:      &huma.Schema{Type: "string"},
	})

	httpx.MustGet(server, "/health", func(ctx context.Context, input *struct{}) (*healthOutput, error) {
		out := &healthOutput{}
		out.Body.Status = "ok"
		return out, nil
	}, huma.OperationTags("system"))

	admin := server.Group("/admin")
	admin.RegisterTags(
		&huma.Tag{Name: "admin", Description: "Administrative APIs"},
		&huma.Tag{Name: "tenants", Description: "Tenant management"},
	)
	admin.DefaultTags("admin", "tenants")
	admin.DefaultSecurity(map[string][]string{"bearerAuth": {}})
	admin.DefaultParameters(&huma.Param{
		Name:        "X-Tenant",
		In:          "header",
		Description: "tenant scope",
		Schema:      &huma.Schema{Type: "string"},
	})
	admin.DefaultSummaryPrefix("Admin")
	admin.DefaultDescription("Administrative operations with shared docs metadata")
	admin.DefaultExternalDocs(&huma.ExternalDocs{
		Description: "Admin handbook",
		URL:         "https://example.com/admin-handbook",
	})
	admin.DefaultExtensions(map[string]any{
		"x-owner": "platform",
	})

	httpx.MustGroupGet(admin, "/tenants/{id}", func(ctx context.Context, input *tenantInput) (*tenantOutput, error) {
		out := &tenantOutput{}
		out.Body.ID = input.ID
		out.Body.Name = "tenant-" + input.ID
		return out, nil
	}, func(op *huma.Operation) {
		op.Summary = "Get tenant"
	})

	port := randomport.MustFind()
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Server starting on %s\n", addr)
	fmt.Printf("Docs:    http://localhost%s/reference\n", addr)
	fmt.Printf("OpenAPI: http://localhost%s/spec.json\n", addr)

	if err := server.ListenAndServe(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
