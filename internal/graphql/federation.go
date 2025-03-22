package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type FederationService struct {
	registry      *SchemaRegistry
	logger        logging.Logger
	services      map[string]*FederatedService
	servicesMutex sync.RWMutex
	client        *http.Client
}

type FederatedService struct {
	Name      string
	URL       string
	Schema    *SchemaInfo
	Namespace string
	Entities  []string
}

type FederationConfig struct {
	Services []struct {
		Name      string `yaml:"name"`
		URL       string `yaml:"url"`
		Namespace string `yaml:"namespace"`
	} `yaml:"services"`
	QueryTimeout time.Duration `yaml:"query_timeout"`
}

func NewFederationService(registry *SchemaRegistry, logger logging.Logger) *FederationService {
	return &FederationService{
		registry: registry,
		logger:   logger,
		services: make(map[string]*FederatedService),
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (f *FederationService) Initialize(ctx context.Context, config *FederationConfig) error {
	// Add each federated service
	for _, svc := range config.Services {
		if err := f.AddService(ctx, svc.Name, svc.URL, svc.Namespace); err != nil {
			f.logger.Error("Error adding federated service",
				"name", svc.Name,
				"url", svc.URL,
				"error", err)
			continue
		}
	}

	// Log federated services
	f.logger.Info("GraphQL federation initialized",
		"services", len(f.services))

	return nil
}

func (f *FederationService) AddService(ctx context.Context, name, url, namespace string) error {
	// Fetch schema from the service
	schema, err := f.registry.FetchSchema(ctx, name, url)
	if err != nil {
		return fmt.Errorf("failed to fetch schema for service %s: %w", name, err)
	}

	// Analyze schema to find entities
	entities := f.findEntities(schema)

	// Create federated service
	service := &FederatedService{
		Name:      name,
		URL:       url,
		Schema:    schema,
		Namespace: namespace,
		Entities:  entities,
	}

	// Store service
	f.servicesMutex.Lock()
	f.services[name] = service
	f.servicesMutex.Unlock()

	f.logger.Info("Added federated GraphQL service",
		"name", name,
		"url", url,
		"namespace", namespace,
		"entities", entities)

	return nil
}

func (f *FederationService) RemoveService(name string) bool {
	f.servicesMutex.Lock()
	defer f.servicesMutex.Unlock()

	_, exists := f.services[name]
	if exists {
		delete(f.services, name)
		f.registry.RemoveSchema(name)
		return true
	}
	return false
}

func (f *FederationService) GetService(name string) (*FederatedService, bool) {
	f.servicesMutex.RLock()
	defer f.servicesMutex.RUnlock()

	service, exists := f.services[name]
	return service, exists
}

func (f *FederationService) ListServices() []*FederatedService {
	f.servicesMutex.RLock()
	defer f.servicesMutex.RUnlock()

	services := make([]*FederatedService, 0, len(f.services))
	for _, service := range f.services {
		services = append(services, service)
	}

	return services
}

func (f *FederationService) ExecuteQuery(ctx context.Context, query *GraphQLRequest) (*GraphQLResponse, error) {
	// Parse the query to determine which services need to be queried
	servicesToQuery, err := f.analyzeQuery(query.Query)
	if err != nil {
		return nil, fmt.Errorf("error analyzing query: %w", err)
	}

	if len(servicesToQuery) == 0 {
		return nil, fmt.Errorf("no services matched the query")
	}

	// If only one service is needed, forward the query directly
	if len(servicesToQuery) == 1 {
		service, exists := f.GetService(servicesToQuery[0])
		if !exists {
			return nil, fmt.Errorf("service %s not found", servicesToQuery[0])
		}

		return f.queryService(ctx, service, query)
	}

	// For multiple services, we need to split the query and combine results
	// This is a simplified implementation, a real federation service would
	// need to handle complex scenarios like entity references between services

	// For now, simply execute the query against each service and merge results
	var combinedResponse GraphQLResponse
	var combinedErrors []GraphQLError
	combinedData := make(map[string]interface{})

	for _, serviceName := range servicesToQuery {
		service, exists := f.GetService(serviceName)
		if !exists {
			combinedErrors = append(combinedErrors, GraphQLError{
				Message: fmt.Sprintf("Service %s not found", serviceName),
				Extensions: map[string]interface{}{
					"code": "SERVICE_NOT_FOUND",
				},
			})
			continue
		}

		// Execute query against this service
		resp, err := f.queryService(ctx, service, query)
		if err != nil {
			combinedErrors = append(combinedErrors, GraphQLError{
				Message: fmt.Sprintf("Error querying service %s: %s", serviceName, err.Error()),
				Extensions: map[string]interface{}{
					"code":    "QUERY_ERROR",
					"service": serviceName,
				},
			})
			continue
		}

		// Add any errors from this service
		if resp.Errors != nil && len(resp.Errors) > 0 {
			for _, e := range resp.Errors {
				// Add service name to error extensions
				if e.Extensions == nil {
					e.Extensions = make(map[string]interface{})
				}
				e.Extensions["service"] = serviceName
				combinedErrors = append(combinedErrors, e)
			}
		}

		// Merge data from this service
		if resp.Data != nil {
			if data, ok := resp.Data.(map[string]interface{}); ok {
				for k, v := range data {
					// Use service namespace as prefix if configured
					key := k
					if service.Namespace != "" {
						key = service.Namespace + "_" + k
					}
					combinedData[key] = v
				}
			}
		}
	}

	// Build combined response
	combinedResponse.Data = combinedData
	if len(combinedErrors) > 0 {
		combinedResponse.Errors = combinedErrors
	}

	return &combinedResponse, nil
}

func (f *FederationService) queryService(ctx context.Context, service *FederatedService, query *GraphQLRequest) (*GraphQLResponse, error) {
	// Create request body
	requestBody, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("error serializing GraphQL query: %w", err)
	}

	// Create HTTP request
	url := service.URL
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	if !strings.Contains(url, "/graphql") {
		url += "graphql"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending query to service %s: %w", service.Name, err)
	}
	defer resp.Body.Close()

	// Parse response
	var graphqlResponse GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&graphqlResponse); err != nil {
		return nil, fmt.Errorf("error parsing GraphQL response: %w", err)
	}

	return &graphqlResponse, nil
}

func (f *FederationService) findEntities(schema *SchemaInfo) []string {
	// Look for types with @key directive or _Entity union type
	// This is a simplification; Federation has specific requirements
	var entities []string

	for _, t := range schema.Types {
		// Skip built-in types
		if strings.HasPrefix(t.Name, "__") {
			continue
		}

		// Check if type has @key directive (indicating an entity)
		// This is just a simple check; a real implementation would parse directives
		if t.Kind == "OBJECT" {
			// For now, just consider all object types as potential entities
			// In a real implementation, we'd check for @key directive
			entities = append(entities, t.Name)
		} else if t.Name == "_Entity" && t.Kind == "UNION" {
			// Federation uses _Entity union type for entity resolution
			for _, pt := range t.PossibleTypes {
				entities = append(entities, pt.Name)
			}
		}
	}

	return entities
}

func (f *FederationService) analyzeQuery(query string) ([]string, error) {
	// A proper implementation would parse the GraphQL query and determine
	// which services need to be queried based on types/fields used
	// For simplicity, we'll just return all services

	f.servicesMutex.RLock()
	defer f.servicesMutex.RUnlock()

	services := make([]string, 0, len(f.services))
	for name := range f.services {
		services = append(services, name)
	}

	return services, nil
}
