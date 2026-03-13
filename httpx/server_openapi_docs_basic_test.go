package httpx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
)

func TestServer_WithOpenAPIInfo_UpdatesDocument(t *testing.T) {
	server := newServer(WithOpenAPIInfo("Arc API", "2.0.0", "typed service"))

	openAPI := server.OpenAPI()
	if assert.NotNil(t, openAPI) && assert.NotNil(t, openAPI.Info) {
		assert.Equal(t, "Arc API", openAPI.Info.Title)
		assert.Equal(t, "2.0.0", openAPI.Info.Version)
		assert.Equal(t, "typed service", openAPI.Info.Description)
	}
}

func TestServer_WithOpenAPIDocs_DisablesDefaultDocsRoutes(t *testing.T) {
	server := newServer(WithOpenAPIDocs(false))

	docsReq := httptest.NewRequest(http.MethodGet, "/docs", nil)
	docsRec := httptest.NewRecorder()
	server.ServeHTTP(docsRec, docsReq)
	assert.Equal(t, http.StatusNotFound, docsRec.Code)

	openAPIReq := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	openAPIRec := httptest.NewRecorder()
	server.ServeHTTP(openAPIRec, openAPIReq)
	assert.Equal(t, http.StatusNotFound, openAPIRec.Code)
}

func TestServer_ConfigureOpenAPI_PatchesDocument(t *testing.T) {
	server := newServer()
	server.ConfigureOpenAPI(func(doc *huma.OpenAPI) {
		doc.Tags = append(doc.Tags, &huma.Tag{Name: "internal"})
	})

	openAPI := server.OpenAPI()
	if assert.NotNil(t, openAPI) {
		assert.Len(t, openAPI.Tags, 1)
		assert.Equal(t, "internal", openAPI.Tags[0].Name)
	}
}

func TestGroup_HumaMiddlewareAndModifier(t *testing.T) {
	server := newServer(WithBasePath("/api"))
	group := server.Group("/v1")
	group.UseHumaMiddleware(func(ctx huma.Context, next func(huma.Context)) {
		ctx.AppendHeader("X-Group", "v1")
		next(ctx)
	})
	group.UseSimpleOperationModifier(func(op *huma.Operation) {
		op.Tags = append(op.Tags, "group-tag")
	})

	err := GroupGet(group, "/items", func(ctx context.Context, input *struct{}) (*pingOutput, error) {
		out := &pingOutput{}
		out.Body.Message = "ok"
		return out, nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "v1", rec.Header().Get("X-Group"))

	pathItem := server.OpenAPI().Paths["/api/v1/items"]
	if assert.NotNil(t, pathItem) && assert.NotNil(t, pathItem.Get) {
		assert.Contains(t, pathItem.Get.Tags, "group-tag")
	}
}

func TestServer_WithDocs_CustomPaths(t *testing.T) {
	server := newServer(WithDocs(DocsOptions{
		Enabled:     true,
		DocsPath:    "/reference",
		OpenAPIPath: "/spec",
		SchemasPath: "/contracts",
		Renderer:    DocsRendererScalar,
	}))

	docsReq := httptest.NewRequest(http.MethodGet, "/reference", nil)
	docsRec := httptest.NewRecorder()
	server.ServeHTTP(docsRec, docsReq)
	assert.Equal(t, http.StatusOK, docsRec.Code)

	oldDocsReq := httptest.NewRequest(http.MethodGet, "/docs", nil)
	oldDocsRec := httptest.NewRecorder()
	server.ServeHTTP(oldDocsRec, oldDocsReq)
	assert.Equal(t, http.StatusNotFound, oldDocsRec.Code)

	specReq := httptest.NewRequest(http.MethodGet, "/spec.json", nil)
	specRec := httptest.NewRecorder()
	server.ServeHTTP(specRec, specReq)
	assert.Equal(t, http.StatusOK, specRec.Code)
	assert.Contains(t, specRec.Body.String(), "\"openapi\"")

	docs := server.Docs()
	assert.Equal(t, "/reference", docs.DocsPath)
	assert.Equal(t, "/spec", docs.OpenAPIPath)
	assert.Equal(t, "/contracts", docs.SchemasPath)
	assert.Equal(t, DocsRendererScalar, docs.Renderer)
}
