package ez

import "net/http"

type Route struct {
	Handler  func(http.ResponseWriter, *http.Request)
	Pattern  string
	Method   []string // http.Method
	Request  interface{}
	Response interface{}
}

type routeKeyType struct{}

var RouteKey = routeKeyType{}
