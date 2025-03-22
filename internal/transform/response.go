// internal/transform/response.go
package transform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"text/template"

	"github.com/tidwall/sjson"
)

// ResponseTransformer implements response transformations
type ResponseTransformer struct {
	config            *TransformConfig
	headerTransforms  []CompiledHeaderTransform
	bodyTransform     *CompiledBodyTransform
	statusCodeMap     map[int]int
	defaultStatusCode int
	templateFuncs     template.FuncMap
}

// NewResponseTransformer creates a new response transformer
func NewResponseTransformer(config *TransformConfig) (*ResponseTransformer, error) {
	transformer := &ResponseTransformer{
		config:        config,
		templateFuncs: GetTemplateFuncs(),
	}

	// Skip if there are no response transformations
	if config.Response == nil {
		return transformer, nil
	}

	// Compile header transforms
	if len(config.Response.Headers) > 0 {
		for _, ht := range config.Response.Headers {
			compiled := CompiledHeaderTransform{
				Name:      ht.Name,
				Operation: ht.Operation,
				Value:     ht.Value,
			}

			// Compile regex if needed
			if ht.Operation == TransformOperationRegex && ht.Pattern != "" {
				regex, err := regexp.Compile(ht.Pattern)
				if err != nil {
					return nil, fmt.Errorf("invalid header regex pattern '%s': %w", ht.Pattern, err)
				}
				compiled.RegexPattern = regex
				compiled.Replacement = ht.Replacement
			}

			transformer.headerTransforms = append(transformer.headerTransforms, compiled)
		}
	}

	// Compile body transform
	if config.Response.Body != nil {
		body := config.Response.Body
		compiled := &CompiledBodyTransform{
			Operation: body.Operation,
			Template:  body.Template,
			JSONPath:  body.JSONPath,
		}

		// Build content type map
		if len(body.ContentTypes) > 0 {
			compiled.ContentTypes = make(map[string]bool)
			for _, ct := range body.ContentTypes {
				compiled.ContentTypes[ct] = true
			}
		}

		// Compile regex if needed
		if body.Operation == TransformOperationRegex && body.Pattern != "" {
			regex, err := regexp.Compile(body.Pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid body regex pattern '%s': %w", body.Pattern, err)
			}
			compiled.RegexPattern = regex
			compiled.Replacement = body.Replacement
		}

		transformer.bodyTransform = compiled
	}

	// Set up status code transforms
	if config.Response.StatusCode != nil {
		statusCode := config.Response.StatusCode

		// Set up status code mapping
		if len(statusCode.Mapping) > 0 {
			transformer.statusCodeMap = statusCode.Mapping
		}

		// Set default status code
		transformer.defaultStatusCode = statusCode.Default
	}

	return transformer, nil
}

// TransformResponse applies transformations to a response
func (t *ResponseTransformer) TransformResponse(resp *http.Response) error {
	// Skip if no transformations
	if t.config.Response == nil {
		return nil
	}

	// Transform headers
	if len(t.headerTransforms) > 0 {
		if err := t.transformHeaders(resp); err != nil {
			return fmt.Errorf("header transformation failed: %w", err)
		}
	}

	// Transform status code
	if t.statusCodeMap != nil || t.defaultStatusCode > 0 {
		t.transformStatusCode(resp)
	}

	// Transform body
	if t.bodyTransform != nil {
		if err := t.transformBody(resp); err != nil {
			return fmt.Errorf("body transformation failed: %w", err)
		}
	}

	return nil
}

// transformHeaders applies header transformations
func (t *ResponseTransformer) transformHeaders(resp *http.Response) error {
	for _, transform := range t.headerTransforms {
		switch transform.Operation {
		case TransformOperationAdd:
			// Only add if not exists
			if resp.Header.Get(transform.Name) == "" {
				resp.Header.Set(transform.Name, transform.Value)
			}

		case TransformOperationReplace:
			// Replace regardless of existence
			resp.Header.Set(transform.Name, transform.Value)

		case TransformOperationRemove:
			// Remove header
			resp.Header.Del(transform.Name)

		case TransformOperationRegex:
			// Apply regex to existing header
			if value := resp.Header.Get(transform.Name); value != "" && transform.RegexPattern != nil {
				newValue := transform.RegexPattern.ReplaceAllString(value, transform.Replacement)
				resp.Header.Set(transform.Name, newValue)
			}
		}
	}

	return nil
}

// transformStatusCode applies status code transformations
func (t *ResponseTransformer) transformStatusCode(resp *http.Response) {
	// Apply status code mapping if exists
	if t.statusCodeMap != nil {
		if newCode, exists := t.statusCodeMap[resp.StatusCode]; exists {
			resp.StatusCode = newCode
		}
	}

	// Apply default status code if specified
	if t.defaultStatusCode > 0 {
		resp.StatusCode = t.defaultStatusCode
	}
}

