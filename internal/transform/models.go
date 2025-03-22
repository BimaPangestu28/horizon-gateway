package transform

import (
	"net/http"
	"regexp"
)

// TransformType represents the type of transformation
type TransformType string

const (
	TransformTypeHeader       TransformType = "header"
	TransformTypeURL          TransformType = "url"
	TransformTypeRequestBody  TransformType = "request_body"
	TransformTypeResponseBody TransformType = "response_body"
	TransformTypeContentType  TransformType = "content_type"
)

// TransformOperation represents the operation to perform
type TransformOperation string

const (
	TransformOperationAdd      TransformOperation = "add"
	TransformOperationRemove   TransformOperation = "remove"
	TransformOperationReplace  TransformOperation = "replace"
	TransformOperationRegex    TransformOperation = "regex"
	TransformOperationTemplate TransformOperation = "template"
	TransformOperationJSONPath TransformOperation = "jsonpath"
)

// TransformConfig defines request/response transformation configuration
type TransformConfig struct {
	Enabled  bool                     `yaml:"enabled"`
	Request  *RequestTransformConfig  `yaml:"request,omitempty"`
	Response *ResponseTransformConfig `yaml:"response,omitempty"`
}

// RequestTransformConfig defines request transformation rules
type RequestTransformConfig struct {
	Headers []HeaderTransform `yaml:"headers,omitempty"`
	URL     *URLTransform     `yaml:"url,omitempty"`
	Body    *BodyTransform    `yaml:"body,omitempty"`
}

// ResponseTransformConfig defines response transformation rules
type ResponseTransformConfig struct {
	Headers    []HeaderTransform    `yaml:"headers,omitempty"`
	Body       *BodyTransform       `yaml:"body,omitempty"`
	StatusCode *StatusCodeTransform `yaml:"status_code,omitempty"`
}

// HeaderTransform defines a header transformation
type HeaderTransform struct {
	Name        string             `yaml:"name"`
	Operation   TransformOperation `yaml:"operation"`
	Value       string             `yaml:"value,omitempty"`
	Pattern     string             `yaml:"pattern,omitempty"`
	Replacement string             `yaml:"replacement,omitempty"`
}

// URLTransform defines a URL transformation
type URLTransform struct {
	Path  *URLPathTransform     `yaml:"path,omitempty"`
	Query []QueryParamTransform `yaml:"query,omitempty"`
}

// URLPathTransform defines a URL path transformation
type URLPathTransform struct {
	Operation   TransformOperation `yaml:"operation"`
	Pattern     string             `yaml:"pattern,omitempty"`
	Replacement string             `yaml:"replacement,omitempty"`
	Template    string             `yaml:"template,omitempty"`
}

// QueryParamTransform defines a query parameter transformation
type QueryParamTransform struct {
	Name      string             `yaml:"name"`
	Operation TransformOperation `yaml:"operation"`
	Value     string             `yaml:"value,omitempty"`
}

// BodyTransform defines a body transformation
type BodyTransform struct {
	Operation    TransformOperation  `yaml:"operation"`
	ContentTypes []string            `yaml:"content_types,omitempty"`
	Template     string              `yaml:"template,omitempty"`
	Pattern      string              `yaml:"pattern,omitempty"`
	Replacement  string              `yaml:"replacement,omitempty"`
	JSONPath     []JSONPathTransform `yaml:"jsonpath,omitempty"`
}

// JSONPathTransform defines a JSONPath transformation
type JSONPathTransform struct {
	Path      string             `yaml:"path"`
	Operation TransformOperation `yaml:"operation"`
	Value     interface{}        `yaml:"value,omitempty"`
}

// StatusCodeTransform defines a status code transformation
type StatusCodeTransform struct {
	Operation TransformOperation `yaml:"operation"`
	Mapping   map[int]int        `yaml:"mapping,omitempty"`
	Default   int                `yaml:"default,omitempty"`
}

// Transformer interface defines transformation functionality
type Transformer interface {
	TransformRequest(req *http.Request) error
	TransformResponse(resp *http.Response) error
}

// Compiled transform types for efficient runtime operations
type CompiledHeaderTransform struct {
	Name         string
	Operation    TransformOperation
	Value        string
	RegexPattern *regexp.Regexp
	Replacement  string
}

type CompiledURLPathTransform struct {
	Operation    TransformOperation
	RegexPattern *regexp.Regexp
	Replacement  string
	Template     string
}

type CompiledBodyTransform struct {
	Operation    TransformOperation
	ContentTypes map[string]bool
	Template     string
	RegexPattern *regexp.Regexp
	Replacement  string
	JSONPath     []JSONPathTransform
}
