package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	collectionset "github.com/DaiYuANg/arcgo/collectionx/set"
	"github.com/DaiYuANg/arcgo/httpx/adapter"
	"github.com/DaiYuANg/arcgo/httpx/adapter/std"
	"github.com/danielgtaylor/huma/v2"
	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
)

// Server documents related behavior.
//
// Note.
// Note.
// Note.
// Note.
// Note.
type Server struct {
	adapter     adapter.Adapter
	basePath    string
	routesMu    sync.RWMutex
	routes      []RouteInfo
	routeKeys   *collectionset.Set[string]
	logger      *slog.Logger
	printRoutes bool
	humaOpts    adapter.HumaOptions
	validator   *validator.Validate
}

// Group documents related behavior.
type Group struct {
	server *Server
	prefix string
}

// ServerOption documents related behavior.
type ServerOption func(*Server)

// WithAdapter configures related behavior.
func WithAdapter(adapter adapter.Adapter) ServerOption {
	return func(s *Server) {
		s.adapter = adapter
	}
}

// WithAdapterName configures related behavior.
// Deprecated: use the new replacement API documented in this package.
func WithAdapterName(name string) ServerOption {
	return func(s *Server) {
		s.logger.Warn("WithAdapterName is deprecated, use adapter subpackages directly")
	}
}

// WithBasePath configures related behavior.
func WithBasePath(path string) ServerOption {
	return func(s *Server) {
		s.basePath = normalizeRoutePrefix(path)
	}
}

// WithLogger configures related behavior.
func WithLogger(logger *slog.Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithPrintRoutes configures related behavior.
func WithPrintRoutes(enabled bool) ServerOption {
	return func(s *Server) {
		s.printRoutes = enabled
	}
}

// WithOpenAPIInfo configures related behavior.
func WithOpenAPIInfo(title, version, description string) ServerOption {
	return func(s *Server) {
		if strings.TrimSpace(title) != "" {
			s.humaOpts.Title = strings.TrimSpace(title)
		}
		if strings.TrimSpace(version) != "" {
			s.humaOpts.Version = strings.TrimSpace(version)
		}
		if strings.TrimSpace(description) != "" {
			s.humaOpts.Description = strings.TrimSpace(description)
		}
	}
}

// WithOpenAPIDocs provides default behavior.
func WithOpenAPIDocs(enabled bool) ServerOption {
	return func(s *Server) {
		s.humaOpts.DisableDocsRoutes = !enabled
	}
}

// WithValidation enables related functionality.
func WithValidation() ServerOption {
	return func(s *Server) {
		if s.validator == nil {
			s.validator = validator.New(validator.WithRequiredStructEnabled())
		}
	}
}

// WithValidator closes related resources.
func WithValidator(v *validator.Validate) ServerOption {
	return func(s *Server) {
		s.validator = v
	}
}

// NewServer creates related functionality.
func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		logger:    slog.Default(),
		routes:    make([]RouteInfo, 0),
		routeKeys: collectionset.NewSet[string](),
		humaOpts:  adapter.DefaultHumaOptions(),
	}

	lo.ForEach(opts, func(opt ServerOption, _ int) {
		opt(s)
	})

	if s.adapter == nil {
		// Note.
		s.adapter = std.New()
	}

	if humaConfigurator, ok := s.adapter.(adapter.HumaConfigurator); ok {
		humaConfigurator.ConfigureHuma(s.humaOpts)
	}

	return s
}

// Group creates related functionality.
func (s *Server) Group(prefix string) *Group {
	return &Group{
		server: s,
		prefix: normalizeRoutePrefix(prefix),
	}
}

// printRoutesIfEnabled documents related behavior.
func (s *Server) printRoutesIfEnabled() {
	if !s.printRoutes {
		return
	}

	routes := s.routesSnapshot()
	s.logger.Info("Registered routes", slog.Int("count", len(routes)))
	lo.ForEach(routes, func(route RouteInfo, _ int) {
		s.logger.Info("  "+route.String(),
			slog.String("method", route.Method),
			slog.String("path", route.Path),
			slog.String("handler", route.HandlerName),
		)
	})
}

