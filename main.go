package ez

import (
	"context"
	"fmt"
	"net/http"
)

type Route struct {
	Handler  func(http.ResponseWriter, *http.Request)
	Pattern  string
	Method   []string // http.Method
	Request  interface{}
	Response interface{}
}

type RouteKeyType string

const (
	RouteKey RouteKeyType = "request"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

type EZServer struct {
	s *http.Server
}

func New(s *http.Server) *EZServer {
	return &EZServer{
		s: s,
	}
}

func (ez *EZServer) NotFound(rw http.ResponseWriter, r *http.Request) {
	http.NotFound(rw, r)
}

func (ez *EZServer) RegisterRoutes(routes []Route) {
	for _, route := range routes {
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
}

func (ez *EZServer) ListenAndServe() {
	err := ez.s.ListenAndServe()

	fmt.Println("Running server on ", ez.s.Addr)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Shutdown") // Todo register shutdown method
}
