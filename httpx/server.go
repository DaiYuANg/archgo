package httpx

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	collectionlist "github.com/DaiYuANg/arcgo/collectionx/list"
	collectionset "github.com/DaiYuANg/arcgo/collectionx/set"
	"github.com/DaiYuANg/arcgo/httpx/adapter"
	"github.com/DaiYuANg/arcgo/httpx/adapter/std"
	"github.com/danielgtaylor/huma/v2"
	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
)

// Server is the central httpx runtime object used to register routes and expose
// Huma/OpenAPI capabilities.
type Server struct {
	adapter            adapter.Adapter
	basePath           string
	routes             *collectionlist.ConcurrentList[RouteInfo]
	routeKeys          *collectionset.ConcurrentSet[string]
	logger             *slog.Logger
	printRoutes        bool
	validator          *validator.Validate
	panicRecover       bool
	accessLog          bool
	humaOptions        adapter.HumaOptions
	openAPIPatches     []func(*huma.OpenAPI)
	humaMiddlewares    []func(huma.Context, func(huma.Context))
	operationModifiers []func(*huma.Operation)
}

// Group represents a route group backed by a Huma group when available.
type Group struct {
	server    *Server
	prefix    string
	humaGroup *huma.Group
}

// ServerOption mutates a server during construction.
type ServerOption func(*Server)

// NewServer constructs a server, creating a default std adapter when none is provided.
func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		logger:       slog.Default(),
		routes:       collectionlist.NewConcurrentList[RouteInfo](),
		routeKeys:    collectionset.NewConcurrentSet[string](),
		panicRecover: true,
	}

	lo.ForEach(opts, func(opt ServerOption, _ int) {
		opt(s)
	})

	if s.adapter == nil {
		s.adapter = std.New(s.humaOptions)
	}

	s.applyPendingHumaConfig()

	return s
}

// Group creates a prefixed route group under the server base path.
func (s *Server) Group(prefix string) *Group {
	normalizedPrefix := normalizeRoutePrefix(prefix)
	var humaGroup *huma.Group
	if api := s.HumaAPI(); api != nil {
		humaGroup = huma.NewGroup(api, joinRoutePath(s.basePath, normalizedPrefix))
	}
	return &Group{
		server:    s,
		prefix:    normalizedPrefix,
		humaGroup: humaGroup,
	}
}

// printRoutesIfEnabled logs registered routes when route printing is enabled.
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
	return s.routesSnapshot()
}

// GetRoutesByMethod returns routes matching the given HTTP method.
func (s *Server) GetRoutesByMethod(method string) []RouteInfo {
	method = strings.ToUpper(method)
	return lo.Filter(s.routesSnapshot(), func(route RouteInfo, _ int) bool {
		return route.Method == method
	})
}

// GetRoutesByPath returns routes whose path starts with the given prefix.
func (s *Server) GetRoutesByPath(prefix string) []RouteInfo {
	if prefix == "" {
		return s.routesSnapshot()
	}
	return lo.Filter(s.routesSnapshot(), func(route RouteInfo, _ int) bool {
		return strings.HasPrefix(route.Path, prefix)
	})
}

// HasRoute reports whether a route has been registered.
func (s *Server) HasRoute(method, path string) bool {
	return s.routeKeys.Contains(routeKey(strings.ToUpper(method), path))
}

// RouteCount returns the number of unique registered routes.
func (s *Server) RouteCount() int {
	return s.routes.Len()
}

// Handler returns the server as an `http.Handler`.
func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.accessLog {
			s.adapter.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		recorder := newAccessLogResponseWriter(w)
		s.adapter.ServeHTTP(recorder, r)
		s.logRequest(r, recorder.Status(), time.Since(start))
	})
}

