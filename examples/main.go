package examples

import (
	"fmt"
	"net/http"

	"github.com/biancarosa/ez"
)

const (
	port = "5000"
)

func MainHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "Okay!")
}

func main() {
	server := http.Server{
		Addr: fmt.Sprintf(":%s", port),
	}

	fmt.Println("Running server on port", port)
	routes := []ez.Route{
		{
			Handler: MainHandler,
			Pattern: "/",
		}}
	ez.RegisterRoutes(routes)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Shutdown")
}
