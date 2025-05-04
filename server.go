package ez

import (
	"context"
	"fmt"
	"net/http"
)

type EZServer struct {
	s        *http.Server
	r        []Route
	capacity int
}

// New creates a new EZServer with an initial capacity for routes
func New(s *http.Server) *EZServer {
	const defaultRouteCapacity = 10
	return &EZServer{
		s:        s,
		r:        make([]Route, 0, defaultRouteCapacity),
		capacity: defaultRouteCapacity,
	}
}

// WithCapacity sets the initial capacity for routes
func (ez *EZServer) WithCapacity(capacity int) *EZServer {
	ez.r = make([]Route, 0, capacity)
	ez.capacity = capacity
	return ez
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

// RegisterRoute registers a new route. If the slice needs to grow, it will double in capacity.
func (ez *EZServer) RegisterRoute(route Route) {
	if len(ez.r) == cap(ez.r) {
		// Double capacity when we need to grow
		newCap := cap(ez.r) * 2
		if newCap == 0 {
			newCap = ez.capacity
		}
		newRoutes := make([]Route, len(ez.r), newCap)
		copy(newRoutes, ez.r)
		ez.r = newRoutes
	}
	ez.r = append(ez.r, route)
	http.Handle(route.Pattern, ez.Handler(route))
}

// RegisterRoutes registers multiple routes at once, preallocating the necessary capacity
func (ez *EZServer) RegisterRoutes(routes []Route) {
	if needed := len(ez.r) + len(routes); needed > cap(ez.r) {
		newRoutes := make([]Route, len(ez.r), needed)
		copy(newRoutes, ez.r)
		ez.r = newRoutes
	}
	for _, route := range routes {
		ez.RegisterRoute(route)
	}
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
