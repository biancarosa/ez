package ez

import (
	"context"
	"fmt"
	"net/http"

	redoc "github.com/go-openapi/runtime/middleware"
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

func (ez *EZServer) RegisterRoute(route Route) {
	ez.r = append(ez.r, route)
	http.HandleFunc(route.Pattern, func(rw http.ResponseWriter, r *http.Request) {
		if contains(route.Method, r.Method) {
			ctx := context.WithValue(r.Context(), RouteKey, route)
			r = r.WithContext(ctx)

			route.Handler(rw, r)
		} else {
			ez.NotFound(rw, r)
		}
	})
}

func (ez *EZServer) GetRoutes() []Route {
	return ez.r
}

func RedocMiddleware(next http.Handler) http.Handler {
	opt := redoc.RedocOpts{
		SpecURL: "/openapi.json",
	}
	return redoc.Redoc(opt, next)
}
func (ez *EZServer) ListenAndServe() {
	fmt.Println("Running server on", ez.s.Addr)
	fmt.Println("Access docs on /docs")
	http.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./openapi.json")
	})

	// Redoc adds a middleware to serve /docs
	ez.s.Handler = RedocMiddleware(ez.s.Handler)
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
