package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

// SchemaRegistry manages the GraphQL schemas for the API Gateway
type SchemaRegistry struct {
	schemas      map[string]*SchemaInfo
	schemasMutex sync.RWMutex
	logger       logging.Logger
	httpClient   *http.Client
}

// SchemaInfo holds information about a GraphQL schema
type SchemaInfo struct {
	Name             string               `json:"name"`
	Version          string               `json:"version"`
	Schema           string               `json:"schema,omitempty"`
	Types            []TypeInfo           `json:"types,omitempty"`
	QueryType        string               `json:"queryType,omitempty"`
	MutationType     string               `json:"mutationType,omitempty"`
	SubscriptionType string               `json:"subscriptionType,omitempty"`
	Directives       []DirectiveInfo      `json:"directives,omitempty"`
	TypeMap          map[string]*TypeInfo `json:"-"`
	LastUpdated      time.Time            `json:"lastUpdated"`
	Upstream         string               `json:"upstream"`
}

// TypeInfo holds information about a GraphQL type
type TypeInfo struct {
	Kind          string          `json:"kind"`
	Name          string          `json:"name"`
	Description   string          `json:"description,omitempty"`
	Fields        []FieldInfo     `json:"fields,omitempty"`
	InputFields   []FieldInfo     `json:"inputFields,omitempty"`
	Interfaces    []InterfaceInfo `json:"interfaces,omitempty"`
	EnumValues    []EnumValueInfo `json:"enumValues,omitempty"`
	PossibleTypes []TypeRefInfo   `json:"possibleTypes,omitempty"`
}

// FieldInfo holds information about a GraphQL field
type FieldInfo struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Type        TypeRefInfo `json:"type"`
	Args        []ArgInfo   `json:"args,omitempty"`
	Deprecation string      `json:"deprecation,omitempty"`
}

// ArgInfo holds information about a GraphQL argument
type ArgInfo struct {
	Name         string      `json:"name"`
	Description  string      `json:"description,omitempty"`
	Type         TypeRefInfo `json:"type"`
	DefaultValue string      `json:"defaultValue,omitempty"`
}

// TypeRefInfo holds information about a GraphQL type reference
type TypeRefInfo struct {
	Kind   string       `json:"kind"`
	Name   string       `json:"name,omitempty"`
	OfType *TypeRefInfo `json:"ofType,omitempty"`
}

// EnumValueInfo holds information about a GraphQL enum value
type EnumValueInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Deprecation string `json:"deprecation,omitempty"`
}

// InterfaceInfo holds information about a GraphQL interface
type InterfaceInfo struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// DirectiveInfo holds information about a GraphQL directive
type DirectiveInfo struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Locations   []string  `json:"locations"`
	Args        []ArgInfo `json:"args,omitempty"`
}

