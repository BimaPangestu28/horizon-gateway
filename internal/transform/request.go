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

// RequestTransformer implements request transformations
type RequestTransformer struct {
	config           *TransformConfig
	headerTransforms []CompiledHeaderTransform
	urlPathTransform *CompiledURLPathTransform
	queryTransforms  []QueryParamTransform
	bodyTransform    *CompiledBodyTransform
	templateFuncs    template.FuncMap
}

// NewRequestTransformer creates a new request transformer
func NewRequestTransformer(config *TransformConfig) (*RequestTransformer, error) {
	transformer := &RequestTransformer{
		config:        config,
		templateFuncs: GetTemplateFuncs(),
	}

	// Skip if there are no request transformations
	if config.Request == nil {
		return transformer, nil
	}

	// Compile header transforms
	if len(config.Request.Headers) > 0 {
		for _, ht := range config.Request.Headers {
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

	// Compile URL path transform
	if config.Request.URL != nil && config.Request.URL.Path != nil {
		path := config.Request.URL.Path
		compiled := &CompiledURLPathTransform{
			Operation: path.Operation,
			Template:  path.Template,
		}

		// Compile regex if needed
		if path.Operation == TransformOperationRegex && path.Pattern != "" {
			regex, err := regexp.Compile(path.Pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid URL path regex pattern '%s': %w", path.Pattern, err)
			}
			compiled.RegexPattern = regex
			compiled.Replacement = path.Replacement
		}

		transformer.urlPathTransform = compiled
	}

	// Store query transforms (no compilation needed)
	if config.Request.URL != nil && len(config.Request.URL.Query) > 0 {
		transformer.queryTransforms = config.Request.URL.Query
	}

	// Compile body transform
	if config.Request.Body != nil {
		body := config.Request.Body
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

	return transformer, nil
}

// TransformRequest applies transformations to a request
func (t *RequestTransformer) TransformRequest(req *http.Request) error {
	// Skip if no transformations
	if t.config.Request == nil {
		return nil
	}

	// Transform headers
	if len(t.headerTransforms) > 0 {
		if err := t.transformHeaders(req); err != nil {
			return fmt.Errorf("header transformation failed: %w", err)
		}
	}

	// Transform URL path
	if t.urlPathTransform != nil {
		if err := t.transformURLPath(req); err != nil {
			return fmt.Errorf("URL path transformation failed: %w", err)
		}
	}

	// Transform query parameters
	if len(t.queryTransforms) > 0 {
		if err := t.transformQuery(req); err != nil {
			return fmt.Errorf("query transformation failed: %w", err)
		}
	}

	// Transform body
	if t.bodyTransform != nil {
		if err := t.transformBody(req); err != nil {
			return fmt.Errorf("body transformation failed: %w", err)
		}
	}

	return nil
}

// transformHeaders applies header transformations
func (t *RequestTransformer) transformHeaders(req *http.Request) error {
	for _, transform := range t.headerTransforms {
		switch transform.Operation {
		case TransformOperationAdd:
			// Only add if not exists
			if req.Header.Get(transform.Name) == "" {
				req.Header.Set(transform.Name, transform.Value)
			}

		case TransformOperationReplace:
			// Replace regardless of existence
			req.Header.Set(transform.Name, transform.Value)

		case TransformOperationRemove:
			// Remove header
			req.Header.Del(transform.Name)

		case TransformOperationRegex:
			// Apply regex to existing header
			if value := req.Header.Get(transform.Name); value != "" && transform.RegexPattern != nil {
				newValue := transform.RegexPattern.ReplaceAllString(value, transform.Replacement)
				req.Header.Set(transform.Name, newValue)
			}
		}
	}

	return nil
}

// transformURLPath applies URL path transformations
func (t *RequestTransformer) transformURLPath(req *http.Request) error {
	switch t.urlPathTransform.Operation {
	case TransformOperationRegex:
		if t.urlPathTransform.RegexPattern != nil {
			path := req.URL.Path
			newPath := t.urlPathTransform.RegexPattern.ReplaceAllString(path, t.urlPathTransform.Replacement)
			req.URL.Path = newPath
		}

	case TransformOperationTemplate:
		if t.urlPathTransform.Template != "" {
			// Create template context
			data := map[string]interface{}{
				"Path":    req.URL.Path,
				"Method":  req.Method,
				"Headers": req.Header,
				"Query":   req.URL.Query(),
			}

			// Parse and execute template
			tmpl, err := template.New("path").
				Funcs(t.templateFuncs).
				Parse(t.urlPathTransform.Template)
			if err != nil {
				return err
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				return err
			}

			req.URL.Path = buf.String()
		}
	}

	return nil
}

// transformQuery applies query parameter transformations
func (t *RequestTransformer) transformQuery(req *http.Request) error {
	// Parse query parameters
	query := req.URL.Query()

	for _, transform := range t.queryTransforms {
		switch transform.Operation {
		case TransformOperationAdd:
			// Only add if not exists
			if query.Get(transform.Name) == "" {
				query.Add(transform.Name, transform.Value)
			}

		case TransformOperationReplace:
			// Replace regardless of existence
			query.Set(transform.Name, transform.Value)

		case TransformOperationRemove:
			// Remove parameter
			query.Del(transform.Name)
		}
	}

	// Update request URL
	req.URL.RawQuery = query.Encode()
	return nil
}

// transformBody applies body transformations
func (t *RequestTransformer) transformBody(req *http.Request) error {
	// Skip if no body
	if req.Body == nil {
		return nil
	}

	// Check content type if specified
	if len(t.bodyTransform.ContentTypes) > 0 {
		contentType := req.Header.Get("Content-Type")
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
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	// Close original body
	req.Body.Close()

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
				"Body":    string(body),
				"Path":    req.URL.Path,
				"Method":  req.Method,
				"Headers": req.Header,
				"Query":   req.URL.Query(),
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
			// Using sjson/gjson for JSONPath operations which is simpler than
			// full JSONPath implementation but covers most common use cases
			jsonStr := string(body)

			for _, jp := range t.bodyTransform.JSONPath {
				switch jp.Operation {
				case TransformOperationAdd, TransformOperationReplace:
					var err error
					// Need to handle different value types
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
	req.Body = ioutil.NopCloser(bytes.NewReader(newBody))
	req.ContentLength = int64(len(newBody))

	// Update Content-Length header if it exists
	if req.Header.Get("Content-Length") != "" {
		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))
	}

	return nil
}
