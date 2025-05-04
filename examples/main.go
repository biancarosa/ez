package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/biancarosa/ez"
)

const (
	port = "5005"
)

func MainHandler(w http.ResponseWriter, req *http.Request) {
	route := req.Context().Value(ez.RouteKey).(ez.Route[any, HealthCheckResponse])
	res := route.Response
	res.Message = "All good with this API."
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

type HealthCheckResponse struct {
	Message string `json:"message"`
}

func main() {
	server := ez.NewWithOptions[any, HealthCheckResponse](&ez.ServerOptions{
		Addr: fmt.Sprintf(":%s", port),
	})
	server.RegisterRoute(ez.Route[any, HealthCheckResponse]{
		Handler:  MainHandler,
		Pattern:  "/",
		Method:   []string{http.MethodGet},
		Response: HealthCheckResponse{},
	})

	server.GenerateDocs()

	server.ListenAndServe()
}
