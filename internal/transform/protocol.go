package transform

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

type ProtocolTransformer struct {
	config   *ProtocolTransformConfig
	logger   logging.Logger
	fileDesc *desc.FileDescriptor
	msgDesc  *desc.MessageDescriptor
}

type ProtocolTransformConfig struct {
	SourceType    string `yaml:"source_type" json:"source_type"` // json, xml, grpc, etc.
	TargetType    string `yaml:"target_type" json:"target_type"` // json, xml, grpc, etc.
	ProtoFile     string `yaml:"proto_file,omitempty" json:"proto_file,omitempty"`
	MessageType   string `yaml:"message_type,omitempty" json:"message_type,omitempty"`
	Mapping       string `yaml:"mapping,omitempty" json:"mapping,omitempty"`
	PreserveNulls bool   `yaml:"preserve_nulls" json:"preserve_nulls"`
}

func NewProtocolTransformer(config *ProtocolTransformConfig, logger logging.Logger) (*ProtocolTransformer, error) {
	t := &ProtocolTransformer{
		config: config,
		logger: logger,
	}

	if config.SourceType == "grpc" || config.TargetType == "grpc" {
		if config.ProtoFile == "" || config.MessageType == "" {
			return nil, fmt.Errorf("proto_file and message_type are required for gRPC transformations")
		}

		// Load proto definitions if needed
		if err := t.loadProtoDefinitions(); err != nil {
			return nil, err
		}
	}

	return t, nil
}

func (t *ProtocolTransformer) Transform(input []byte, headers http.Header) ([]byte, error) {
	if len(input) == 0 {
		return input, nil
	}

	// Determine content type from headers if not specified
	contentType := ""
	if ct := headers.Get("Content-Type"); ct != "" {
		contentType = ct
	}

	// Convert from source to intermediate format (typically JSON)
	var intermediate interface{}
	var err error

	switch t.config.SourceType {
	case "json":
		err = json.Unmarshal(input, &intermediate)
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON input: %w", err)
		}

	case "xml":
		intermediate, err = t.xmlToMap(input)
		if err != nil {
			return nil, fmt.Errorf("error parsing XML input: %w", err)
		}

	case "grpc":
		if t.msgDesc == nil {
			return nil, fmt.Errorf("protobuf message descriptor not loaded")
		}

		// Create dynamic message
		msg := dynamic.NewMessage(t.msgDesc)

		// Unmarshal binary protobuf
		if err := msg.Unmarshal(input); err != nil {
			return nil, fmt.Errorf("error parsing protobuf input: %w", err)
		}

		// Convert to JSON
		marshaler := jsonpb.Marshaler{
			EmitDefaults: true,
			OrigName:     true,
		}

		jsonStr, err := marshaler.MarshalToString(msg)
		if err != nil {
			return nil, fmt.Errorf("error converting protobuf to JSON: %w", err)
		}

		err = json.Unmarshal([]byte(jsonStr), &intermediate)
		if err != nil {
			return nil, fmt.Errorf("error parsing intermediate JSON: %w", err)
		}
	}

	// Apply custom mapping if provided
	if t.config.Mapping != "" {
		intermediate, err = t.applyMapping(intermediate)
		if err != nil {
			return nil, fmt.Errorf("error applying mapping: %w", err)
		}
	}

	// Convert from intermediate format to target format
	var output []byte

	switch t.config.TargetType {
	case "json":
		output, err = json.Marshal(intermediate)
		if err != nil {
			return nil, fmt.Errorf("error generating JSON output: %w", err)
		}

	case "xml":
		output, err = t.mapToXML(intermediate)
		if err != nil {
			return nil, fmt.Errorf("error generating XML output: %w", err)
		}

	case "grpc":
		if t.msgDesc == nil {
			return nil, fmt.Errorf("protobuf message descriptor not loaded")
		}

		// Create dynamic message
		msg := dynamic.NewMessage(t.msgDesc)

		// Convert intermediate to JSON
		jsonData, err := json.Marshal(intermediate)
		if err != nil {
			return nil, fmt.Errorf("error marshaling to JSON: %w", err)
		}

		// Parse JSON into protobuf message
		if err := jsonpb.UnmarshalString(string(jsonData), msg); err != nil {
			return nil, fmt.Errorf("error parsing JSON to protobuf: %w", err)
		}

		// Marshal to binary protobuf
		output, err = msg.Marshal()
		if err != nil {
			return nil, fmt.Errorf("error generating protobuf output: %w", err)
		}
	}

	return output, nil
}

func (t *ProtocolTransformer) loadProtoDefinitions() error {
	// This is a simplified version - in a real implementation,
	// you would use the protoreflect package to load proto files
	// and find message types
	return nil
}

func (t *ProtocolTransformer) xmlToMap(input []byte) (map[string]interface{}, error) {
	// Parse XML to map
	var result map[string]interface{}

	// Create a decoder with the XML data
	decoder := xml.NewDecoder(bytes.NewReader(input))
	result = make(map[string]interface{})

	var current map[string]interface{} = result
	var stack []map[string]interface{}

	for {
		token, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			// Create a new map for this element
			newMap := make(map[string]interface{})

			// Handle attributes
			if len(t.Attr) > 0 {
				attrs := make(map[string]string)
				for _, attr := range t.Attr {
					attrs[attr.Name.Local] = attr.Value
				}
				newMap["@attributes"] = attrs
			}

			// Save the current map to the stack
			stack = append(stack, current)

			// Add the new map to the current one
			name := t.Name.Local
			if existing, ok := current[name]; ok {
				// If this element already exists, make it an array
				switch existingVal := existing.(type) {
				case []interface{}:
					current[name] = append(existingVal, newMap)
				default:
					current[name] = []interface{}{existingVal, newMap}
				}
			} else {
				current[name] = newMap
			}

			// Make the new map the current one
			current = newMap

		case xml.EndElement:
			// Go back to the parent map
			if len(stack) > 0 {
				current = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
			}

		case xml.CharData:
			// Add character data to the current map
			text := string(t)
			text = strings.TrimSpace(text)
			if text != "" {
				current["#text"] = text
			}
		}
	}

	return result, nil
}

