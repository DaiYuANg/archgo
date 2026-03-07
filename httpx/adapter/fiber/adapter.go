//go:build !no_fiber

package fiber

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/DaiYuANg/arcgo/httpx/adapter"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// Adapter implements the fiber runtime bridge for httpx.
type Adapter struct {
	app    *fiber.App
	group  fiber.Router
	logger *slog.Logger
	huma   huma.API
	docs   *adapter.DocsController
	opts   AppOptions
}

// New constructs a fiber adapter backed by a fiber app and Huma API.
func New(app *fiber.App, opts ...adapter.HumaOptions) *Adapter {
	options := DefaultOptions()
	options.Huma = adapter.MergeHumaOptions(opts...)
	return NewWithOptions(app, options)
}

// AppOptions configures the fiber app created by the adapter when no app is supplied.
type AppOptions struct {
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// DefaultAppOptions returns the default fiber adapter app config.
func DefaultAppOptions() AppOptions {
	return AppOptions{
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}
}

// Options configures fiber adapter construction.
type Options struct {
	Huma   adapter.HumaOptions
	Logger *slog.Logger
	App    AppOptions
}

// DefaultOptions returns the default fiber adapter config.
func DefaultOptions() Options {
	return Options{
		Huma:   adapter.DefaultHumaOptions(),
		Logger: slog.Default(),
		App:    DefaultAppOptions(),
	}
}

// NewWithOptions constructs a fiber adapter from explicit construction-time options.
// App timeout settings only apply when the adapter creates the fiber app itself.
func NewWithOptions(app *fiber.App, opts Options) *Adapter {
	var a *fiber.App
	if app != nil {
		a = app
	} else {
		cfg := mergeAppOptions(opts.App)
		a = fiber.New(fiber.Config{
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		})
	}

	humaOpts := adapter.MergeHumaOptions(opts.Huma)
	cfg := huma.DefaultConfig(humaOpts.Title, humaOpts.Version)
	adapter.ApplyHumaConfig(&cfg, humaOpts)

	docsCfg := cfg
	docsCfg.DocsPath = ""
	docsCfg.OpenAPIPath = ""
	docsCfg.SchemasPath = ""
	api := humafiber.New(a, docsCfg)
	docs := adapter.NewDocsController(api, humaOpts)
	a.Use(func(c *fiber.Ctx) error {
		if docs.ServeHTTP(&responseWriter{ctx: c}, convertRequest(c)) {
			return nil
		}
		return c.Next()
	})

	return &Adapter{
		app:    a,
		group:  a,
		logger: defaultLogger(opts.Logger),
		huma:   api,
		docs:   docs,
		opts:   mergeAppOptions(opts.App),
	}
}

// WithLogger replaces the adapter logger.
func (a *Adapter) WithLogger(logger *slog.Logger) *Adapter {
	a.SetLogger(logger)
	return a
}

// SetLogger replaces the adapter logger.
func (a *Adapter) SetLogger(logger *slog.Logger) {
	if a == nil || logger == nil {
		return
	}
	a.logger = logger
}

// Name returns the adapter name.
func (a *Adapter) Name() string {
	return "fiber"
}

// Handle registers a native handler on the current fiber router.
func (a *Adapter) Handle(method, path string, handler adapter.HandlerFunc) {
	a.group.Add(method, path, a.wrapHandler(handler))
}

// Group returns a child adapter scoped to a fiber group.
func (a *Adapter) Group(prefix string) adapter.Adapter {
	return &Adapter{
		app:    a.app,
		group:  a.group.Group(prefix),
		logger: a.logger,
		huma:   a.huma,
		docs:   a.docs,
		opts:   a.opts,
	}
}

// ServeHTTP reports that the fiber adapter is not exposed as a net/http handler.
func (a *Adapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "fiber adapter does not support net/http ServeHTTP; use ListenAndServe", http.StatusNotImplemented)
}

// Router exposes the underlying fiber app.
func (a *Adapter) Router() *fiber.App {
	return a.app
}

// Listen starts the fiber server.
func (a *Adapter) Listen(addr string) error {
	if err := a.app.Listen(addr); err != nil {
		return fmt.Errorf("httpx/fiber: listen on %q: %w", addr, err)
	}
	return nil
}

// Shutdown stops the fiber server.
func (a *Adapter) Shutdown() error {
	if err := a.app.Shutdown(); err != nil {
		return fmt.Errorf("httpx/fiber: shutdown: %w", err)
	}
	return nil
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
		return fmt.Errorf("httpx/fiber: listen on %q: %w", addr, err)
	case <-ctx.Done():
		shutdownErr := a.shutdown()
		listenErr := <-errCh
		if shutdownErr != nil {
			return fmt.Errorf("httpx/fiber: shutdown on %q: %w", addr, shutdownErr)
		}
		if isExpectedFiberClose(listenErr) {
			return nil
		}
		return fmt.Errorf("httpx/fiber: listen on %q: %w", addr, listenErr)
	}
}

func (a *Adapter) shutdown() error {
	if a.opts.ShutdownTimeout > 0 {
		if err := a.app.ShutdownWithTimeout(a.opts.ShutdownTimeout); err != nil {
			return fmt.Errorf("httpx/fiber: shutdown: %w", err)
		}
		return nil
	}
	return a.Shutdown()
}

// wrapHandler adapts an httpx handler to a fiber handler.
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
			return fmt.Errorf("httpx/fiber: handler failed: %w", err)
		}
		return nil
	}
}

// convertRequest converts a fiber request into an `*http.Request`.
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

// responseWriter adapts a fiber response to the `http.ResponseWriter` shape.
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

// HumaAPI exposes the underlying Huma API.
func (a *Adapter) HumaAPI() huma.API {
	return a.huma
}

// ConfigureHumaOptions updates adapter-managed docs/openapi routing.
func (a *Adapter) ConfigureHumaOptions(opts adapter.HumaOptions) {
	if a == nil || a.docs == nil {
		return
	}
	a.docs.Configure(opts)
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

func defaultLogger(logger *slog.Logger) *slog.Logger {
	if logger != nil {
		return logger
	}
	return slog.Default()
}

func mergeAppOptions(opts AppOptions) AppOptions {
	defaults := DefaultAppOptions()
	if opts.ReadTimeout > 0 {
		defaults.ReadTimeout = opts.ReadTimeout
	}
	if opts.WriteTimeout > 0 {
		defaults.WriteTimeout = opts.WriteTimeout
	}
	if opts.IdleTimeout > 0 {
		defaults.IdleTimeout = opts.IdleTimeout
	}
	if opts.ShutdownTimeout > 0 {
		defaults.ShutdownTimeout = opts.ShutdownTimeout
	}
	return defaults
}
