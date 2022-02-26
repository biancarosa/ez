package ez

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"sigs.k8s.io/yaml"
)

type DocsGenerator struct {
	server *EZServer
	docs   OpenAPIDocs
}

type OpenAPIDocs struct {
	Components openapi3.Components `json:"components,omitempty" yaml:"components,omitempty"`
	Paths      openapi3.Paths      `json:"paths,omitempty" yaml:"paths,omitempty"`
}

func (g *DocsGenerator) GenerateDocs() {
	fmt.Println("Generating docs...")

	components := openapi3.NewComponents()
	components.Schemas = make(map[string]*openapi3.SchemaRef)
	paths := make(map[string]*openapi3.PathItem)
	g.docs = OpenAPIDocs{
		Components: components,
		Paths:      paths,
	}

	for _, route := range g.server.GetRoutes() {
		g.GenerateDocsForRoute(route)
	}
	fmt.Println("Creating files...")

	g.GenerateOpenAPIFiles()

	fmt.Println("Docs generated!")
}

func (g *DocsGenerator) GenerateDocsForRoute(route Route) {
	var reqSchema *openapi3.SchemaRef
	var resSchema *openapi3.SchemaRef
	var err error
	if route.Request != nil {
		reqSchema, err := openapi3gen.NewSchemaRefForValue(route.Request, nil)
		if err != nil {
			panic(err)
		}
		t := reflect.TypeOf(route.Request)
		g.docs.Components.Schemas[t.Name()] = reqSchema
	}
	if route.Response != nil {
		resSchema, err = openapi3gen.NewSchemaRefForValue(route.Response, nil)
		if err != nil {
			panic(err)
		}
		t := reflect.TypeOf(route.Response)
		g.docs.Components.Schemas[t.Name()] = resSchema
	}
	g.docs.Paths[route.Pattern] = &openapi3.PathItem{}
	for _, method := range route.Method {
		if method == http.MethodGet {
			g.docs.Paths[route.Pattern].Get = &openapi3.Operation{
				OperationID: fmt.Sprintf("%s-%s", route.Pattern, method),
				Parameters:  []*openapi3.ParameterRef{},
				Responses:   map[string]*openapi3.ResponseRef{},
			}
		}
		if route.Request != nil {
			g.docs.Paths[route.Pattern].Get.RequestBody = &openapi3.RequestBodyRef{
				Value: &openapi3.RequestBody{
					Required: true,
				},
			}
			g.docs.Paths[route.Pattern].Get.RequestBody.Value.Content["application/json"] = &openapi3.MediaType{
				ExtensionProps: openapi3.ExtensionProps{},
				Schema:         reqSchema,
			}
		}
	}
}

func (g *DocsGenerator) GenerateOpenAPIFiles() {
	b := &bytes.Buffer{}
	err := json.NewEncoder(b).Encode(g.docs)
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
