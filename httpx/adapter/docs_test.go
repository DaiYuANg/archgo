package adapter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestDocsController_DefaultPaths(t *testing.T) {
	router := chi.NewMux()
	cfg := huma.DefaultConfig("Test API", "1.0.0")
	cfg.DocsPath = ""
	cfg.OpenAPIPath = ""
	cfg.SchemasPath = ""
	api := humachi.New(router, cfg)

	// 创建 DocsController，使用空路径（期望使用默认值）
	docs := NewDocsController(api, HumaOptions{
		DocsPath:          "",
		OpenAPIPath:       "",
		SchemasPath:       "",
		DocsRenderer:      huma.DocsRendererScalar,
		DisableDocsRoutes: false,
	})

	t.Logf("initial docs controller config: %+v", docs.current)

	// 测试 /docs
	docsReq := httptest.NewRequest(http.MethodGet, "/docs", nil)
	docsRec := httptest.NewRecorder()
	result := docs.ServeHTTP(docsRec, docsReq)
	t.Logf("/docs status: %d, served: %v, body len: %d", docsRec.Code, result, docsRec.Body.Len())
	assert.True(t, result, "/docs should be served by DocsController")
	assert.Equal(t, http.StatusOK, docsRec.Code)

	// 测试 /openapi
	openAPIReq := httptest.NewRequest(http.MethodGet, "/openapi", nil)
	openAPIRec := httptest.NewRecorder()
	result = docs.ServeHTTP(openAPIRec, openAPIReq)
	t.Logf("/openapi status: %d, served: %v, body len: %d", openAPIRec.Code, result, openAPIRec.Body.Len())
	assert.True(t, result, "/openapi should be served by DocsController")
	assert.Equal(t, http.StatusOK, openAPIRec.Code)

	// 测试 /openapi.json
	openAPIJSONReq := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	openAPIJSONRec := httptest.NewRecorder()
	result = docs.ServeHTTP(openAPIJSONRec, openAPIJSONReq)
	t.Logf("/openapi.json status: %d, served: %v, body len: %d", openAPIJSONRec.Code, result, openAPIJSONRec.Body.Len())
	assert.True(t, result, "/openapi.json should be served by DocsController")
	assert.Equal(t, http.StatusOK, openAPIJSONRec.Code)
}

func TestDocsController_Configure(t *testing.T) {
	router := chi.NewMux()
	cfg := huma.DefaultConfig("Test API", "1.0.0")
	cfg.DocsPath = ""
	cfg.OpenAPIPath = ""
	cfg.SchemasPath = ""
	api := humachi.New(router, cfg)

	// 创建 DocsController，使用空路径
	docs := NewDocsController(api, HumaOptions{
		DocsPath:    "",
		OpenAPIPath: "",
		SchemasPath: "",
	})

	// 调用 Configure，传入空路径但 DisableDocsRoutes=false
	docs.Configure(HumaOptions{
		DocsPath:          "",
		OpenAPIPath:       "",
		SchemasPath:       "",
		DocsRenderer:      huma.DocsRendererScalar,
		DisableDocsRoutes: false,
	})

	// 测试 /openapi
	openAPIReq := httptest.NewRequest(http.MethodGet, "/openapi", nil)
	openAPIRec := httptest.NewRecorder()
	result := docs.ServeHTTP(openAPIRec, openAPIReq)
	assert.True(t, result, "/openapi should be served by DocsController")
	assert.Equal(t, http.StatusOK, openAPIRec.Code)
}
