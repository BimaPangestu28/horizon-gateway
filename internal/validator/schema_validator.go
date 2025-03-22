package validator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"

type SchemaValidator struct {
	swagger    *openapi3.T
	router     routers.Router
	logger     logging.Logger
	operations map[string]*openapi3.Operation
}

type OpenAPIValidationConfig struct {
	SchemaPath        string            `yaml:"schema_path" json:"schema_path"`
	SchemaContent     string            `yaml:"schema_content" json:"schema_content"`
	DisableValidation bool              `yaml:"disable_validation" json:"disable_validation"`
	PathMapping       map[string]string `yaml:"path_mapping" json:"path_mapping"`
}

type SchemaValidationResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Message string   `json:"message,omitempty"`
}

func NewSchemaValidator(config *OpenAPIValidationConfig, logger logging.Logger) (*SchemaValidator, error) {
	if config == nil || (config.SchemaPath == "" && config.SchemaContent == "") {
		return nil, fmt.Errorf("OpenAPI schema is required")
	}

	var swagger *openapi3.T
	var err error

	if config.SchemaContent != "" {
		swagger, err = openapi3.NewLoader().LoadFromData([]byte(config.SchemaContent))
	} else {
		swagger, err = openapi3.NewLoader().LoadFromFile(config.SchemaPath)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI schema: %w", err)
	}

	if err := swagger.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI schema: %w", err)
	}

	router, err := legacyrouter.NewRouter(swagger)
	if err != nil {
		return nil, fmt.Errorf("failed to create router: %w", err)
	}

	operations := make(map[string]*openapi3.Operation)
	for path, pathItem := range swagger.Paths {
		for method, operation := range pathItem.Operations() {
			operationID := method + " " + path
			if operation.OperationID != "" {
				operationID = operation.OperationID
			}
			operations[operationID] = operation
		}
	}

	return &SchemaValidator{
		swagger:    swagger,
		router:     router,
		logger:     logger,
		operations: operations,
	}, nil
}

func (v *SchemaValidator) ValidateRequest(req *http.Request) (*SchemaValidationResult, error) {
	result := &SchemaValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	route, pathParams, err := v.router.FindRoute(req)
	if err != nil {
		return nil, fmt.Errorf("route not found: %w", err)
	}

	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	if err := openapi3filter.ValidateRequest(context.Background(), requestValidationInput); err != nil {
		result.Valid = false
		result.Message = "Request validation failed"

		switch e := err.(type) {
		case *openapi3filter.RequestError:
			result.Errors = append(result.Errors, e.Error())
		case *openapi3filter.SecurityRequirementsError:
			for _, err := range e.Errors {
				result.Errors = append(result.Errors, err.Error())
			}
		default:
			result.Errors = append(result.Errors, err.Error())
		}
	}

	return result, nil
}

func (v *SchemaValidator) ValidatePayload(operationID string, payload []byte, payloadType string) (*SchemaValidationResult, error) {
	result := &SchemaValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	operation, exists := v.operations[operationID]
	if !exists {
		return nil, fmt.Errorf("operation %s not found", operationID)
	}

	var schema *openapi3.Schema

	switch payloadType {
	case "request":
		if operation.RequestBody == nil || operation.RequestBody.Value == nil {
			return nil, fmt.Errorf("no request body schema defined for operation %s", operationID)
		}
		
		contentType := "application/json"
		mediaType, exists := operation.RequestBody.Value.Content[contentType]
		if !exists {
			return nil, fmt.Errorf("no %s content type defined for operation %s", contentType, operationID)
		}
		
		if mediaType.Schema == nil {
			return nil, fmt.Errorf("no schema defined for %s in operation %s", contentType, operationID)
		}
		
		schema = mediaType.Schema.Value
	case "response":
		response, exists := operation.Responses.Map()["200"]
		if !exists {
			response, exists = operation.Responses.Map()["201"]
			if !exists {
				return nil, fmt.Errorf("no 200/201 response defined for operation %s", operationID)
			}
		}
		
		if response.Value.Content == nil {
			return nil, fmt.Errorf("no content defined for response in operation %s", operationID)
		}
		
		mediaType, exists := response.Value.Content["application/json"]
		if !exists {
			return nil, fmt.Errorf("no application/json content type defined for response in operation %s", operationID)
		}
		
		if mediaType.Schema == nil {
			return nil, fmt.Errorf("no schema defined for application/json in response for operation %s", operationID)
		}
		
		schema = mediaType.Schema.Value
	default:
		return nil, fmt.Errorf("invalid payload type: %s", payloadType)
	}

	var jsonData interface{}
	if err := json.Unmarshal(payload, &jsonData); err != nil {
		result.Valid = false
		result.Message = "Invalid JSON payload"
		result.Errors = append(result.Errors, err.Error())
		return result, nil
	}

	err := openapi3.ValidateValue(context.Background(), schema, jsonData)
	if err != nil {
		result.Valid = false
		result.Message = fmt.Sprintf("%s validation failed", payloadType)
		
		switch e := err.(type) {
		case *openapi3.SchemaError:
			result.Errors = append(result.Errors, e.Error())
		default:
			result.Errors = append(result.Errors, err.Error())
		}
	}

	return result, nil
}

func (v *SchemaValidator) GetOperationIDs() []string {
	operationIDs := make([]string, 0, len(v.operations))
	for id := range v.operations {
		operationIDs = append(operationIDs, id)
	}
	return operationIDs
}

func (v *SchemaValidator) MapPathToOperation(method, path string) (string, error) {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	route, _, err := v.router.FindRoute(req)
	if err != nil {
		return "", fmt.Errorf("route not found: %w", err)
	}

	if route.Operation.OperationID != "" {
		return route.Operation.OperationID, nil
	}

	return method + " " + route.Path, nil
}

func (v *SchemaValidator) GetRoutePatterns() []string {
	patterns := make([]string, 0)
	
	for path := range v.swagger.Paths {
		patterns = append(patterns, path)
	}
	
	return patterns
}
}

func (v *SchemaValidator) ValidateResponse(req *http.Request, statusCode int, responseBody []byte) (*SchemaValidationResult, error) {
	result := &SchemaValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	route, pathParams, err := v.router.FindRoute(req)
	if err != nil {
		return nil, fmt.Errorf("route not found: %w", err)
	}

	responseHeaders := make(http.Header)
	responseHeaders.Set("Content-Type", "application/json")

	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
		},
		Status:   statusCode,
		Header:   responseHeaders,
		Body:     io.NopCloser(bytes.NewReader(responseBody)),
	}

	if err := openapi3filter.ValidateResponse(context.Background(), responseValidationInput); err != nil {
		result.Valid = false
		result.Message = "Response validation failed"
		result.Errors = append(result.Errors, err.Error())
	}

	return result, nil
}