// ServeHTTP delegates request handling to the underlying adapter.
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
	)

	if listenable, ok := s.adapter.(adapter.ListenableAdapter); ok {
		if err := listenable.Listen(addr); err != nil {
			return fmt.Errorf("httpx: adapter %q listen on %q: %w", s.adapter.Name(), addr, err)
		}
		return nil
	}

	if err := http.ListenAndServe(addr, s.Handler()); err != nil {
		return fmt.Errorf("httpx: listen on %q: %w", addr, err)
	}
	return nil
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
		return fmt.Errorf("httpx: listen on %q: %w", addr, err)
	case <-ctx.Done():
		s.logger.Info("Shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("httpx: shutdown server on %q: %w", addr, err)
		}
		err := <-errCh
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("httpx: listen on %q: %w", addr, err)
	}
}

// Adapter returns the underlying runtime adapter.
func (s *Server) Adapter() adapter.Adapter {
	return s.adapter
}

// Logger returns the server logger.
func (s *Server) Logger() *slog.Logger {
	return s.logger
}

// PanicRecoverEnabled reports whether typed handlers are wrapped with panic recovery.
func (s *Server) PanicRecoverEnabled() bool {
	return s != nil && s.panicRecover
}

// AccessLogEnabled reports whether requests are logged through the server logger.
func (s *Server) AccessLogEnabled() bool {
	return s != nil && s.accessLog
}

// Validator returns the configured request validator, if any.
func (s *Server) Validator() *validator.Validate {
	return s.validator
}

// HumaAPI exposes the underlying Huma API.
func (s *Server) HumaAPI() huma.API {
	if s == nil || s.adapter == nil {
		return nil
	}
	return s.adapter.HumaAPI()
}

// OpenAPI returns the underlying Huma OpenAPI document.
func (s *Server) OpenAPI() *huma.OpenAPI {
	api := s.HumaAPI()
	if api == nil {
		return nil
	}
	return api.OpenAPI()
}

// Docs returns the server's tracked docs configuration.
func (s *Server) Docs() DocsOptions {
	if s == nil {
		return DefaultDocsOptions()
	}

	docs := DefaultDocsOptions()
	docs.Enabled = !s.humaOptions.DisableDocsRoutes
	if s.humaOptions.DocsPath != "" {
		docs.DocsPath = s.humaOptions.DocsPath
	}
	if s.humaOptions.OpenAPIPath != "" {
		docs.OpenAPIPath = s.humaOptions.OpenAPIPath
	}
	if s.humaOptions.SchemasPath != "" {
		docs.SchemasPath = s.humaOptions.SchemasPath
	}
	if s.humaOptions.DocsRenderer != "" {
		docs.Renderer = s.humaOptions.DocsRenderer
	}
	return docs
}

// ConfigureDocs mutates the tracked docs config.
// Route rebinding is adapter-dependent, so docs path changes are primarily intended
// for construction-time use via WithDocs.
func (s *Server) ConfigureDocs(fn func(*DocsOptions)) {
	if s == nil || fn == nil {
		return
	}

	docs := s.Docs()
	fn(&docs)
	applyDocsOptionsToHumaOptions(&s.humaOptions, docs)
	if configurable, ok := s.adapter.(adapter.HumaOptionsConfigurer); ok {
		configurable.ConfigureHumaOptions(s.humaOptions)
	}
}

// ConfigureOpenAPI mutates the underlying Huma OpenAPI document.
func (s *Server) ConfigureOpenAPI(fn func(*huma.OpenAPI)) {
	if fn == nil {
		return
	}
	openAPI := s.OpenAPI()
	if openAPI == nil {
		return
	}
	fn(openAPI)
}

// PatchOpenAPI mutates the underlying Huma OpenAPI document.
func (s *Server) PatchOpenAPI(fn func(*huma.OpenAPI)) {
	s.ConfigureOpenAPI(fn)
}

// UseHumaMiddleware registers API-level Huma middleware.
func (s *Server) UseHumaMiddleware(middlewares ...func(huma.Context, func(huma.Context))) {
	if len(middlewares) == 0 {
		return
	}
	s.humaMiddlewares = append(s.humaMiddlewares, middlewares...)
	if api := s.HumaAPI(); api != nil {
		api.UseMiddleware(middlewares...)
	}
}

// UseOperationModifier registers a server-level operation modifier for future operations.
func (s *Server) UseOperationModifier(modifier func(*huma.Operation)) {
	if s == nil || modifier == nil {
		return
	}
	s.operationModifiers = append(s.operationModifiers, modifier)
}

