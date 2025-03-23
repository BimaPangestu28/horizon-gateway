package transform

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/Jeffail/gabs/v2"
	"github.com/PaesslerAG/jsonpath"
	"github.com/antchfx/xmlquery"
	"github.com/clbanning/mxj/v2"
	"github.com/jmoiron/jsonq"
	luajson "github.com/layeh/gopher-json"
	lua "github.com/yuin/gopher-lua"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

var (
	ErrTransformFailed          = errors.New("transform failed")
	ErrInvalidTransformInput    = errors.New("invalid transform input")
	ErrTransformTemplateInvalid = errors.New("transform template invalid")
	ErrJSONPathInvalid          = errors.New("jsonpath expression invalid")
	ErrXPathInvalid             = errors.New("xpath expression invalid")
	ErrScriptExecutionFailed    = errors.New("script execution failed")
)

type AdvancedTransformer struct {
	config *AdvancedTransformConfig
	logger logging.Logger
}

type AdvancedTransformConfig struct {
	Type           string            `yaml:"type" json:"type"`
	JSONPath       JSONPathTransform `yaml:"jsonpath,omitempty" json:"jsonpath,omitempty"`
	XPath          XPathTransform    `yaml:"xpath,omitempty" json:"xpath,omitempty"`
	Template       TemplateTransform `yaml:"template,omitempty" json:"template,omitempty"`
	LuaScript      ScriptTransform   `yaml:"lua_script,omitempty" json:"lua_script,omitempty"`
	RegexReplace   RegexTransform    `yaml:"regex,omitempty" json:"regex,omitempty"`
	ContentTypes   []string          `yaml:"content_types,omitempty" json:"content_types,omitempty"`
	RequestMapping []MappingRule     `yaml:"request_mapping,omitempty" json:"request_mapping,omitempty"`
	Aggregation    AggregationConfig `yaml:"aggregation,omitempty" json:"aggregation,omitempty"`
	Protocol       ProtocolTransform `yaml:"protocol,omitempty" json:"protocol,omitempty"`
}

type JSONPathTransform struct {
	Operations []JSONPathOperation `yaml:"operations" json:"operations"`
}

type JSONPathOperation struct {
	Path      string      `yaml:"path" json:"path"`
	Operation string      `yaml:"operation" json:"operation"`
	Value     interface{} `yaml:"value,omitempty" json:"value,omitempty"`
	ValueFrom string      `yaml:"value_from,omitempty" json:"value_from,omitempty"`
	Condition string      `yaml:"condition,omitempty" json:"condition,omitempty"`
}

type XPathTransform struct {
	Operations []XPathOperation `yaml:"operations" json:"operations"`
}

type XPathOperation struct {
	Path      string      `yaml:"path" json:"path"`
	Operation string      `yaml:"operation" json:"operation"`
	Value     interface{} `yaml:"value,omitempty" json:"value,omitempty"`
	ValueFrom string      `yaml:"value_from,omitempty" json:"value_from,omitempty"`
	Condition string      `yaml:"condition,omitempty" json:"condition,omitempty"`
}

type TemplateTransform struct {
	Template   string `yaml:"template" json:"template"`
	Engine     string `yaml:"engine" json:"engine"` // go, handlebars, etc.
	EscapeHTML bool   `yaml:"escape_html" json:"escape_html"`
}

type ScriptTransform struct {
	Script      string            `yaml:"script" json:"script"`
	Environment map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
	Timeout     int               `yaml:"timeout" json:"timeout"` // in milliseconds
}

type RegexTransform struct {
	Pattern     string `yaml:"pattern" json:"pattern"`
	Replacement string `yaml:"replacement" json:"replacement"`
	Global      bool   `yaml:"global" json:"global"`
}

type MappingRule struct {
	Source    string `yaml:"source" json:"source"`
	Target    string `yaml:"target" json:"target"`
	Default   string `yaml:"default,omitempty" json:"default,omitempty"`
	Converter string `yaml:"converter,omitempty" json:"converter,omitempty"`
}