// transformBody applies body transformations
func (t *ResponseTransformer) transformBody(resp *http.Response) error {
	// Skip if no body
	if resp.Body == nil {
		return nil
	}

	// Check content type if specified
	if len(t.bodyTransform.ContentTypes) > 0 {
		contentType := resp.Header.Get("Content-Type")
		// Extract base content type without parameters
		if i := strings.Index(contentType, ";"); i > 0 {
			contentType = contentType[:i]
		}
		contentType = strings.TrimSpace(contentType)

		// Skip if content type doesn't match
		if !t.bodyTransform.ContentTypes[contentType] {
			return nil
		}
	}

	// Read body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Close original body
	resp.Body.Close()

	// Apply transformations
	var newBody []byte

	switch t.bodyTransform.Operation {
	case TransformOperationRegex:
		if t.bodyTransform.RegexPattern != nil {
			newBody = t.bodyTransform.RegexPattern.ReplaceAll(body, []byte(t.bodyTransform.Replacement))
		} else {
			newBody = body
		}

	case TransformOperationTemplate:
		if t.bodyTransform.Template != "" {
			// Create template context
			data := map[string]interface{}{
				"Body":       string(body),
				"StatusCode": resp.StatusCode,
				"Headers":    resp.Header,
			}

			// Try to parse body as JSON if it looks like JSON
			if len(body) > 0 && (body[0] == '{' || body[0] == '[') {
				var jsonData interface{}
				if json.Unmarshal(body, &jsonData) == nil {
					data["JSON"] = jsonData
				}
			}

			// Parse and execute template
			tmpl, err := template.New("body").
				Funcs(t.templateFuncs).
				Parse(t.bodyTransform.Template)
			if err != nil {
				return err
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				return err
			}

			newBody = buf.Bytes()
		} else {
			newBody = body
		}

	case TransformOperationJSONPath:
		if len(t.bodyTransform.JSONPath) > 0 && isJSON(body) {
			// Use sjson/gjson for JSONPath operations
			jsonStr := string(body)

			for _, jp := range t.bodyTransform.JSONPath {
				switch jp.Operation {
				case TransformOperationAdd, TransformOperationReplace:
					var err error
					// Handle different value types
					switch v := jp.Value.(type) {
					case string, bool, float64, int, int64:
						jsonStr, err = sjson.Set(jsonStr, jp.Path, v)
					default:
						// For complex objects, convert through JSON
						jsonStr, err = sjson.SetRaw(jsonStr, jp.Path, fmt.Sprintf("%v", v))
					}
					if err != nil {
						return fmt.Errorf("JSONPath set operation failed: %w", err)
					}

				case TransformOperationRemove:
					jsonStr, err = sjson.Delete(jsonStr, jp.Path)
					if err != nil {
						return fmt.Errorf("JSONPath delete operation failed: %w", err)
					}
				}
			}

			newBody = []byte(jsonStr)
		} else {
			newBody = body
		}

	default:
		newBody = body
	}

	// Create new body reader
	resp.Body = ioutil.NopCloser(bytes.NewReader(newBody))
	resp.ContentLength = int64(len(newBody))

	// Update Content-Length header if it exists
	if resp.Header.Get("Content-Length") != "" {
		resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))
	}

	return nil
}

// AddExtraInfo adds additional information to a response (metadata, timestamps, etc.)
func (t *ResponseTransformer) AddExtraInfo(resp *http.Response, info map[string]interface{}) error {
	// Only apply to JSON responses
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return nil
	}

	// Read body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// Only process valid JSON
	if !isJSON(body) {
		resp.Body = ioutil.NopCloser(bytes.NewReader(body))
		return nil
	}

	// Parse JSON
	var respData map[string]interface{}
	if err := json.Unmarshal(body, &respData); err != nil {
		resp.Body = ioutil.NopCloser(bytes.NewReader(body))
		return nil
	}

	// Create result with metadata wrapper
	result := map[string]interface{}{
		"data": respData,
		"meta": info,
	}

	// Serialize back to JSON
	newBody, err := json.Marshal(result)
	if err != nil {
		resp.Body = ioutil.NopCloser(bytes.NewReader(body))
		return err
	}

	// Update response
	resp.Body = ioutil.NopCloser(bytes.NewReader(newBody))
	resp.ContentLength = int64(len(newBody))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))

	return nil
}

// AddSecurityHeaders adds common security headers to a response
func (t *ResponseTransformer) AddSecurityHeaders(resp *http.Response) {
	// Common security headers
	securityHeaders := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Content-Security-Policy":   "default-src 'self'",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
	}

	// Add headers if not already present
	for name, value := range securityHeaders {
		if resp.Header.Get(name) == "" {
			resp.Header.Set(name, value)
		}
	}
}

// ReplaceErrorWithCustomResponse replaces error responses with custom templates
func (t *ResponseTransformer) ReplaceErrorWithCustomResponse(resp *http.Response, templates map[int]string) error {
	// Check if status code has a custom template
	template, exists := templates[resp.StatusCode]
	if !exists || resp.StatusCode < 400 {
		return nil
	}

	// Read original body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// Create template context
	data := map[string]interface{}{
		"StatusCode":   resp.StatusCode,
		"StatusText":   http.StatusText(resp.StatusCode),
		"Headers":      resp.Header,
		"OriginalBody": string(body),
	}

	// Parse and execute template
	tmpl, err := template.New("error").
		Funcs(t.templateFuncs).
		Parse(template)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	// Update response
	newBody := buf.Bytes()
	resp.Body = ioutil.NopCloser(bytes.NewReader(newBody))
	resp.ContentLength = int64(len(newBody))
	resp.Header.Set("Content-Type", "application/json")
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))

	return nil
}