// AddTag registers OpenAPI tag metadata.
func (s *Server) AddTag(tag *huma.Tag) {
	if s == nil || tag == nil || tag.Name == "" {
		return
	}
	s.ConfigureOpenAPI(func(doc *huma.OpenAPI) {
		if findTag(doc.Tags, tag.Name) >= 0 {
			doc.Tags[findTag(doc.Tags, tag.Name)] = cloneTag(tag)
			return
		}
		doc.Tags = append(doc.Tags, cloneTag(tag))
	})
}

// RegisterSecurityScheme registers an OpenAPI security scheme component.
func (s *Server) RegisterSecurityScheme(name string, scheme *huma.SecurityScheme) {
	if s == nil || name == "" || scheme == nil {
		return
	}
	s.ConfigureOpenAPI(func(doc *huma.OpenAPI) {
		components := ensureComponents(doc)
		if components.SecuritySchemes == nil {
			components.SecuritySchemes = map[string]*huma.SecurityScheme{}
		}
		components.SecuritySchemes[name] = cloneSecurityScheme(scheme)
	})
}

// SetDefaultSecurity configures top-level OpenAPI security requirements.
func (s *Server) SetDefaultSecurity(requirements ...map[string][]string) {
	if s == nil {
		return
	}
	s.ConfigureOpenAPI(func(doc *huma.OpenAPI) {
		doc.Security = cloneSecurityRequirements(requirements)
	})
}

// RegisterComponentParameter registers a reusable OpenAPI parameter component.
func (s *Server) RegisterComponentParameter(name string, param *huma.Param) {
	if s == nil || name == "" || param == nil {
		return
	}
	s.ConfigureOpenAPI(func(doc *huma.OpenAPI) {
		components := ensureComponents(doc)
		if components.Parameters == nil {
			components.Parameters = map[string]*huma.Param{}
		}
		components.Parameters[name] = cloneParam(param)
	})
}

// RegisterComponentHeader registers a reusable OpenAPI header component.
func (s *Server) RegisterComponentHeader(name string, header *huma.Header) {
	if s == nil || name == "" || header == nil {
		return
	}
	s.ConfigureOpenAPI(func(doc *huma.OpenAPI) {
		components := ensureComponents(doc)
		if components.Headers == nil {
			components.Headers = map[string]*huma.Header{}
		}
		components.Headers[name] = cloneParam(header)
	})
}

// RegisterGlobalParameter adds a parameter to all current and future operations.
func (s *Server) RegisterGlobalParameter(param *huma.Param) {
	if s == nil || param == nil || param.Name == "" || param.In == "" {
		return
	}

	cloned := cloneParam(param)
	s.UseOperationModifier(func(op *huma.Operation) {
		appendOperationParameter(op, cloned)
	})
	s.ConfigureOpenAPI(func(doc *huma.OpenAPI) {
		forEachOperation(doc, func(op *huma.Operation) {
			appendOperationParameter(op, cloned)
		})
	})
}

// RegisterGlobalHeader adds a request header parameter to all current and future operations.
func (s *Server) RegisterGlobalHeader(header *huma.Param) {
	if header == nil {
		return
	}
	cloned := cloneParam(header)
	cloned.In = "header"
	s.RegisterGlobalParameter(cloned)
}

// HumaGroup exposes the underlying Huma group when one is available.
func (g *Group) HumaGroup() *huma.Group {
	if g == nil {
		return nil
	}
	return g.humaGroup
}

// UseHumaMiddleware registers Huma middleware on the group.
func (g *Group) UseHumaMiddleware(middlewares ...func(huma.Context, func(huma.Context))) {
	if g == nil || g.humaGroup == nil || len(middlewares) == 0 {
		return
	}
	g.humaGroup.UseMiddleware(middlewares...)
}

// UseOperationModifier registers a Huma operation modifier on the group.
func (g *Group) UseOperationModifier(modifier func(*huma.Operation, func(*huma.Operation))) {
	if g == nil || g.humaGroup == nil || modifier == nil {
		return
	}
	g.humaGroup.UseModifier(modifier)
}