type AggregationConfig struct {
	Enabled     bool              `yaml:"enabled" json:"enabled"`
	Endpoints   []string          `yaml:"endpoints" json:"endpoints"`
	Timeout     int               `yaml:"timeout" json:"timeout"`
	MergeMethod string            `yaml:"merge_method" json:"merge_method"`
	ResultKey   string            `yaml:"result_key,omitempty" json:"result_key,omitempty"`
	Headers     map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
}

type ProtocolTransform struct {
	FromType string `yaml:"from_type" json:"from_type"` // json, xml, grpc, etc.
	ToType   string `yaml:"to_type" json:"to_type"`     // json, xml, grpc, etc.
	Mapping  string `yaml:"mapping,omitempty" json:"mapping,omitempty"`
}

type TransformContext struct {
	Request  *http.Request
	Response *http.Response
	Body     []byte
	Metadata map[string]interface{}
	Logger   logging.Logger
}

func NewAdvancedTransformer(config *AdvancedTransformConfig, logger logging.Logger) *AdvancedTransformer {
	return &AdvancedTransformer{
		config: config,
		logger: logger,
	}
}

func (t *AdvancedTransformer) TransformRequest(req *http.Request) error {
	if !t.shouldTransform(req.Header.Get("Content-Type")) {
		return nil
	}

	// Read request body
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("%w: failed to read request body: %v", ErrTransformFailed, err)
	}
	defer req.Body.Close()

	// Create transform context
	ctx := &TransformContext{
		Request:  req,
		Body:     body,
		Metadata: make(map[string]interface{}),
		Logger:   t.logger,
	}

	// Apply transformations
	transformedBody, err := t.applyTransformations(ctx)
	if err != nil {
		return err
	}

	// Replace request body
	req.Body = ioutil.NopCloser(bytes.NewReader(transformedBody))
	req.ContentLength = int64(len(transformedBody))

	// Update content-length header
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(transformedBody)))

	return nil
}

func (t *AdvancedTransformer) TransformResponse(resp *http.Response) error {
	if !t.shouldTransform(resp.Header.Get("Content-Type")) {
		return nil
	}

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: failed to read response body: %v", ErrTransformFailed, err)
	}
	defer resp.Body.Close()

	// Create transform context
	ctx := &TransformContext{
		Response: resp,
		Body:     body,
		Metadata: make(map[string]interface{}),
		Logger:   t.logger,
	}

	// Apply transformations
	transformedBody, err := t.applyTransformations(ctx)
	if err != nil {
		return err
	}

	// Replace response body
	resp.Body = ioutil.NopCloser(bytes.NewReader(transformedBody))
	resp.ContentLength = int64(len(transformedBody))

	// Update content-length header
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(transformedBody)))

	return nil
}

func (t *AdvancedTransformer) shouldTransform(contentType string) bool {
	if len(t.config.ContentTypes) == 0 {
		return true
	}

	for _, ct := range t.config.ContentTypes {
		if strings.Contains(contentType, ct) {
			return true
		}
	}

	return false
}

func (t *AdvancedTransformer) applyTransformations(ctx *TransformContext) ([]byte, error) {
	var err error
	body := ctx.Body

	switch t.config.Type {
	case "jsonpath":
		body, err = t.applyJSONPathTransformations(body, t.config.JSONPath)
	case "xpath":
		body, err = t.applyXPathTransformations(body, t.config.XPath)
	case "template":
		body, err = t.applyTemplateTransformation(body, ctx, t.config.Template)
	case "lua_script":
		body, err = t.applyLuaScriptTransformation(body, ctx, t.config.LuaScript)
	case "regex":
		body, err = t.applyRegexTransformation(body, t.config.RegexReplace)
	case "request_mapping":
		body, err = t.applyRequestMapping(body, ctx, t.config.RequestMapping)
	case "protocol":
		body, err = t.applyProtocolTransformation(body, t.config.Protocol)
	case "aggregation":
		if t.config.Aggregation.Enabled {
			body, err = t.applyAggregation(body, ctx, t.config.Aggregation)
		}
	default:
		t.logger.Warn("Unknown transformation type", "type", t.config.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransformFailed, err)
	}

	return body, nil
}

