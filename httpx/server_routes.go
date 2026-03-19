package httpx

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/samber/lo"
)

// GetRoutes returns related data.
func (s *Server) GetRoutes() []RouteInfo {
	return s.routesSnapshot()
}

// GetRoutesByMethod returns routes matching the given HTTP method.
func (s *Server) GetRoutesByMethod(method string) []RouteInfo {
	method = strings.ToUpper(method)
	if s == nil || method == "" {
		return nil
	}
	return s.routesByMethod.Get(method)
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
	if s == nil {
		return false
	}
	key := routeKey(strings.ToUpper(method), path)
	_, ok := s.routeExact.Get(key)
	return ok
}

// RouteCount returns the number of unique registered routes.
func (s *Server) RouteCount() int {
	return s.routes.Len()
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

func (s *Server) addRoute(route RouteInfo) {
	if s == nil {
		return
	}
	route.Method = strings.ToUpper(route.Method)
	key := routeKey(route.Method, route.Path)
	if _, loaded := s.routeExact.GetOrStore(key, route); loaded {
		return
	}

	s.routesByMethod.Put(route.Method, route)
	if isParameterizedRoute(route.Path) {
		s.parameterizedRouteMatcher(route.Method).Add(route.Path, route, s.routeSequence.Add(1))
	}

	s.routes.Add(route)
	s.printRoutesIfEnabled()
}

func (s *Server) routesSnapshot() []RouteInfo {
	return s.routes.Values()
}

func routeKey(method, path string) string {
	return method + " " + path
}

func (s *Server) matchRoute(method, path string) (RouteInfo, bool) {
	if s == nil {
		return RouteInfo{}, false
	}

	method = strings.ToUpper(method)
	key := routeKey(method, path)

	if route, ok := s.routeExact.Get(key); ok {
		return route, true
	}

	matcher, ok := s.routeMatchers.Get(method)
	if !ok || matcher == nil {
		return RouteInfo{}, false
	}
	return matcher.Match(path)
}

func isParameterizedRoute(path string) bool {
	return strings.Contains(path, "{") && strings.Contains(path, "}")
}

func (s *Server) parameterizedRouteMatcher(method string) *routeMatcher {
	if s == nil {
		return nil
	}

	matcher := newRouteMatcher()
	actual, _ := s.routeMatchers.GetOrStore(method, matcher)
	return actual
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