// UseSimpleOperationModifier registers a simple operation modifier on the group.
func (g *Group) UseSimpleOperationModifier(modifier func(*huma.Operation)) {
	if g == nil || g.humaGroup == nil || modifier == nil {
		return
	}
	g.humaGroup.UseSimpleModifier(modifier)
}

// UseResponseTransformer registers response transformers on the group.
func (g *Group) UseResponseTransformer(transformers ...huma.Transformer) {
	if g == nil || g.humaGroup == nil || len(transformers) == 0 {
		return
	}
	g.humaGroup.UseTransformer(transformers...)
}

// DefaultTags applies group-level default tags to future operations.
func (g *Group) DefaultTags(tags ...string) {
	if g == nil || g.humaGroup == nil || len(tags) == 0 {
		return
	}
	g.humaGroup.UseSimpleModifier(func(op *huma.Operation) {
		for _, tag := range tags {
			if tag != "" && !containsString(op.Tags, tag) {
				op.Tags = append(op.Tags, tag)
			}
		}
	})
}

// DefaultSecurity applies group-level default security to operations that do not override it.
func (g *Group) DefaultSecurity(requirements ...map[string][]string) {
	if g == nil || g.humaGroup == nil || len(requirements) == 0 {
		return
	}
	cloned := cloneSecurityRequirements(requirements)
	g.humaGroup.UseSimpleModifier(func(op *huma.Operation) {
		if len(op.Security) == 0 {
			op.Security = cloneSecurityRequirements(cloned)
		}
	})
}

// DefaultParameters applies group-level parameters to future operations.
func (g *Group) DefaultParameters(params ...*huma.Param) {
	if g == nil || g.humaGroup == nil || len(params) == 0 {
		return
	}
	cloned := make([]*huma.Param, 0, len(params))
	for _, param := range params {
		if param != nil {
			cloned = append(cloned, cloneParam(param))
		}
	}
	if len(cloned) == 0 {
		return
	}
	g.humaGroup.UseSimpleModifier(func(op *huma.Operation) {
		for _, param := range cloned {
			appendOperationParameter(op, param)
		}
	})
}

// DefaultSummaryPrefix prepends a group-level summary prefix to future operations.
func (g *Group) DefaultSummaryPrefix(prefix string) {
	if g == nil || g.humaGroup == nil || strings.TrimSpace(prefix) == "" {
		return
	}
	trimmed := strings.TrimSpace(prefix)
	g.humaGroup.UseSimpleModifier(func(op *huma.Operation) {
		if strings.TrimSpace(op.Summary) == "" {
			op.Summary = trimmed
			return
		}
		if strings.HasPrefix(op.Summary, trimmed) {
			return
		}
		op.Summary = trimmed + " " + op.Summary
	})
}

// DefaultDescription applies a group-level description when an operation does not define one.
func (g *Group) DefaultDescription(description string) {
	if g == nil || g.humaGroup == nil || strings.TrimSpace(description) == "" {
		return
	}
	trimmed := strings.TrimSpace(description)
	g.humaGroup.UseSimpleModifier(func(op *huma.Operation) {
		if strings.TrimSpace(op.Description) == "" {
			op.Description = trimmed
		}
	})
}

// RegisterTags adds OpenAPI tag metadata for this group context.
func (g *Group) RegisterTags(tags ...*huma.Tag) {
	if g == nil || g.server == nil || len(tags) == 0 {
		return
	}
	for _, tag := range tags {
		g.server.AddTag(tag)
	}
}

// DefaultExternalDocs applies group-level external docs to future operations
// when an operation does not define its own external docs.
func (g *Group) DefaultExternalDocs(docs *huma.ExternalDocs) {
	if g == nil || g.humaGroup == nil || docs == nil {
		return
	}
	cloned := cloneExternalDocs(docs)
	g.humaGroup.UseSimpleModifier(func(op *huma.Operation) {
		if op.ExternalDocs == nil {
			op.ExternalDocs = cloneExternalDocs(cloned)
		}
	})
}