func (t *AdvancedTransformer) applyJSONPathTransformations(body []byte, config JSONPathTransform) ([]byte, error) {
	if len(body) == 0 || len(config.Operations) == 0 {
		return body, nil
	}

	parsed, err := gabs.ParseJSON(body)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing JSON: %v", ErrInvalidTransformInput, err)
	}

	for _, op := range config.Operations {
		switch op.Operation {
		case "add", "replace":
			var value interface{} = op.Value
			if op.ValueFrom != "" {
				// Extract value from another path
				extracted, err := jsonpath.Get(op.ValueFrom, parsed.Data())
				if err == nil {
					value = extracted
				}
			}
			_, err = parsed.SetP(value, op.Path)
			if err != nil {
				t.logger.Warn("JSONPath set operation failed", "path", op.Path, "error", err)
			}
		case "remove":
			err = parsed.DeleteP(op.Path)
			if err != nil {
				t.logger.Warn("JSONPath delete operation failed", "path", op.Path, "error", err)
			}
		case "append":
			// Find the array to append to
			arrayData := parsed.Path(op.Path).Data()
			if err != nil || arrayData == nil {
				t.logger.Warn("JSONPath append failed - path not found", "path", op.Path)
				continue
			}

			// Ensure it's an array
			arraySlice, ok := arrayData.([]interface{})
			if !ok {
				t.logger.Warn("JSONPath append failed - not an array", "path", op.Path)
				continue
			}

			// Append the value
			arraySlice = append(arraySlice, op.Value)
			_, err = parsed.SetP(arraySlice, op.Path)
			if err != nil {
				t.logger.Warn("JSONPath append operation failed", "path", op.Path, "error", err)
			}
		}
	}

	return parsed.Bytes(), nil
}

func (t *AdvancedTransformer) applyXPathTransformations(body []byte, config XPathTransform) ([]byte, error) {
	if len(body) == 0 || len(config.Operations) == 0 {
		return body, nil
	}

	// Parse XML
	doc, err := xmlquery.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing XML: %v", ErrInvalidTransformInput, err)
	}

	for _, op := range config.Operations {
		// Find nodes matching the XPath
		nodes, err := xmlquery.QueryAll(doc, op.Path)
		if err != nil {
			t.logger.Warn("XPath query failed", "path", op.Path, "error", err)
			continue
		}

		if len(nodes) == 0 {
			t.logger.Warn("XPath query returned no nodes", "path", op.Path)
			continue
		}

		switch op.Operation {
		case "replace":
			for _, node := range nodes {
				if node.Type == xmlquery.ElementNode {
					// Clear existing children
					node.FirstChild = nil

					// Add new text node with the value
					valueStr := fmt.Sprintf("%v", op.Value)
					textNode := &xmlquery.Node{
						Type: xmlquery.TextNode,
						Data: valueStr,
					}

					// Add as child node
					if node.FirstChild == nil {
						node.FirstChild = textNode
					} else {
						// Find the last child
						lastChild := node.FirstChild
						for lastChild.NextSibling != nil {
							lastChild = lastChild.NextSibling
						}
						lastChild.NextSibling = textNode
					}
				}
			}
		case "remove":
			for _, node := range nodes {
				if node.Parent != nil {
					xmlquery.RemoveFromTree(node)
				}
			}
		case "setAttribute":
			valueStr := fmt.Sprintf("%v", op.Value)
			parts := strings.SplitN(op.Path, "/@", 2)
			if len(parts) != 2 {
				t.logger.Warn("Invalid attribute path", "path", op.Path)
				continue
			}

			// Find elements
			elements, err := xmlquery.QueryAll(doc, parts[0])
			if err != nil || len(elements) == 0 {
				t.logger.Warn("XPath query for setAttribute failed", "path", parts[0])
				continue
			}

			// Set attribute on each element
			attrName := parts[1]
			for _, elem := range elements {
				if elem.Type == xmlquery.ElementNode {
					// Initialize attributes slice if nil
					if elem.Attr == nil {
						elem.Attr = make([]xmlquery.Attr, 0)
					}

					// Check if attribute already exists
					found := false
					for i, attr := range elem.Attr {
						if attr.Name.Local == attrName {
							elem.Attr[i].Value = valueStr
							found = true
							break
						}
					}

					// Add new attribute if not found
					if !found {
						elem.Attr = append(elem.Attr, xmlquery.Attr{
							Name:  xml.Name{Local: attrName},
							Value: valueStr,
						})
					}
				}
			}
		}
	}

	// Convert back to XML bytes
	var buf bytes.Buffer
	err = writeXML(doc, &buf)
	if err != nil {
		return nil, fmt.Errorf("%w: error writing XML: %v", ErrTransformFailed, err)
	}

	return buf.Bytes(), nil
}

