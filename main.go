package ez

import "net/http"

type Route struct {
	Handler func(http.ResponseWriter, *http.Request)
	Pattern string
}

func RegisterRoutes(routes []Route) {
	for _, route := range routes {
		http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
			route.Handler(rw, r)
		})
	}
}
