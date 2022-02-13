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
			reqSchema, err = openapi3gen.NewSchemaRefForValue(route.Request, nil)
			if err != nil {
				panic(err)
			}
			t := reflect.TypeOf(route.Request)
			components.Schemas[t.Name()] = reqSchema
		}
		if route.Response != nil {
			resSchema, err = openapi3gen.NewSchemaRefForValue(route.Response, nil)
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
	checkError(err)

	schema, err := yaml.JSONToYAML(b.Bytes())
	checkError(err)

	b = &bytes.Buffer{}
	b.Write(schema)

	doc, err := openapi3.NewLoader().LoadFromData(b.Bytes())
	checkError(err)

	jsonB, err := json.MarshalIndent(doc, "", "  ")
	checkError(err)
	err = ioutil.WriteFile("./openapi.json", jsonB, 0666)
	checkError(err)
	err = ioutil.WriteFile("./openapi.yaml", b.Bytes(), 0666)
	checkError(err)
}
