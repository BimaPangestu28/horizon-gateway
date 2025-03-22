package validator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/xeipuuv/gojsonschema"
)

type RequestValidator struct {
	schemas      map[string]*gojsonschema.Schema
	logger       logging.Logger
	headerRules  map[string][]HeaderRule
	queryRules   map[string][]QueryRule
	contentTypes map[string][]string
}

type ValidationConfig struct {
	JSONSchemas  map[string]string           `yaml:"json_schemas" json:"json_schemas"`
	HeaderRules  map[string][]HeaderRule     `yaml:"header_rules" json:"header_rules"`
	QueryRules   map[string][]QueryRule      `yaml:"query_rules" json:"query_rules"`
	ContentTypes map[string][]string         `yaml:"content_types" json:"content_types"`
	PathRules    map[string][]PathRule       `yaml:"path_rules" json:"path_rules"`
	CustomRules  map[string][]map[string]any `yaml:"custom_rules" json:"custom_rules"`
}

type HeaderRule struct {
	Name     string `yaml:"name" json:"name"`
	Required bool   `yaml:"required" json:"required"`
	Pattern  string `yaml:"pattern" json:"pattern"`
}

type QueryRule struct {
	Name     string `yaml:"name" json:"name"`
	Required bool   `yaml:"required" json:"required"`
	Pattern  string `yaml:"pattern" json:"pattern"`
}

type PathRule struct {
	Pattern  string `yaml:"pattern" json:"pattern"`
	Required bool   `yaml:"required" json:"required"`
}

type ValidationResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Message string   `json:"message,omitempty"`
}

func NewRequestValidator(config *ValidationConfig, logger logging.Logger) (*RequestValidator, error) {
	validator := &RequestValidator{
		schemas:      make(map[string]*gojsonschema.Schema),
		logger:       logger,
		headerRules:  make(map[string][]HeaderRule),
		queryRules:   make(map[string][]QueryRule),
		contentTypes: make(map[string][]string),
	}

	if config != nil {
		for name, schemaStr := range config.JSONSchemas {
			schemaLoader := gojsonschema.NewStringLoader(schemaStr)
			schema, err := gojsonschema.NewSchema(schemaLoader)
			if err != nil {
				return nil, fmt.Errorf("failed to load JSON schema %s: %w", name, err)
			}
			validator.schemas[name] = schema
		}

		validator.headerRules = config.HeaderRules
		validator.queryRules = config.QueryRules
		validator.contentTypes = config.ContentTypes
	}

	return validator, nil
}

func (v *RequestValidator) ValidateRequest(req *http.Request, routeName string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	if headerRules, exists := v.headerRules[routeName]; exists {
		if err := v.validateHeaders(req, headerRules); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, err.Error())
		}
	}

	if queryRules, exists := v.queryRules[routeName]; exists {
		if err := v.validateQueryParams(req, queryRules); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, err.Error())
		}
	}

	if contentTypes, exists := v.contentTypes[routeName]; exists && len(contentTypes) > 0 {
		if err := v.validateContentType(req, contentTypes); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, err.Error())
		}
	}

	schema, exists := v.schemas[routeName]
	if exists {
		if err := v.validateBody(req, schema); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, err.Error())
		}
	}

	if !result.Valid {
		result.Message = fmt.Sprintf("Request validation failed for route %s", routeName)
	}

	return result, nil
}

func (v *RequestValidator) validateHeaders(req *http.Request, rules []HeaderRule) error {
	for _, rule := range rules {
		headerValue := req.Header.Get(rule.Name)

		if rule.Required && headerValue == "" {
			return fmt.Errorf("required header %s is missing", rule.Name)
		}

		if headerValue != "" && rule.Pattern != "" {
			match, err := regexp.MatchString(rule.Pattern, headerValue)
			if err != nil {
				return fmt.Errorf("error matching header pattern: %w", err)
			}

			if !match {
				return fmt.Errorf("header %s with value %s does not match pattern %s",
					rule.Name, headerValue, rule.Pattern)
			}
		}
	}

	return nil
}