// NewSchemaRegistry creates a new schema registry
func NewSchemaRegistry(logger logging.Logger) *SchemaRegistry {
	return &SchemaRegistry{
		schemas:    make(map[string]*SchemaInfo),
		logger:     logger,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchSchema fetches a GraphQL schema from an upstream service
func (r *SchemaRegistry) FetchSchema(ctx context.Context, name, upstream string) (*SchemaInfo, error) {
	introspectionQuery := `
	{
		__schema {
			queryType { name }
			mutationType { name }
			subscriptionType { name }
			types {
				kind
				name
				description
				fields(includeDeprecated: true) {
					name
					description
					args {
						name
						description
						type {
							kind
							name
							ofType {
								kind
								name
								ofType {
									kind
									name
									ofType {
										kind
										name
									}
								}
							}
						}
						defaultValue
					}
					type {
						kind
						name
						ofType {
							kind
							name
							ofType {
								kind
								name
								ofType {
									kind
									name
								}
							}
						}
					}
					isDeprecated
					deprecationReason
				}
				inputFields {
					name
					description
					type {
						kind
						name
						ofType {
							kind
							name
							ofType {
								kind
								name
								ofType {
									kind
									name
								}
							}
						}
					}
					defaultValue
				}
				interfaces {
					kind
					name
				}
				enumValues(includeDeprecated: true) {
					name
					description
					isDeprecated
					deprecationReason
				}
				possibleTypes {
					kind
					name
				}
			}
			directives {
				name
				description
				locations
				args {
					name
					description
					type {
						kind
						name
						ofType {
							kind
							name
							ofType {
								kind
								name
								ofType {
									kind
									name
								}
							}
						}
					}
					defaultValue
				}
			}
		}
	}`

	// Ensure the upstream URL has a trailing slash
	if !strings.HasSuffix(upstream, "/") {
		upstream += "/"
	}

	// If the URL doesn't have a /graphql path, add it
	if !strings.Contains(upstream, "/graphql") {
		upstream += "graphql"
	}

	// Prepare request
	reqBody := map[string]interface{}{
		"query": introspectionQuery,
	}
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling introspection query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", upstream, strings.NewReader(string(reqJSON)))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending introspection query: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-OK response from GraphQL server: %d %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var respData struct {
		Data struct {
			Schema struct {
				QueryType struct {
					Name string `json:"name"`
				} `json:"queryType"`
				MutationType struct {
					Name string `json:"name"`
				} `json:"mutationType"`
				SubscriptionType struct {
					Name string `json:"name"`
				} `json:"subscriptionType"`
				Types      []TypeInfo      `json:"types"`
				Directives []DirectiveInfo `json:"directives"`
			} `json:"__schema"`
		} `json:"data"`
		Errors []GraphQLError `json:"errors,omitempty"`
	}

	if err := json.Unmarshal(respBody, &respData); err != nil {
		return nil, fmt.Errorf("error parsing introspection query response: %w", err)
	}

	if len(respData.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL server returned errors: %v", respData.Errors)
	}

	// Create schema info
	schemaInfo := &SchemaInfo{
		Name:             name,
		Version:          time.Now().Format(time.RFC3339),
		Types:            respData.Data.Schema.Types,
		QueryType:        respData.Data.Schema.QueryType.Name,
		MutationType:     respData.Data.Schema.MutationType.Name,
		SubscriptionType: respData.Data.Schema.SubscriptionType.Name,
		Directives:       respData.Data.Schema.Directives,
		TypeMap:          make(map[string]*TypeInfo),
		LastUpdated:      time.Now(),
		Upstream:         upstream,
	}

	// Build type map
	for i := range schemaInfo.Types {
		typeInfo := &schemaInfo.Types[i]
		schemaInfo.TypeMap[typeInfo.Name] = typeInfo
	}

	// Store schema
	r.schemasMutex.Lock()
	r.schemas[name] = schemaInfo
	r.schemasMutex.Unlock()

	r.logger.Info("Fetched GraphQL schema",
		"name", name,
		"upstream", upstream,
		"types", len(schemaInfo.Types),
		"directives", len(schemaInfo.Directives))

	return schemaInfo, nil
}

// GetSchema retrieves a schema by name
func (r *SchemaRegistry) GetSchema(name string) (*SchemaInfo, bool) {
	r.schemasMutex.RLock()
	defer r.schemasMutex.RUnlock()

	schema, exists := r.schemas[name]
	return schema, exists
}

// ListSchemas lists all available schemas
func (r *SchemaRegistry) ListSchemas() []*SchemaInfo {
	r.schemasMutex.RLock()
	defer r.schemasMutex.RUnlock()

	schemas := make([]*SchemaInfo, 0, len(r.schemas))
	for _, schema := range r.schemas {
		schemas = append(schemas, schema)
	}

	return schemas
}

// ValidateQuery validates a GraphQL query against a schema
func (r *SchemaRegistry) ValidateQuery(schemaName, query string) []GraphQLError {
	// Get schema
	schema, exists := r.GetSchema(schemaName)
	if !exists {
		return []GraphQLError{
			{
				Message: fmt.Sprintf("Schema '%s' not found", schemaName),
				Extensions: map[string]interface{}{
					"code": "SCHEMA_NOT_FOUND",
				},
			},
		}
	}

	// For now, we're just implementing a dummy validator
	// A real validator would parse the query and validate it against the schema
	return []GraphQLError{}
}

// RemoveSchema removes a schema from the registry
func (r *SchemaRegistry) RemoveSchema(name string) bool {
	r.schemasMutex.Lock()
	defer r.schemasMutex.Unlock()

	_, exists := r.schemas[name]
	if exists {
		delete(r.schemas, name)
		return true
	}
	return false
}