// DefaultExtensions applies group-level OpenAPI extensions to future operations.
func (g *Group) DefaultExtensions(extensions map[string]any) {
	if g == nil || g.humaGroup == nil || len(extensions) == 0 {
		return
	}
	cloned := cloneExtensions(extensions)
	g.humaGroup.UseSimpleModifier(func(op *huma.Operation) {
		if op.Extensions == nil {
			op.Extensions = map[string]any{}
		}
		for key, value := range cloned {
			if _, exists := op.Extensions[key]; !exists {
				op.Extensions[key] = value
			}
		}
	})
}

func (s *Server) addRoute(route RouteInfo) {
	key := routeKey(route.Method, route.Path)
	if s.routeKeys.Contains(key) {
		return
	}

	s.routeKeys.Add(key)
	s.routes.Add(route)
	s.printRoutesIfEnabled()
}

func (s *Server) routesSnapshot() []RouteInfo {
	//s.routesMu.RLock()
	//defer s.routesMu.RUnlock()
	return s.routes.Values()
}

func routeKey(method, path string) string {
	return method + " " + path
}

func (s *Server) logRequest(r *http.Request, status int, duration time.Duration) {
	if s == nil || s.logger == nil || r == nil {
		return
	}

	attrs := []any{
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.Int("status", status),
		slog.Duration("duration", duration),
	}

	if route, ok := s.matchRoute(r.Method, r.URL.Path); ok {
		attrs = append(attrs,
			slog.String("route", route.Path),
			slog.String("handler", route.HandlerName),
		)
	}

	s.logger.Info("httpx request", attrs...)
}

func (s *Server) matchRoute(method, path string) (RouteInfo, bool) {
	for _, route := range s.routesSnapshot() {
		if route.Method != method {
			continue
		}
		if route.Path == path || routePatternMatches(route.Path, path) {
			return route, true
		}
	}
	return RouteInfo{}, false
}

func routePatternMatches(pattern, path string) bool {
	pattern = strings.Trim(pattern, "/")
	path = strings.Trim(path, "/")

	if pattern == "" || path == "" {
		return pattern == path
	}

	patternSegments := strings.Split(pattern, "/")
	pathSegments := strings.Split(path, "/")
	if len(patternSegments) != len(pathSegments) {
		return false
	}

	for i, segment := range patternSegments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			continue
		}
		if segment != pathSegments[i] {
			return false
		}
	}
	return true
}

type accessLogResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func newAccessLogResponseWriter(w http.ResponseWriter) *accessLogResponseWriter {
	return &accessLogResponseWriter{ResponseWriter: w}
}

