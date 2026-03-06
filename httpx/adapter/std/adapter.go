package std

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/DaiYuANg/arcgo/httpx/adapter"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

// Adapter documents related behavior.
//
// Note.
// Note.
// Note.
// Note.
type Adapter struct {
	router  *chi.Mux
	prefix  string
	logger  *slog.Logger
	huma    huma.API
	humaCfg adapter.HumaOptions
}

// New creates related functionality.
func New() *Adapter {
	router := chi.NewMux()

	return &Adapter{
		router: router,
		logger: slog.Default(),
	}
}

// ConfigureHuma configures related behavior.
func (a *Adapter) ConfigureHuma(opts adapter.HumaOptions) {
	a.humaCfg = opts
	cfg := huma.DefaultConfig(opts.Title, opts.Version)
	adapter.ApplyHumaConfig(&cfg, opts)
	a.huma = humachi.New(a.router, cfg)
}

// WithLogger configures related behavior.
func (a *Adapter) WithLogger(logger *slog.Logger) *Adapter {
	a.logger = logger
	return a
}

// Name returns related data.
func (a *Adapter) Name() string {
	return "std"
}

// Handle registers related handlers.
func (a *Adapter) Handle(method, path string, handler adapter.HandlerFunc) {
	fullPath := joinPath(a.prefix, path)
	a.router.Method(method, fullPath, a.wrapHandler(handler))
}

// Group creates related functionality.
func (a *Adapter) Group(prefix string) adapter.Adapter {
	nextPrefix := a.prefix
	if prefix != "" && prefix != "/" {
		nextPrefix = joinPath(a.prefix, prefix)
	}
	return &Adapter{
		router:  a.router,
		prefix:  nextPrefix,
		logger:  a.logger,
		huma:    a.huma,
		humaCfg: a.humaCfg,
	}
}

// ServeHTTP documents related behavior.
func (a *Adapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

// Router returns related data.
// Note.
// Note.
func (a *Adapter) Router() *chi.Mux {
	return a.router
}

// Listen starts related services.
func (a *Adapter) Listen(addr string) error {
	return http.ListenAndServe(addr, a.router)
}

// ListenContext starts related services.
func (a *Adapter) ListenContext(ctx context.Context, addr string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	server := &http.Server{
		Addr:    addr,
		Handler: a.router,
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
func (a *Adapter) wrapHandler(handler adapter.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(r.Context(), w, r); err != nil {
			a.logger.Error("Handler error",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("error", err.Error()),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

func joinPath(prefix, path string) string {
	cleanPrefix := strings.TrimRight(prefix, "/")
	if cleanPrefix == "" {
		if path == "" {
			return "/"
		}
		if strings.HasPrefix(path, "/") {
			return path
		}
		return "/" + path
	}

	if path == "" || path == "/" {
		return cleanPrefix
	}
	if strings.HasPrefix(path, "/") {
		return cleanPrefix + path
	}
	return cleanPrefix + "/" + path
}
