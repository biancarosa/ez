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

func (ez *EZServer) NotFound(rw http.ResponseWriter, r *http.Request) {
	http.NotFound(rw, r)
}

func matchesPattern(path, pattern string) bool {
	return path == pattern
}

func (ez *EZServer) HandleFunc(route Route) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		if contains(route.Method, r.Method) && matchesPattern(r.URL.Path, route.Pattern) {
			ctx := context.WithValue(r.Context(), RouteKey, route)
			r = r.WithContext(ctx)

			route.Handler(rw, r)
		} else {
			ez.NotFound(rw, r)
		}
	}
}

func (ez *EZServer) RegisterRoute(route Route) {
	ez.r = append(ez.r, route)
	http.HandleFunc(route.Pattern, ez.HandleFunc(route))
}

func (ez *EZServer) GetRoutes() []Route {
	return ez.r
}

func (ez *EZServer) ListenAndServe() {
	fmt.Println("Running server on", ez.s.Addr)
	err := ez.s.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Shutdown") // Todo register shutdown method
}

func (ez *EZServer) GenerateDocs() {
	generator := DocsGenerator{
		server: ez,
	}
	generator.GenerateDocs()
}
