package httpx

import (
	"context"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_MatchRoute_ExactRouteWinsOverParameterizedRoute(t *testing.T) {
	server := newServer()

	require.NoError(t, Get(server, "/users/{id}", func(ctx context.Context, input *struct{}) (*pingOutput, error) {
		out := &pingOutput{}
		out.Body.Message = "param"
		return out, nil
	}))
	require.NoError(t, Get(server, "/users/me", func(ctx context.Context, input *struct{}) (*pingOutput, error) {
		out := &pingOutput{}
		out.Body.Message = "exact"
		return out, nil
	}))

	matched, ok := server.matchRoute(http.MethodGet, "/users/me")
	require.True(t, ok)
	assert.Equal(t, "/users/me", matched.Path)

	matched, ok = server.matchRoute(http.MethodGet, "/users/42")
	require.True(t, ok)
	assert.Equal(t, "/users/{id}", matched.Path)
}

func TestServer_MatchRoute_OverlappingParameterizedRoutesKeepRegistrationOrder(t *testing.T) {
	server := newServer()

	require.NoError(t, Get(server, "/{kind}/list", func(ctx context.Context, input *struct{}) (*pingOutput, error) {
		out := &pingOutput{}
		out.Body.Message = "generic"
		return out, nil
	}))
	require.NoError(t, Get(server, "/users/{id}", func(ctx context.Context, input *struct{}) (*pingOutput, error) {
		out := &pingOutput{}
		out.Body.Message = "specific"
		return out, nil
	}))

	matched, ok := server.matchRoute(http.MethodGet, "/users/list")
	require.True(t, ok)
	assert.Equal(t, "/{kind}/list", matched.Path)
}

func TestServer_MatchRoute_OverlappingParameterizedRoutesKeepRegistrationOrderWhenReversed(t *testing.T) {
	server := newServer()

	require.NoError(t, Get(server, "/users/{id}", func(ctx context.Context, input *struct{}) (*pingOutput, error) {
		out := &pingOutput{}
		out.Body.Message = "specific"
		return out, nil
	}))
	require.NoError(t, Get(server, "/{kind}/list", func(ctx context.Context, input *struct{}) (*pingOutput, error) {
		out := &pingOutput{}
		out.Body.Message = "generic"
		return out, nil
	}))

	matched, ok := server.matchRoute(http.MethodGet, "/users/list")
	require.True(t, ok)
	assert.Equal(t, "/users/{id}", matched.Path)
}

func TestServer_AddTag_ReplacesExistingTagByName(t *testing.T) {
	server := newServer()

	server.AddTag(&huma.Tag{Name: "users", Description: "first"})
	server.AddTag(&huma.Tag{Name: "users", Description: "updated"})

	doc := server.OpenAPI()
	require.NotNil(t, doc)
	require.Len(t, doc.Tags, 1)
	assert.Equal(t, "updated", doc.Tags[0].Description)
}