// Helper function to write XML document to buffer
func writeXML(doc *xmlquery.Node, w *bytes.Buffer) error {
	var traverse func(node *xmlquery.Node, level int) error
	traverse = func(node *xmlquery.Node, level int) error {
		switch node.Type {
		case xmlquery.DocumentNode:
			w.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
			if node.FirstChild != nil {
				return traverse(node.FirstChild, level)
			}
		case xmlquery.ElementNode:
			w.WriteString("<")
			w.WriteString(node.Data)
			for _, attr := range node.Attr {
				w.WriteString(" ")
				w.WriteString(attr.Name.Local)
				w.WriteString(`="`)
				w.WriteString(attr.Value)
				w.WriteString(`"`)
			}
			if node.FirstChild == nil {
				w.WriteString("/>")
			} else {
				w.WriteString(">")
				for child := node.FirstChild; child != nil; child = child.NextSibling {
					if err := traverse(child, level+1); err != nil {
						return err
					}
				}
				w.WriteString("</")
				w.WriteString(node.Data)
				w.WriteString(">")
			}
		case xmlquery.TextNode:
			w.WriteString(node.Data)
		case xmlquery.CommentNode:
			w.WriteString("<!--")
			w.WriteString(node.Data)
			w.WriteString("-->")
		}

		// Handle siblings at the top level
		if level == 0 && node.NextSibling != nil {
			return traverse(node.NextSibling, level)
		}

		return nil
	}

	return traverse(doc, 0)
}

func (t *AdvancedTransformer) applyTemplateTransformation(body []byte, ctx *TransformContext, config TemplateTransform) ([]byte, error) {
	if len(config.Template) == 0 {
		return body, nil
	}

	// Create template data
	data := make(map[string]interface{})

	// Add original body
	data["Body"] = string(body)

	// Parse body as JSON/XML if possible
	var parsedBody interface{}
	if json.Unmarshal(body, &parsedBody) == nil {
		data["JSON"] = parsedBody
	} else {
		// Try as XML
		mv, err := mxj.NewMapXml(body)
		if err == nil {
			data["XML"] = mv.Old()
		}
	}

	// Add request/response data
	if ctx.Request != nil {
		reqData := map[string]interface{}{
			"Method": ctx.Request.Method,
			"Path":   ctx.Request.URL.Path,
			"Query":  ctx.Request.URL.Query(),
		}

		// Add headers
		headers := make(map[string]string)
		for name, values := range ctx.Request.Header {
			if len(values) > 0 {
				headers[name] = values[0]
			}
		}
		reqData["Headers"] = headers

		data["Request"] = reqData
	}

	if ctx.Response != nil {
		respData := map[string]interface{}{
			"StatusCode": ctx.Response.StatusCode,
		}

		// Add headers
		headers := make(map[string]string)
		for name, values := range ctx.Response.Header {
			if len(values) > 0 {
				headers[name] = values[0]
			}
		}
		respData["Headers"] = headers

		data["Response"] = respData
	}

	// Add metadata
	data["Metadata"] = ctx.Metadata

	// Add timestamp
	data["Timestamp"] = time.Now().Format(time.RFC3339)

	// Select template engine
	var result bytes.Buffer
	switch config.Engine {
	case "go", "golang", "":
		// Use Go's template engine
		tmpl, err := template.New("transform").
			Funcs(GetTemplateFuncs()).
			Option("missingkey=zero").
			Parse(config.Template)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrTransformTemplateInvalid, err)
		}

		err = tmpl.Execute(&result, data)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrTransformFailed, err)
		}

	// Add more template engines as needed
	default:
		return nil, fmt.Errorf("%w: unknown template engine: %s", ErrTransformTemplateInvalid, config.Engine)
	}

	return result.Bytes(), nil
}

