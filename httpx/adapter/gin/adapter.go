//go:build !no_gin

package gin

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/DaiYuANg/arcgo/httpx/adapter"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
)

// Adapter documents related behavior.
//
// Note.
// Note.
// Note.
// Note.
type Adapter struct {
	engine  *gin.Engine
	group   *gin.RouterGroup
	logger  *slog.Logger
	huma    huma.API
	humaCfg adapter.HumaOptions
}

// New creates related functionality.
func New(engine ...*gin.Engine) *Adapter {
	var eng *gin.Engine
	if len(engine) > 0 {
		eng = engine[0]
	} else {
		eng = gin.New()
	}

	return &Adapter{
		engine: eng,
		group:  &eng.RouterGroup,
		logger: slog.Default(),
	}
}

// ConfigureHuma configures related behavior.
func (a *Adapter) ConfigureHuma(opts adapter.HumaOptions) {
	a.humaCfg = opts
	cfg := huma.DefaultConfig(opts.Title, opts.Version)
	adapter.ApplyHumaConfig(&cfg, opts)
	a.huma = humagin.New(a.engine, cfg)
}

// WithLogger configures related behavior.
func (a *Adapter) WithLogger(logger *slog.Logger) *Adapter {
	a.logger = logger
	return a
}

// Name returns related data.
func (a *Adapter) Name() string {
	return "gin"
}

// Handle registers related handlers.
func (a *Adapter) Handle(method, path string, handler adapter.HandlerFunc) {
	a.group.Handle(method, path, a.wrapHandler(handler))
}

// Group creates related functionality.
func (a *Adapter) Group(prefix string) adapter.Adapter {
	return &Adapter{
		engine:  a.engine,
		group:   a.group.Group(prefix),
		logger:  a.logger,
		huma:    a.huma,
		humaCfg: a.humaCfg,
	}
}

// ServeHTTP documents related behavior.
func (a *Adapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.engine.ServeHTTP(w, r)
}

// Router returns related data.
// Note.
// Note.
func (a *Adapter) Router() *gin.Engine {
	return a.engine
}

// Listen starts related services.
func (a *Adapter) Listen(addr string) error {
	return a.engine.Run(addr)
}

// ListenContext starts related services.
func (a *Adapter) ListenContext(ctx context.Context, addr string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	server := &http.Server{
		Addr:    addr,
		Handler: a.engine,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		err := <-errCh
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

// wrapHandler wraps related logic.
func (a *Adapter) wrapHandler(handler adapter.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		params := make(map[string]string, len(c.Params))
		for _, p := range c.Params {
			params[p.Key] = p.Value
		}

		req := c.Request.WithContext(adapter.WithRouteParams(c.Request.Context(), params))

		if err := handler(req.Context(), c.Writer, req); err != nil {
			a.logger.Error("Handler error",
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.String("error", err.Error()),
			)
			_ = c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}

// HumaAPI returns related data.
func (a *Adapter) HumaAPI() huma.API {
	return a.huma
}

// HasHuma checks related state.
func (a *Adapter) HasHuma() bool {
	return a.huma != nil
}
