//go:build !no_fiber

package fiber

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/DaiYuANg/arcgo/httpx/adapter"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// Adapter documents related behavior.
//
// Note.
// Note.
// Note.
// Note.
type Adapter struct {
	app     *fiber.App
	group   fiber.Router
	logger  *slog.Logger
	huma    huma.API
	humaCfg adapter.HumaOptions
}

// New creates related functionality.
func New(app ...*fiber.App) *Adapter {
	var a *fiber.App
	if len(app) > 0 {
		a = app[0]
	} else {
		a = fiber.New()
	}

	return &Adapter{
		app:    a,
		group:  a,
		logger: slog.Default(),
	}
}

// ConfigureHuma configures related behavior.
func (a *Adapter) ConfigureHuma(opts adapter.HumaOptions) {
	a.humaCfg = opts
	cfg := huma.DefaultConfig(opts.Title, opts.Version)
	adapter.ApplyHumaConfig(&cfg, opts)
	a.huma = humafiber.New(a.app, cfg)
}

// WithLogger configures related behavior.
func (a *Adapter) WithLogger(logger *slog.Logger) *Adapter {
	a.logger = logger
	return a
}

// Name returns related data.
func (a *Adapter) Name() string {
	return "fiber"
}

// Handle registers related handlers.
func (a *Adapter) Handle(method, path string, handler adapter.HandlerFunc) {
	a.group.Add(method, path, a.wrapHandler(handler))
}

// Group creates related functionality.
func (a *Adapter) Group(prefix string) adapter.Adapter {
	return &Adapter{
		app:     a.app,
		group:   a.group.Group(prefix),
		logger:  a.logger,
		huma:    a.huma,
		humaCfg: a.humaCfg,
	}
}

// ServeHTTP supports related behavior.
func (a *Adapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "fiber adapter does not support net/http ServeHTTP; use ListenAndServe", http.StatusNotImplemented)
}

// Router returns related data.
// Note.
// Note.
func (a *Adapter) Router() *fiber.App {
	return a.app
}

// Listen documents related behavior.
func (a *Adapter) Listen(addr string) error {
	return a.app.Listen(addr)
}

// Shutdown documents related behavior.
func (a *Adapter) Shutdown() error {
	return a.app.Shutdown()
}

// ListenContext starts related services.
func (a *Adapter) ListenContext(ctx context.Context, addr string) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Listen(addr)
	}()

	select {
	case err := <-errCh:
		if isExpectedFiberClose(err) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownErr := a.Shutdown()
		listenErr := <-errCh
		if shutdownErr != nil {
			return shutdownErr
		}
		if isExpectedFiberClose(listenErr) {
			return nil
		}
		return listenErr
	}
}

// wrapHandler wraps related logic.
func (a *Adapter) wrapHandler(handler adapter.HandlerFunc) fiber.Handler {
	return func(c *fiber.Ctx) error {
		w := &responseWriter{ctx: c}
		r := convertRequest(c)

		if err := handler(r.Context(), w, r); err != nil {
			a.logger.Error("Handler error",
				slog.String("method", c.Method()),
				slog.String("path", c.Path()),
				slog.String("error", err.Error()),
			)
			return err
		}
		return nil
	}
}

// convertRequest converts related values.
func convertRequest(c *fiber.Ctx) *http.Request {
	u := &url.URL{
		Path:     c.Path(),
		RawQuery: string(c.Request().URI().QueryString()),
	}

	header := make(http.Header)
	for k, v := range c.Request().Header.All() {
		header.Add(string(k), string(v))
	}

	req := &http.Request{
		Method:        c.Method(),
		URL:           u,
		Header:        header,
		Body:          io.NopCloser(bytes.NewReader(c.Body())),
		ContentLength: int64(len(c.Body())),
		Host:          string(c.Request().Header.Host()),
		RemoteAddr:    c.IP(),
	}

	return req.WithContext(adapter.WithRouteParams(userContext(c), c.AllParams()))
}

// responseWriter documents related behavior.
type responseWriter struct {
	ctx        *fiber.Ctx
	statusCode int
	header     http.Header
	applied    bool
}

func (w *responseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.ctx.Status(http.StatusOK)
	}
	w.applyHeaders()
	return w.ctx.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ctx.Status(statusCode)
	w.applyHeaders()
}

// HumaAPI returns related data.
func (a *Adapter) HumaAPI() huma.API {
	return a.huma
}

// HasHuma checks related state.
func (a *Adapter) HasHuma() bool {
	return a.huma != nil
}

func (w *responseWriter) applyHeaders() {
	if w.applied || w.header == nil {
		return
	}
	lo.ForEach(lo.Keys(w.header), func(key string, _ int) {
		values := w.header[key]
		w.ctx.Response().Header.Del(key)
		lo.ForEach(values, func(value string, _ int) {
			w.ctx.Response().Header.Add(key, value)
		})
	})
	w.applied = true
}

func userContext(c *fiber.Ctx) context.Context {
	ctx := c.UserContext()
	return mo.TupleToOption(ctx, ctx != nil).OrElse(context.Background())
}

func isExpectedFiberClose(err error) bool {
	if err == nil {
		return true
	}

	if errors.Is(err, http.ErrServerClosed) {
		return true
	}

	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "server is not running") ||
		strings.Contains(lower, "use of closed network connection")
}