func (t *AdvancedTransformer) applyLuaScriptTransformation(body []byte, ctx *TransformContext, config ScriptTransform) ([]byte, error) {
	if len(config.Script) == 0 {
		return body, nil
	}

	// Create Lua state
	L := lua.NewState()
	defer L.Close()

	// Set timeout (if available in your version of gopher-lua)
	// Note: SetExecutionLimit may not be available in all versions
	// If not available, consider using a context with timeout instead

	// Register JSON module
	luajson.Preload(L)

	// Create input table
	input := L.NewTable()

	// Add body
	L.SetField(input, "body", lua.LString(string(body)))

	// Try to parse body as JSON
	var parsedBody interface{}
	if json.Unmarshal(body, &parsedBody) == nil {
		// Use lua.LValue directly instead of converting from Go types
		// We'll construct the Lua table manually
		jsonTable := L.NewTable()

		// Convert parsedBody to Lua table structure
		// This is a simplified version - in practice you might need a recursive function
		if mapData, ok := parsedBody.(map[string]interface{}); ok {
			for k, v := range mapData {
				switch vt := v.(type) {
				case string:
					L.SetField(jsonTable, k, lua.LString(vt))
				case float64:
					L.SetField(jsonTable, k, lua.LNumber(vt))
				case bool:
					L.SetField(jsonTable, k, lua.LBool(vt))
				default:
					// For complex types, you might need more handling
					L.SetField(jsonTable, k, lua.LString(fmt.Sprintf("%v", vt)))
				}
			}
		}

		L.SetField(input, "json", jsonTable)
	}

	// Add request data if available
	if ctx.Request != nil {
		reqTable := L.NewTable()
		L.SetField(reqTable, "method", lua.LString(ctx.Request.Method))
		L.SetField(reqTable, "path", lua.LString(ctx.Request.URL.Path))

		// Add headers
		headersTable := L.NewTable()
		for name, values := range ctx.Request.Header {
			if len(values) > 0 {
				L.SetField(headersTable, name, lua.LString(values[0]))
			}
		}
		L.SetField(reqTable, "headers", headersTable)

		// Add query params
		queryTable := L.NewTable()
		for name, values := range ctx.Request.URL.Query() {
			if len(values) > 0 {
				L.SetField(queryTable, name, lua.LString(values[0]))
			}
		}
		L.SetField(reqTable, "query", queryTable)

		L.SetField(input, "request", reqTable)
	}

	// Add response data if available
	if ctx.Response != nil {
		respTable := L.NewTable()
		L.SetField(respTable, "status_code", lua.LNumber(ctx.Response.StatusCode))

		// Add headers
		headersTable := L.NewTable()
		for name, values := range ctx.Response.Header {
			if len(values) > 0 {
				L.SetField(headersTable, name, lua.LString(values[0]))
			}
		}
		L.SetField(respTable, "headers", headersTable)

		L.SetField(input, "response", respTable)
	}

	// Add environment variables
	envTable := L.NewTable()
	for name, value := range config.Environment {
		L.SetField(envTable, name, lua.LString(value))
	}
	L.SetField(input, "env", envTable)

	// Set input as global
	L.SetGlobal("input", input)

	// Create output table
	output := L.NewTable()
	L.SetGlobal("output", output)

	// Execute script
	err := L.DoString(config.Script)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrScriptExecutionFailed, err)
	}

	// Get result
	result := L.GetGlobal("output")
	resultTable, ok := result.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("%w: script did not set output table", ErrScriptExecutionFailed)
	}

	// Get body from output
	bodyValue := resultTable.RawGetString("body")
	if bodyValue == lua.LNil {
		// Try to get JSON from output and encode it
		jsonValue := resultTable.RawGetString("json")
		if jsonValue != lua.LNil {
			// Convert Lua table to Go structure
			var goValue interface{}
			err := luaTableToGoValue(jsonValue.(*lua.LTable), &goValue)
			if err != nil {
				return nil, fmt.Errorf("%w: failed to convert Lua table to Go value: %v",
					ErrScriptExecutionFailed, err)
			}

			// Convert to JSON bytes
			jsonBytes, err := json.Marshal(goValue)
			if err != nil {
				return nil, fmt.Errorf("%w: failed to marshal JSON output: %v",
					ErrScriptExecutionFailed, err)
			}

			return jsonBytes, nil
		}

		// If no body or JSON, return original body
		return body, nil
	}

	return []byte(bodyValue.String()), nil
}

