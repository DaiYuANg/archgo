package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DaiYuANg/arcgo/httpx/adapter"
	adapterecho "github.com/DaiYuANg/arcgo/httpx/adapter/echo"
	adapterfiber "github.com/DaiYuANg/arcgo/httpx/adapter/fiber"
	adaptergin "github.com/DaiYuANg/arcgo/httpx/adapter/gin"
	"github.com/DaiYuANg/arcgo/httpx/adapter/std"
	"github.com/danielgtaylor/huma/v2"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type pingOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

type echoInput struct {
	Body struct {
		Name string `json:"name"`
	}
}

type echoOutput struct {
	Body struct {
		Name string `json:"name"`
	}
}

type customBindInput struct {
	ID    int    `query:"user_id"`
	Token string `header:"X-Token"`
}

type customBindOutput struct {
	Body struct {
		ID    int    `json:"id"`
		Token string `json:"token"`
	}
}

type paramsInput struct {
	ID    int    `query:"id"`
	Flag  bool   `query:"flag"`
	Trace string `header:"X-Trace-ID"`
}

type paramsOutput struct {
	Body struct {
		ID    int    `json:"id"`
		Flag  bool   `json:"flag"`
		Trace string `json:"trace"`
	}
}

type validatedBodyInput struct {
	Body struct {
		Name string `json:"name" validate:"required,min=3"`
	}
}

type validatedBodyOutput struct {
	Body struct {
		Name string `json:"name"`
	}
}

type validatedQueryInput struct {
	ID int `query:"id" validate:"required,min=1"`
}

type customValidatedInput struct {
	Body struct {
		Name string `json:"name" validate:"arc"`
	}
}

type humaPingOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

func TestServer_GenericGetWithDefaultHuma(t *testing.T) {
	server := NewServer()

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
	assert.True(t, server.HasHuma())
}

func TestServer_GenericPostDecodeBody(t *testing.T) {
	server := NewServer()

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
	server := NewServer()

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
	server := NewServer(WithValidation())

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
	server := NewServer(WithValidation())

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
	server := NewServer()

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
	server := NewServer()

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
	server := NewServer(WithBasePath("/api"))
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

func TestServer_StrongTypedQueryAndHeaderBinding(t *testing.T) {
	server := NewServer()

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

func TestServer_StrongTypedPathBindingOnStdAdapter(t *testing.T) {
	server := NewServer()

	type in struct {
		UserID int `path:"id"`
	}
	type out struct {
		Body struct {
			ID int `json:"id"`
		}
	}

	err := Get(server, "/users/{id}", func(ctx context.Context, input *in) (*out, error) {
		result := &out{}
		result.Body.ID = input.UserID
		return result, nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":123`)
}

func TestServer_StrongTypedPathBindingOnGinAdapter(t *testing.T) {
	server := NewServer(WithAdapter(adaptergin.New()))

	type in struct {
		UserID int `path:"id"`
	}
	type out struct {
		Body struct {
			ID int `json:"id"`
		}
	}

	err := Get(server, "/users/{id}", func(ctx context.Context, input *in) (*out, error) {
		result := &out{}
		result.Body.ID = input.UserID
		return result, nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/users/88", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":88`)
}

func TestServer_StrongTypedPathBindingOnEchoAdapter(t *testing.T) {
	server := NewServer(WithAdapter(adapterecho.New()))

	type in struct {
		UserID int `path:"id"`
	}
	type out struct {
		Body struct {
			ID int `json:"id"`
		}
	}

	err := Get(server, "/users/{id}", func(ctx context.Context, input *in) (*out, error) {
		result := &out{}
		result.Body.ID = input.UserID
		return result, nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/users/77", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":77`)
}

func TestServer_StrongTypedPathBindingOnFiberAdapter(t *testing.T) {
	server := NewServer(WithAdapter(adapterfiber.New()))

	type in struct {
		UserID int `path:"id"`
	}
	type out struct {
		Body struct {
			ID int `json:"id"`
		}
	}

	err := Get(server, "/users/{id}", func(ctx context.Context, input *in) (*out, error) {
		result := &out{}
		result.Body.ID = input.UserID
		return result, nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/users/66", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestServer_WithMiddleware(t *testing.T) {
	var middlewareCalled bool

	stdAdapter := std.New()
	stdAdapter.Router().Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, r)
		})
	})

	server := NewServer(WithAdapter(stdAdapter))
	err := Get(server, "/items", func(ctx context.Context, input *struct{}) (*pingOutput, error) {
		out := &pingOutput{}
		out.Body.Message = "ok"
		return out, nil
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.True(t, middlewareCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestServer_DefaultHumaEnabled(t *testing.T) {
	server := NewServer()

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
	assert.True(t, server.HasHuma())
	assert.NotNil(t, server.HumaAPI())
}

func TestServer_WithValidation_WorksWithHuma(t *testing.T) {
	server := NewServer(
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

	server := NewServer(WithValidator(customValidator))

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

type fakeHumaAdapter struct {
	lastOpts adapter.HumaOptions
	enabled  bool
}

func (f *fakeHumaAdapter) Name() string { return "fake" }

func (f *fakeHumaAdapter) Handle(method, path string, handler adapter.HandlerFunc) {}

func (f *fakeHumaAdapter) Group(prefix string) adapter.Adapter { return f }

func (f *fakeHumaAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

func (f *fakeHumaAdapter) ConfigureHuma(opts adapter.HumaOptions) {
	f.lastOpts = opts
	f.enabled = true
}

func (f *fakeHumaAdapter) HumaAPI() huma.API { return nil }

func (f *fakeHumaAdapter) HasHuma() bool { return f.enabled }

func TestServer_OpenAPIOptionsCallAdapter(t *testing.T) {
	fake := &fakeHumaAdapter{}

	_ = NewServer(
		WithAdapter(fake),
		WithOpenAPIInfo("Test API", "1.0.0", "desc"),
		WithOpenAPIDocs(false),
	)

	assert.True(t, fake.enabled)
	assert.Equal(t, "Test API", fake.lastOpts.Title)
	assert.Equal(t, "1.0.0", fake.lastOpts.Version)
	assert.Equal(t, "desc", fake.lastOpts.Description)
	assert.True(t, fake.lastOpts.DisableDocsRoutes)
}

func TestServer_GetRoutesAndFilters(t *testing.T) {
	server := NewServer()

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
