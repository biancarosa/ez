package ez

import (
	"context"
	"fmt"
	"net/http"
)

type EZServer struct {
	s *http.Server
	r []Route
}

func New(s *http.Server) *EZServer {
	return &EZServer{
		s: s,
		r: make([]Route, 0),
	}
}

func (ez *EZServer) NotFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func matchesPattern(path, pattern string) bool {
	return path == pattern
}

// Handler returns an http.Handler that processes the route
func (ez *EZServer) Handler(route Route) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !contains(route.Method, r.Method) || !matchesPattern(r.URL.Path, route.Pattern) {
			ez.NotFound(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), RouteKey, route)
		route.Handler(w, r.WithContext(ctx))
	})
}

func (ez *EZServer) RegisterRoute(route Route) {
	ez.r = append(ez.r, route)
	http.Handle(route.Pattern, ez.Handler(route))
}

func (ez *EZServer) GetRoutes() []Route {
	return ez.r
}

func (ez *EZServer) ListenAndServe() error {
	fmt.Println("Running server on", ez.s.Addr)
	err := ez.s.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	fmt.Println("Shutdown")
	return nil
}

func (ez *EZServer) GenerateDocs() error {
	generator := DocsGenerator{
		server: ez,
	}
	return generator.GenerateDocs()
}