// Helper function to convert Lua table to Go value
func luaTableToGoValue(table *lua.LTable, result *interface{}) error {
	goMap := make(map[string]interface{})
	goSlice := make([]interface{}, 0, table.Len())

	isArray := true
	maxN := 0

	table.ForEach(func(k, v lua.LValue) {
		// Check if we're dealing with an array
		if n, ok := k.(lua.LNumber); ok {
			index := int(n)
			if index > 0 {
				if index > maxN {
					maxN = index
				}
			} else {
				isArray = false
			}
		} else {
			isArray = false
		}

		var goValue interface{}

		switch v.Type() {
		case lua.LTNil:
			goValue = nil
		case lua.LTBool:
			goValue = bool(v.(lua.LBool))
		case lua.LTNumber:
			num := float64(v.(lua.LNumber))
			// Check if it's an integer
			if num == float64(int(num)) {
				goValue = int(num)
			} else {
				goValue = num
			}
		case lua.LTString:
			goValue = string(v.(lua.LString))
		case lua.LTTable:
			// Recursively convert nested tables
			var nestedValue interface{}
			if err := luaTableToGoValue(v.(*lua.LTable), &nestedValue); err != nil {
				return
			}
			goValue = nestedValue
		default:
			// Just use string representation for other types
			goValue = v.String()
		}

		if isArray {
			index := int(k.(lua.LNumber)) - 1 // Lua is 1-indexed
			if len(goSlice) <= index {
				// Expand the slice if needed
				newSlice := make([]interface{}, index+1)
				copy(newSlice, goSlice)
				goSlice = newSlice
			}
			goSlice[index] = goValue
		} else if kstr, ok := k.(lua.LString); ok {
			goMap[string(kstr)] = goValue
		}
	})

	if isArray && len(goMap) == 0 && maxN > 0 {
		*result = goSlice
	} else {
		*result = goMap
	}

	return nil
}

func (t *AdvancedTransformer) applyRegexTransformation(body []byte, config RegexTransform) ([]byte, error) {
	if len(config.Pattern) == 0 {
		return body, nil
	}

	re, err := regexp.Compile(config.Pattern)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid regex pattern: %v", ErrTransformFailed, err)
	}

	if config.Global {
		result := re.ReplaceAllString(string(body), config.Replacement)
		return []byte(result), nil
	}

	// Replace first occurrence only
	result := re.ReplaceAllStringFunc(string(body), func(match string) string {
		// Only replace the first occurrence, then return subsequent matches unchanged
		result := re.ReplaceAllString(match, config.Replacement)
		re = regexp.MustCompile("a^") // This will never match anything
		return result
	})

	return []byte(result), nil
}