func (v *RequestValidator) validateQueryParams(req *http.Request, rules []QueryRule) error {
	query := req.URL.Query()

	for _, rule := range rules {
		queryValues, exists := query[rule.Name]

		if rule.Required && (!exists || len(queryValues) == 0) {
			return fmt.Errorf("required query parameter %s is missing", rule.Name)
		}

		if exists && len(queryValues) > 0 && rule.Pattern != "" {
			for _, value := range queryValues {
				match, err := regexp.MatchString(rule.Pattern, value)
				if err != nil {
					return fmt.Errorf("error matching query parameter pattern: %w", err)
				}

				if !match {
					return fmt.Errorf("query parameter %s with value %s does not match pattern %s",
						rule.Name, value, rule.Pattern)
				}
			}
		}
	}

	return nil
}

func (v *RequestValidator) validateContentType(req *http.Request, allowedTypes []string) error {
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		return fmt.Errorf("missing Content-Type header")
	}

	contentTypeParts := strings.Split(contentType, ";")
	mainType := strings.TrimSpace(contentTypeParts[0])

	for _, allowedType := range allowedTypes {
		if strings.EqualFold(mainType, allowedType) {
			return nil
		}

		if strings.HasSuffix(allowedType, "/*") {
			prefix := strings.TrimSuffix(allowedType, "/*")
			if strings.HasPrefix(strings.ToLower(mainType), strings.ToLower(prefix)) {
				return nil
			}
		}
	}

	return fmt.Errorf("content type %s is not allowed. Allowed types: %s",
		mainType, strings.Join(allowedTypes, ", "))
}

func (v *RequestValidator) validateBody(req *http.Request, schema *gojsonschema.Schema) error {
	if req.Body == nil {
		return fmt.Errorf("request body is nil")
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("error reading request body: %w", err)
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if len(body) == 0 {
		return fmt.Errorf("request body is empty")
	}

	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return fmt.Errorf("request body is not valid JSON: %w", err)
	}

	documentLoader := gojsonschema.NewGoLoader(jsonData)
	result, err := schema.Validate(documentLoader)
	if err != nil {
		return fmt.Errorf("error validating request body: %w", err)
	}

	if !result.Valid() {
		var errMsgs []string
		for _, err := range result.Errors() {
			errMsgs = append(errMsgs, err.String())
		}
		return fmt.Errorf("request body validation failed: %s", strings.Join(errMsgs, "; "))
	}

	return nil
}

func (v *RequestValidator) AddSchema(name string, schemaStr string) error {
	schemaLoader := gojsonschema.NewStringLoader(schemaStr)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return fmt.Errorf("failed to load JSON schema: %w", err)
	}

	v.schemas[name] = schema
	return nil
}

func (v *RequestValidator) RemoveSchema(name string) {
	delete(v.schemas, name)
}

func (v *RequestValidator) AddHeaderRules(routeName string, rules []HeaderRule) {
	v.headerRules[routeName] = rules
}

func (v *RequestValidator) AddQueryRules(routeName string, rules []QueryRule) {
	v.queryRules[routeName] = rules
}

func (v *RequestValidator) AddContentTypes(routeName string, contentTypes []string) {
	v.contentTypes[routeName] = contentTypes
}

func (v *RequestValidator) GetSupportedSchemas() []string {
	schemas := make([]string, 0, len(v.schemas))
	for name := range v.schemas {
		schemas = append(schemas, name)
	}
	return schemas
}

func (v *RequestValidator) ValidateJSONData(data interface{}, schemaName string) (*ValidationResult, error) {
	schema, exists := v.schemas[schemaName]
	if !exists {
		return nil, fmt.Errorf("schema %s not found", schemaName)
	}

	result := &ValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	documentLoader := gojsonschema.NewGoLoader(data)
	validationResult, err := schema.Validate(documentLoader)
	if err != nil {
		return nil, fmt.Errorf("error validating JSON data: %w", err)
	}

	if !validationResult.Valid() {
		result.Valid = false
		for _, err := range validationResult.Errors() {
			result.Errors = append(result.Errors, err.String())
		}
		result.Message = "JSON validation failed"
	}

	return result, nil
}