func (t *ProtocolTransformer) mapToXML(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")

	// Convert to map if it's not already
	m, ok := data.(map[string]interface{})
	if !ok {
		// If it's not a map, wrap it in a root element
		// We need to choose a name for the root element
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(jsonBytes, &m)
		if err != nil {
			return nil, err
		}
		m = map[string]interface{}{"root": data}
	}

	// Find the root element
	if len(m) != 1 {
		// If there are multiple root elements, wrap them in a single root
		buf.WriteString("<root>\n")
		if err := t.writeXMLMap(&buf, m, "  "); err != nil {
			return nil, err
		}
		buf.WriteString("</root>")
	} else {
		// If there's just one root element, use it
		for key, val := range m {
			if err := t.writeXMLElement(&buf, key, val, ""); err != nil {
				return nil, err
			}
		}
	}

	return buf.Bytes(), nil
}

func (t *ProtocolTransformer) writeXMLMap(buf *bytes.Buffer, m map[string]interface{}, indent string) error {
	for key, val := range m {
		if err := t.writeXMLElement(buf, key, val, indent); err != nil {
			return err
		}
	}
	return nil
}

func (t *ProtocolTransformer) writeXMLElement(buf *bytes.Buffer, name string, val interface{}, indent string) error {
	if name == "@attributes" || name == "#text" {
		return nil // These are handled specially
	}

	// Start tag
	buf.WriteString(indent)
	buf.WriteString("<")
	buf.WriteString(name)

	// Write attributes if present
	if m, ok := val.(map[string]interface{}); ok {
		if attrs, ok := m["@attributes"].(map[string]string); ok {
			for attrName, attrVal := range attrs {
				buf.WriteString(" ")
				buf.WriteString(attrName)
				buf.WriteString("=\"")
				xml.EscapeText(buf, []byte(attrVal))
				buf.WriteString("\"")
			}
		}
	}

	buf.WriteString(">")

	// Write value
	switch v := val.(type) {
	case map[string]interface{}:
		// Check if there's text content
		if text, ok := v["#text"]; ok {
			xml.EscapeText(buf, []byte(fmt.Sprintf("%v", text)))
		} else if len(v) == 0 || (len(v) == 1 && v["@attributes"] != nil) {
			// Empty element or just attributes
		} else {
			// Complex element with children
			buf.WriteString("\n")
			for childName, childVal := range v {
				if childName != "@attributes" && childName != "#text" {
					t.writeXMLElement(buf, childName, childVal, indent+"  ")
				}
			}
			buf.WriteString(indent)
		}
	case []interface{}:
		// Array of elements - we need to repeat the tag
		buf.WriteString("\n")
		for _, item := range v {
			switch item.(type) {
			case map[string]interface{}:
				if err := t.writeXMLMap(buf, item.(map[string]interface{}), indent+"  "); err != nil {
					return err
				}
			default:
				buf.WriteString(indent + "  ")
				xml.EscapeText(buf, []byte(fmt.Sprintf("%v", item)))
				buf.WriteString("\n")
			}
		}
		buf.WriteString(indent)
	default:
		// Simple value
		xml.EscapeText(buf, []byte(fmt.Sprintf("%v", v)))
	}

	// End tag
	buf.WriteString("</")
	buf.WriteString(name)
	buf.WriteString(">\n")

	return nil
}

func (t *ProtocolTransformer) applyMapping(data interface{}) (interface{}, error) {
	// Parse mapping template
	var mapping map[string]string
	if err := json.Unmarshal([]byte(t.config.Mapping), &mapping); err != nil {
		return nil, fmt.Errorf("invalid mapping JSON: %w", err)
	}

	// Apply mapping
	result := make(map[string]interface{})

	// Convert data to map if not already
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return data, nil // Can't apply mapping to non-map data
	}

	for targetPath, sourcePath := range mapping {
		// Get value from source path
		value, err := getValueByPath(dataMap, sourcePath)
		if err != nil {
			t.logger.Warn("Failed to get value from source path",
				"source_path", sourcePath,
				"error", err.Error())
			continue
		}

		// Set value in target path
		err = setValueByPath(result, targetPath, value)
		if err != nil {
			t.logger.Warn("Failed to set value in target path",
				"target_path", targetPath,
				"error", err.Error())
		}
	}

	return result, nil
}

func getValueByPath(data map[string]interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - return the value
			return current[part], nil
		}

		// Navigate to the next level
		next, ok := current[part]
		if !ok {
			return nil, fmt.Errorf("path element %s not found", part)
		}

		// Convert to map
		nextMap, ok := next.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("path element %s is not a map", part)
		}

		current = nextMap
	}

	return nil, fmt.Errorf("invalid path: %s", path)
}

func setValueByPath(data map[string]interface{}, path string, value interface{}) error {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - set the value
			current[part] = value
			return nil
		}

		// Create next level if needed
		next, ok := current[part]
		if !ok {
			next = make(map[string]interface{})
			current[part] = next
		}

		// Convert to map
		nextMap, ok := next.(map[string]interface{})
		if !ok {
			return fmt.Errorf("path element %s is not a map", part)
		}

		current = nextMap
	}

	return fmt.Errorf("invalid path: %s", path)
}