func (t *AdvancedTransformer) applyRequestMapping(body []byte, ctx *TransformContext, mapping []MappingRule) ([]byte, error) {
	if len(mapping) == 0 {
		return body, nil
	}

	var sourceData map[string]interface{}
	err := json.Unmarshal(body, &sourceData)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing JSON for mapping: %v", ErrInvalidTransformInput, err)
	}

	// Create JSON query from source data
	jq := jsonq.NewQuery(sourceData)

	// Create target data
	targetData := make(map[string]interface{})

	for _, rule := range mapping {
		// Extract value from source using JSON path
		value, err := jq.Interface(strings.Split(rule.Source, ".")...)
		if err != nil || value == nil {
			// Use default value if provided
			if rule.Default != "" {
				value = rule.Default
			} else {
				continue // Skip this mapping
			}
		}

		// Apply converter if specified
		if rule.Converter != "" {
			value = applyConverter(value, rule.Converter)
		}

		// Set value in target data
		parts := strings.Split(rule.Target, ".")
		setNestedValue(targetData, parts, value)
	}

	// Convert target data to JSON
	result, err := json.Marshal(targetData)
	if err != nil {
		return nil, fmt.Errorf("%w: error marshaling mapped data: %v", ErrTransformFailed, err)
	}

	return result, nil
}

func (t *AdvancedTransformer) applyProtocolTransformation(body []byte, config ProtocolTransform) ([]byte, error) {
	if config.FromType == "" || config.ToType == "" || config.FromType == config.ToType {
		return body, nil
	}

	// JSON to XML
	if config.FromType == "json" && config.ToType == "xml" {
		var data interface{}
		err := json.Unmarshal(body, &data)
		if err != nil {
			return nil, fmt.Errorf("%w: error parsing JSON: %v", ErrInvalidTransformInput, err)
		}

		// Convert to XML
		mv := mxj.Map(data.(map[string]interface{}))
		xmlData, err := mv.Xml()
		if err != nil {
			return nil, fmt.Errorf("%w: error converting to XML: %v", ErrTransformFailed, err)
		}

		return xmlData, nil
	}

	// XML to JSON
	if config.FromType == "xml" && config.ToType == "json" {
		mv, err := mxj.NewMapXml(body)
		if err != nil {
			return nil, fmt.Errorf("%w: error parsing XML: %v", ErrInvalidTransformInput, err)
		}

		// Convert to JSON
		jsonData, err := json.Marshal(mv.Old())
		if err != nil {
			return nil, fmt.Errorf("%w: error converting to JSON: %v", ErrTransformFailed, err)
		}

		return jsonData, nil
	}

	// Add more protocol transformations as needed

	return nil, fmt.Errorf("%w: unsupported protocol transformation: %s to %s", ErrTransformFailed, config.FromType, config.ToType)
}

