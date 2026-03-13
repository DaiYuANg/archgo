package httpx

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer_GenericGetWithDefaultHuma(t *testing.T) {
	server := newServer()

	err := Get(server, "/ping", func(ctx context.Context, input *struct{}) (*pingOutput, error) {
		out := &pingOutput{}
		out.Body.Message = "pong"
		return out, nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "pong")
}

func TestServer_GenericPostDecodeBody(t *testing.T) {
	server := newServer()

	err := Post(server, "/echo", func(ctx context.Context, input *echoInput) (*echoOutput, error) {
		out := &echoOutput{}
		out.Body.Name = input.Body.Name
		return out, nil
	})
	assert.NoError(t, err)

	body := []byte(`{"name":"arcgo"}`)
	req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "arcgo")
}

func TestServer_GenericPostInvalidJSON(t *testing.T) {
	server := newServer()

	err := Post(server, "/echo", func(ctx context.Context, input *echoInput) (*echoOutput, error) {
		out := &echoOutput{}
		out.Body.Name = input.Body.Name
		return out, nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewReader([]byte(`{"name":`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")
}

func TestServer_WithValidation_InvalidBody(t *testing.T) {
	server := newServer(WithValidation())

	err := Post(server, "/validated", func(ctx context.Context, input *validatedBodyInput) (*validatedBodyOutput, error) {
		out := &validatedBodyOutput{}
		out.Body.Name = input.Body.Name
		return out, nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/validated", bytes.NewReader([]byte(`{"name":"ab"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "request validation failed")
}

func TestServer_WithValidation_ValidBody(t *testing.T) {
	server := newServer(WithValidation())

	err := Post(server, "/validated", func(ctx context.Context, input *validatedBodyInput) (*validatedBodyOutput, error) {
		out := &validatedBodyOutput{}
		out.Body.Name = input.Body.Name
		return out, nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/validated", bytes.NewReader([]byte(`{"name":"arcgo"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"name\":\"arcgo\"")
}

func TestServer_CustomRequestBinder(t *testing.T) {
	server := newServer()

	err := Get(server, "/custom-bind", func(ctx context.Context, input *customBindInput) (*customBindOutput, error) {
		out := &customBindOutput{}
		out.Body.ID = input.ID
		out.Body.Token = input.Token
		return out, nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/custom-bind?user_id=123", nil)
	req.Header.Set("X-Token", "token-abc")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":123`)
	assert.Contains(t, w.Body.String(), `"token":"token-abc"`)
}

func TestServer_CustomRequestBinderError(t *testing.T) {
	server := newServer()

	err := Get(server, "/custom-bind", func(ctx context.Context, input *customBindInput) (*customBindOutput, error) {
		out := &customBindOutput{}
		out.Body.ID = input.ID
		return out, nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/custom-bind?user_id=not-an-int", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "user_id")
}

func TestServer_GroupWithBasePath(t *testing.T) {
	server := newServer(WithBasePath("/api"))
	v1 := server.Group("/v1")

	err := GroupGet(v1, "/health", func(ctx context.Context, input *struct{}) (*pingOutput, error) {
		out := &pingOutput{}
		out.Body.Message = "ok"
		return out, nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, server.HasRoute(http.MethodGet, "/api/v1/health"))
}