// GetRoutes returns related data.
func (s *Server) GetRoutes() []RouteInfo {
	routes := s.routesSnapshot()
	return lo.Map(routes, func(route RouteInfo, _ int) RouteInfo {
		return route
	})
}

// GetRoutesByMethod documents related behavior.
func (s *Server) GetRoutesByMethod(method string) []RouteInfo {
	routes := s.routesSnapshot()
	return lo.Filter(routes, func(route RouteInfo, _ int) bool {
		return route.Method == strings.ToUpper(method)
	})
}

// GetRoutesByPath documents related behavior.
func (s *Server) GetRoutesByPath(prefix string) []RouteInfo {
	routes := s.routesSnapshot()
	return lo.Filter(routes, func(route RouteInfo, _ int) bool {
		return len(prefix) == 0 || strings.HasPrefix(route.Path, prefix)
	})
}

// HasRoute checks related state.
func (s *Server) HasRoute(method, path string) bool {
	s.routesMu.RLock()
	defer s.routesMu.RUnlock()
	return s.routeKeys.Contains(routeKey(strings.ToUpper(method), path))
}

// RouteCount returns related data.
func (s *Server) RouteCount() int {
	s.routesMu.RLock()
	defer s.routesMu.RUnlock()
	return len(s.routes)
}

// Handler returns related data.
func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.adapter.ServeHTTP(w, r)
	})
}

// ServeHTTP documents related behavior.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Handler().ServeHTTP(w, r)
}

// ListenAndServe starts related services.
func (s *Server) ListenAndServe(addr string) error {
	routeCount := s.RouteCount()
	s.logger.Info("Starting server",
		slog.String("address", addr),
		slog.String("adapter", s.adapter.Name()),
		slog.Int("routes", routeCount),
		slog.Bool("huma_enabled", s.adapter.HasHuma()),
		slog.Bool("openapi_docs_enabled", !s.humaOpts.DisableDocsRoutes),
	)

	if listenable, ok := s.adapter.(adapter.ListenableAdapter); ok {
		return listenable.Listen(addr)
	}

	return http.ListenAndServe(addr, s.Handler())
}

// ListenAndServeContext starts related services.
func (s *Server) ListenAndServeContext(ctx context.Context, addr string) error {
	if listenable, ok := s.adapter.(adapter.ContextListenableAdapter); ok {
		return listenable.ListenContext(ctx, addr)
	}

	server := &http.Server{
		Addr:    addr,
		Handler: s.Handler(),
	}

	s.logger.Info("Starting server with context", slog.String("address", addr))

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
		s.logger.Info("Shutting down server")
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

// Adapter returns related data.
func (s *Server) Adapter() adapter.Adapter {
	return s.adapter
}

// Logger returns related data.
func (s *Server) Logger() *slog.Logger {
	return s.logger
}

// Validator returns related data.
func (s *Server) Validator() *validator.Validate {
	return s.validator
}

// HumaAPI returns related data.
func (s *Server) HumaAPI() huma.API {
	return s.adapter.HumaAPI()
}

// HasHuma checks related state.
func (s *Server) HasHuma() bool {
	return s.adapter.HasHuma()
}

func (s *Server) addRoute(route RouteInfo) {
	s.routesMu.Lock()
	defer s.routesMu.Unlock()
	key := routeKey(route.Method, route.Path)
	if s.routeKeys.Contains(key) {
		return
	}
	s.routeKeys.Add(key)
	s.routes = append(s.routes, route)
}

func (s *Server) routesSnapshot() []RouteInfo {
	s.routesMu.RLock()
	defer s.routesMu.RUnlock()
	return append([]RouteInfo(nil), s.routes...)
}

func routeKey(method, path string) string {
	return method + " " + path
}
