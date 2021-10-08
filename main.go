package ez

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"sigs.k8s.io/yaml"
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

func (ez *EZServer) ListenAndServe() {
	err := ez.s.ListenAndServe()

	fmt.Println("Running server on ", ez.s.Addr)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Shutdown") // Todo register shutdown method
}

func (ez *EZServer) GenerateDocs() {
	components := openapi3.NewComponents()
	components.Schemas = make(map[string]*openapi3.SchemaRef)
	paths := make(map[string]*openapi3.PathItem)
	for _, route := range ez.GetRoutes() {
		var reqSchema *openapi3.SchemaRef
		var resSchema *openapi3.SchemaRef
		var err error
		if route.Request != nil {
			reqSchema, _, err = openapi3gen.NewSchemaRefForValue(route.Request)
			if err != nil {
				panic(err)
			}
			t := reflect.TypeOf(route.Request)
			components.Schemas[t.Name()] = reqSchema
		}
		if route.Response != nil {
			resSchema, _, err = openapi3gen.NewSchemaRefForValue(route.Response)
			if err != nil {
				panic(err)
			}
			t := reflect.TypeOf(route.Response)
			components.Schemas[t.Name()] = resSchema
		}
		paths[route.Pattern] = &openapi3.PathItem{}
		for _, method := range route.Method {
			if method == http.MethodGet {
				paths[route.Pattern].Get = &openapi3.Operation{
					OperationID: fmt.Sprintf("%s-%s", route.Pattern, method),
					Parameters:  []*openapi3.ParameterRef{},
					Responses:   map[string]*openapi3.ResponseRef{},
				}
			}
			if route.Request != nil {
				paths[route.Pattern].Get.RequestBody = &openapi3.RequestBodyRef{
					Value: &openapi3.RequestBody{
						Required: true,
					},
				}
				paths[route.Pattern].Get.RequestBody.Value.Content["application/json"] = &openapi3.MediaType{
					ExtensionProps: openapi3.ExtensionProps{},
					Schema:         reqSchema,
				}
			}
		}
	}

	type Swagger struct {
		Components openapi3.Components `json:"components,omitempty" yaml:"components,omitempty"`
		Paths      openapi3.Paths      `json:"paths,omitempty" yaml:"paths,omitempty"`
	}

	swagger := Swagger{}
	swagger.Components = components
	swagger.Paths = paths
	b := &bytes.Buffer{}
	err := json.NewEncoder(b).Encode(swagger)
	checkErr(err)

	schema, err := yaml.JSONToYAML(b.Bytes())
	checkErr(err)

	b = &bytes.Buffer{}
	b.Write(schema)

	doc, err := openapi3.NewLoader().LoadFromData(b.Bytes())
	checkErr(err)

	jsonB, err := json.MarshalIndent(doc, "", "  ")
	checkErr(err)
	err = ioutil.WriteFile("./openapi.json", jsonB, 0666)
	checkErr(err)
	err = ioutil.WriteFile("./openapi.yaml", b.Bytes(), 0666)
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
