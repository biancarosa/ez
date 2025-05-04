package ez

import "net/http"

// Route represents a single HTTP route with generic request and response types
type Route[T any, U any] struct {
	Handler  func(http.ResponseWriter, *http.Request)
	Pattern  string
	Method   []string // http.Method
	Request  T
	Response U
}

type routeKeyType struct{}

var RouteKey = routeKeyType{}
