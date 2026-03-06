//go:build !no_echo

package echo

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/DaiYuANg/arcgo/httpx/adapter"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
)

// Adapter documents related behavior.
//
// Note.
// Note.
// Note.
// Note.
type Adapter struct {
	engine  *echo.Echo
	group   *echo.Group
	logger  *slog.Logger
	huma    huma.API
	humaCfg adapter.HumaOptions
}

// New creates related functionality.
func New(engine ...*echo.Echo) *Adapter {
	var eng *echo.Echo
	if len(engine) > 0 {
		eng = engine[0]
	} else {
		eng = echo.New()
	}

	return &Adapter{
		engine: eng,
		group:  nil,
		logger: slog.Default(),
	}
}

// ConfigureHuma configures related behavior.
func (a *Adapter) ConfigureHuma(opts adapter.HumaOptions) {
	a.humaCfg = opts
	cfg := huma.DefaultConfig(opts.Title, opts.Version)
	adapter.ApplyHumaConfig(&cfg, opts)
	a.huma = humaecho.New(a.engine, cfg)
}

// WithLogger configures related behavior.
func (a *Adapter) WithLogger(logger *slog.Logger) *Adapter {
	a.logger = logger
	return a
}

// Name returns related data.
func (a *Adapter) Name() string {
	return "echo"
}

// Handle registers related handlers.
func (a *Adapter) Handle(method, path string, handler adapter.HandlerFunc) {
	if a.group != nil {
		a.group.Add(method, path, a.echoHandler(handler))
	} else {
		a.engine.Add(method, path, a.echoHandler(handler))
	}
}

// Group creates related functionality.
func (a *Adapter) Group(prefix string) adapter.Adapter {
	var g *echo.Group
	if a.group != nil {
		g = a.group.Group(prefix)
	} else {
		g = a.engine.Group(prefix)
	}

	return &Adapter{
		engine:  a.engine,
		group:   g,
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
func (a *Adapter) Router() *echo.Echo {
	return a.engine
}

// Listen starts related services.
func (a *Adapter) Listen(addr string) error {
	return a.engine.Start(addr)
}

// ListenContext starts related services.
func (a *Adapter) ListenContext(ctx context.Context, addr string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.engine.Start(addr)
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
		if err := a.engine.Shutdown(shutdownCtx); err != nil {
			return err
		}
		err := <-errCh
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

// echoHandler wraps related logic.
func (a *Adapter) echoHandler(handler adapter.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		params := make(map[string]string, len(c.ParamNames()))
		for _, name := range c.ParamNames() {
			params[name] = c.Param(name)
		}

		req := c.Request().WithContext(adapter.WithRouteParams(c.Request().Context(), params))
		if err := handler(req.Context(), c.Response(), req); err != nil {
			a.logger.Error("Handler error",
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.String("error", err.Error()),
			)
			return err
		}
		return nil
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