func (t *AdvancedTransformer) applyAggregation(body []byte, ctx *TransformContext, config AggregationConfig) ([]byte, error) {
	if !config.Enabled || len(config.Endpoints) == 0 {
		return body, nil
	}

	// Use the original request body as the template for other requests
	templateBody := body

	// Create HTTP client
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Millisecond,
	}

	// Create a channel for results
	resultCh := make(chan map[string]interface{}, len(config.Endpoints))
	errorCh := make(chan error, len(config.Endpoints))

	// Create context with timeout
	httpCtx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Millisecond)
	defer cancel()

	// Make requests to all endpoints
	for _, endpoint := range config.Endpoints {
		go func(endpoint string) {
			// Create request
			req, err := http.NewRequestWithContext(httpCtx, "POST", endpoint, bytes.NewReader(templateBody))
			if err != nil {
				errorCh <- fmt.Errorf("error creating request to %s: %v", endpoint, err)
				return
			}

			// Copy headers from original request
			if ctx.Request != nil {
				for name, values := range ctx.Request.Header {
					for _, value := range values {
						req.Header.Add(name, value)
					}
				}
			}

			// Add custom headers
			for name, value := range config.Headers {
				req.Header.Set(name, value)
			}

			// Send request
			resp, err := client.Do(req)
			if err != nil {
				errorCh <- fmt.Errorf("error sending request to %s: %v", endpoint, err)
				return
			}
			defer resp.Body.Close()

			// Read response body
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				errorCh <- fmt.Errorf("error reading response from %s: %v", endpoint, err)
				return
			}

			// Parse response as JSON
			var respData map[string]interface{}
			err = json.Unmarshal(respBody, &respData)
			if err != nil {
				errorCh <- fmt.Errorf("error parsing response from %s: %v", endpoint, err)
				return
			}

			// Send result
			resultCh <- respData
		}(endpoint)
	}

	// Collect results
	var results []map[string]interface{}
	for i := 0; i < len(config.Endpoints); i++ {
		select {
		case result := <-resultCh:
			results = append(results, result)
		case err := <-errorCh:
			t.logger.Error("Error in aggregation request", "error", err)
		case <-httpCtx.Done():
			t.logger.Error("Aggregation timed out")
			return nil, fmt.Errorf("%w: aggregation timed out", ErrTransformFailed)
		}
	}

	// Merge results
	mergedResult := make(map[string]interface{})

	switch config.MergeMethod {
	case "append":
		// Append all results in an array
		mergedResult[config.ResultKey] = results
	case "merge":
		// Merge all results into a single object
		for i, result := range results {
			for k, v := range result {
				key := k
				if len(results) > 1 {
					key = fmt.Sprintf("%s_%d", k, i+1)
				}
				mergedResult[key] = v
			}
		}
	case "combine":
		// Combine matching keys
		for _, result := range results {
			for k, v := range result {
				if existing, ok := mergedResult[k]; ok {
					// If the existing value is an array, append to it
					if arr, ok := existing.([]interface{}); ok {
						mergedResult[k] = append(arr, v)
					} else {
						// Convert to array
						mergedResult[k] = []interface{}{existing, v}
					}
				} else {
					mergedResult[k] = v
				}
			}
		}
	default:
		// Default: use the first result
		if len(results) > 0 {
			mergedResult = results[0]
		}
	}

	// Convert merged result to JSON
	result, err := json.Marshal(mergedResult)
	if err != nil {
		return nil, fmt.Errorf("%w: error marshaling aggregated data: %v", ErrTransformFailed, err)
	}

	return result, nil
}

func setNestedValue(data map[string]interface{}, path []string, value interface{}) {
	if len(path) == 1 {
		data[path[0]] = value
		return
	}

	if _, ok := data[path[0]]; !ok {
		data[path[0]] = make(map[string]interface{})
	}

	if nestedMap, ok := data[path[0]].(map[string]interface{}); ok {
		setNestedValue(nestedMap, path[1:], value)
	}
}

func applyConverter(value interface{}, converter string) interface{} {
	switch converter {
	case "toString":
		return fmt.Sprintf("%v", value)
	case "toInt":
		if strVal, ok := value.(string); ok {
			if intVal, err := strconv.Atoi(strVal); err == nil {
				return intVal
			}
		} else if floatVal, ok := value.(float64); ok {
			return int(floatVal)
		}
	case "toFloat":
		if strVal, ok := value.(string); ok {
			if floatVal, err := strconv.ParseFloat(strVal, 64); err == nil {
				return floatVal
			}
		} else if intVal, ok := value.(int); ok {
			return float64(intVal)
		}
	case "toBool":
		if strVal, ok := value.(string); ok {
			return strVal == "true" || strVal == "yes" || strVal == "1"
		} else if intVal, ok := value.(int); ok {
			return intVal != 0
		}
	case "toLowerCase":
		if strVal, ok := value.(string); ok {
			return strings.ToLower(strVal)
		}
	case "toUpperCase":
		if strVal, ok := value.(string); ok {
			return strings.ToUpper(strVal)
		}
	}

	return value
}
