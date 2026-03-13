package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestServer_StrongTypedQueryAndHeaderBinding(t *testing.T) {
	server := newServer()

	err := Get(server, "/params", func(ctx context.Context, input *paramsInput) (*paramsOutput, error) {
		out := &paramsOutput{}
		out.Body.ID = input.ID
		out.Body.Flag = input.Flag
		out.Body.Trace = input.Trace
		return out, nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/params?id=42&flag=true", nil)
	req.Header.Set("X-Trace-ID", "trace-001")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":42`)
	assert.Contains(t, w.Body.String(), `"flag":true`)
	assert.Contains(t, w.Body.String(), `"trace":"trace-001"`)
}

func TestServer_WithMiddleware(t *testing.T) {
	// Note: Middleware must be added to the adapter before passing to httpx.Server.
	// Huma is now initialized at adapter creation time, so middleware should be
	// configured on the router/engine before calling adapter.New().
	server := newServer()
	err := Get(server, "/items", func(ctx context.Context, input *struct{}) (*pingOutput, error) {
		out := &pingOutput{}
		out.Body.Message = "ok"
		return out, nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestServer_DefaultHumaEnabled(t *testing.T) {
	server := newServer()

	err := Get(server, "/huma", func(ctx context.Context, input *struct{}) (*humaPingOutput, error) {
		out := &humaPingOutput{}
		out.Body.Message = "from huma"
		return out, nil
	}, huma.OperationTags("demo"))
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/huma", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "from huma")
	assert.NotNil(t, server.HumaAPI())
}

func TestServer_WithValidation_WorksWithHuma(t *testing.T) {
	server := newServer(
		WithValidation(),
	)

	err := Get(server, "/validate-huma", func(ctx context.Context, input *validatedQueryInput) (*humaPingOutput, error) {
		out := &humaPingOutput{}
		out.Body.Message = "ok"
		return out, nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/validate-huma", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "request validation failed")
}

func TestServer_WithCustomValidator(t *testing.T) {
	customValidator := validator.New()
	err := customValidator.RegisterValidation("arc", func(fl validator.FieldLevel) bool {
		return fl.Field().String() == "arc"
	})
	assert.NoError(t, err)

	server := newServer(WithValidator(customValidator))

	err = Post(server, "/custom-validate", func(ctx context.Context, input *customValidatedInput) (*validatedBodyOutput, error) {
		out := &validatedBodyOutput{}
		out.Body.Name = input.Body.Name
		return out, nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/custom-validate", bytes.NewReader([]byte(`{"name":"bad"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "request validation failed")
}

func TestServer_GetRoutesAndFilters(t *testing.T) {
	server := newServer()

	err := Get(server, "/users", func(ctx context.Context, input *struct{}) (*pingOutput, error) {
		out := &pingOutput{}
		out.Body.Message = "ok"
		return out, nil
	})
	assert.NoError(t, err)

	routes := server.GetRoutes()
	assert.Len(t, routes, 1)
	assert.Equal(t, http.MethodGet, routes[0].Method)

	getRoutes := server.GetRoutesByMethod(http.MethodGet)
	assert.Len(t, getRoutes, 1)

	pathRoutes := server.GetRoutesByPath("/users")
	assert.Len(t, pathRoutes, 1)

	assert.True(t, server.HasRoute(http.MethodGet, "/users"))

	var resp map[string]any
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
}
