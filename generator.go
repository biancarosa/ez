package ez

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

func (g *DocsGenerator) GenerateDocs() error {
	fmt.Println("Generating docs...")

	components := openapi3.NewComponents()
	components.Schemas = make(map[string]*openapi3.SchemaRef)
	paths := make(map[string]*openapi3.PathItem)
	g.docs = OpenAPIDocs{
		Components: components,
		Paths:      paths,
	}

	for _, route := range g.server.GetRoutes() {
		if err := g.GenerateDocsForRoute(route); err != nil {
			return fmt.Errorf("failed to generate docs for route %s: %w", route.Pattern, err)
		}
	}
	fmt.Println("Creating files...")

	if err := g.GenerateOpenAPIFiles(); err != nil {
		return fmt.Errorf("failed to generate OpenAPI files: %w", err)
	}

	fmt.Println("Docs generated!")
	return nil
}

func (g *DocsGenerator) GenerateDocsForRoute(route Route) error {
	var reqSchema *openapi3.SchemaRef
	var resSchema *openapi3.SchemaRef
	var err error
	if route.Request != nil {
		reqSchema, err = openapi3gen.NewSchemaRefForValue(route.Request, nil)
		if err != nil {
			return fmt.Errorf("failed to generate request schema: %w", err)
		}
		t := reflect.TypeOf(route.Request)
		g.docs.Components.Schemas[t.Name()] = reqSchema
	}
	if route.Response != nil {
		resSchema, err = openapi3gen.NewSchemaRefForValue(route.Response, nil)
		if err != nil {
			return fmt.Errorf("failed to generate response schema: %w", err)
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
	return nil
}

func (g *DocsGenerator) GenerateOpenAPIFiles() error {
	b := &bytes.Buffer{}
	err := json.NewEncoder(b).Encode(g.docs)
	if err != nil {
		return fmt.Errorf("failed to encode docs to JSON: %w", err)
	}

	schema, err := yaml.JSONToYAML(b.Bytes())
	if err != nil {
		return fmt.Errorf("failed to convert JSON to YAML: %w", err)
	}

	b = &bytes.Buffer{}
	b.Write(schema)

	doc, err := openapi3.NewLoader().LoadFromData(b.Bytes())
	if err != nil {
		return fmt.Errorf("failed to load OpenAPI doc: %w", err)
	}

	jsonB, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	err = os.WriteFile("./openapi.json", jsonB, 0666)
	if err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	err = os.WriteFile("./openapi.yaml", b.Bytes(), 0666)
	if err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}

	return nil
}