func (w *accessLogResponseWriter) WriteHeader(status int) {
	if !w.wroteHeader {
		w.status = status
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *accessLogResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func (w *accessLogResponseWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *accessLogResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *accessLogResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("httpx: response writer does not support hijacking")
	}
	return hijacker.Hijack()
}

func (w *accessLogResponseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

func (w *accessLogResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	readerFrom, ok := w.ResponseWriter.(io.ReaderFrom)
	if !ok {
		return io.Copy(w.ResponseWriter, r)
	}
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return readerFrom.ReadFrom(r)
}

func (s *Server) applyPendingHumaConfig() {
	if configurable, ok := s.adapter.(adapter.HumaOptionsConfigurer); ok {
		if !isZeroDocsOptions(s.humaOptions) {
			configurable.ConfigureHumaOptions(s.humaOptions)
		}
	}
	s.applyStoredOpenAPIPatches()
	if api := s.HumaAPI(); api != nil && len(s.humaMiddlewares) > 0 {
		api.UseMiddleware(s.humaMiddlewares...)
	}
}

func (s *Server) applyStoredOpenAPIPatches() {
	openAPI := s.OpenAPI()
	if openAPI == nil {
		return
	}
	for _, patch := range s.openAPIPatches {
		if patch != nil {
			patch(openAPI)
		}
	}
}

func applyDocsOptionsToHumaOptions(dst *adapter.HumaOptions, docs DocsOptions) {
	if dst == nil {
		return
	}
	dst.DisableDocsRoutes = !docs.Enabled
	if docs.DocsPath != "" {
		dst.DocsPath = docs.DocsPath
	}
	if docs.OpenAPIPath != "" {
		dst.OpenAPIPath = docs.OpenAPIPath
	}
	if docs.SchemasPath != "" {
		dst.SchemasPath = docs.SchemasPath
	}
	if docs.Renderer != "" {
		dst.DocsRenderer = docs.Renderer
	}
}

func ensureComponents(doc *huma.OpenAPI) *huma.Components {
	if doc.Components == nil {
		doc.Components = &huma.Components{}
	}
	return doc.Components
}

func forEachOperation(doc *huma.OpenAPI, fn func(*huma.Operation)) {
	if doc == nil || fn == nil {
		return
	}
	for _, item := range doc.Paths {
		if item == nil {
			continue
		}
		for _, op := range []*huma.Operation{item.Get, item.Put, item.Post, item.Delete, item.Options, item.Head, item.Patch, item.Trace} {
			if op != nil {
				fn(op)
			}
		}
	}
}

func appendOperationParameter(op *huma.Operation, param *huma.Param) {
	if op == nil || param == nil {
		return
	}
	for _, existing := range op.Parameters {
		if existing != nil && existing.Name == param.Name && existing.In == param.In {
			return
		}
	}
	op.Parameters = append(op.Parameters, cloneParam(param))
}

func cloneParam(param *huma.Param) *huma.Param {
	if param == nil {
		return nil
	}
	cloned := *param
	if param.Schema != nil {
		schema := *param.Schema
		cloned.Schema = &schema
	}
	if param.Examples != nil {
		cloned.Examples = make(map[string]*huma.Example, len(param.Examples))
		for k, v := range param.Examples {
			cloned.Examples[k] = v
		}
	}
	if param.Extensions != nil {
		cloned.Extensions = make(map[string]any, len(param.Extensions))
		for k, v := range param.Extensions {
			cloned.Extensions[k] = v
		}
	}
	return &cloned
}

func cloneTag(tag *huma.Tag) *huma.Tag {
	if tag == nil {
		return nil
	}
	cloned := *tag
	if tag.Extensions != nil {
		cloned.Extensions = make(map[string]any, len(tag.Extensions))
		for k, v := range tag.Extensions {
			cloned.Extensions[k] = v
		}
	}
	return &cloned
}

func cloneExternalDocs(docs *huma.ExternalDocs) *huma.ExternalDocs {
	if docs == nil {
		return nil
	}
	cloned := *docs
	cloned.Extensions = cloneExtensions(docs.Extensions)
	return &cloned
}

func cloneSecurityScheme(scheme *huma.SecurityScheme) *huma.SecurityScheme {
	if scheme == nil {
		return nil
	}
	cloned := *scheme
	if scheme.Extensions != nil {
		cloned.Extensions = make(map[string]any, len(scheme.Extensions))
		for k, v := range scheme.Extensions {
			cloned.Extensions[k] = v
		}
	}
	return &cloned
}

func cloneExtensions(values map[string]any) map[string]any {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]any, len(values))
	for k, v := range values {
		cloned[k] = v
	}
	return cloned
}

func cloneSecurityRequirements(requirements []map[string][]string) []map[string][]string {
	if len(requirements) == 0 {
		return nil
	}
	cloned := make([]map[string][]string, 0, len(requirements))
	for _, req := range requirements {
		if req == nil {
			cloned = append(cloned, nil)
			continue
		}
		item := make(map[string][]string, len(req))
		for k, scopes := range req {
			if scopes == nil {
				item[k] = []string{}
				continue
			}
			item[k] = append([]string(nil), scopes...)
		}
		cloned = append(cloned, item)
	}
	return cloned
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func findTag(tags []*huma.Tag, name string) int {
	for i, tag := range tags {
		if tag != nil && tag.Name == name {
			return i
		}
	}
	return -1
}

func isZeroDocsOptions(opts adapter.HumaOptions) bool {
	return opts.DocsPath == "" &&
		opts.OpenAPIPath == "" &&
		opts.SchemasPath == "" &&
		opts.DocsRenderer == "" &&
		!opts.DisableDocsRoutes
}
