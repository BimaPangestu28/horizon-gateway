package aggregation

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type Aggregator struct {
	client      *http.Client
	logger      logging.Logger
	maxBodySize int64
}

type AggregationConfig struct {
	Endpoints     []EndpointConfig  `yaml:"endpoints"`
	Timeout       time.Duration     `yaml:"timeout"`
	MergeMethod   string            `yaml:"merge_method"`
	ResultKey     string            `yaml:"result_key"`
	Headers       map[string]string `yaml:"headers"`
	MaxBodySize   int64             `yaml:"max_body_size"`
	ParallelCalls bool              `yaml:"parallel_calls"`
}

type EndpointConfig struct {
	Name         string            `yaml:"name"`
	URL          string            `yaml:"url"`
	Method       string            `yaml:"method"`
	Headers      map[string]string `yaml:"headers"`
	BodyTemplate string            `yaml:"body_template"`
	ResultKey    string            `yaml:"result_key"`
	TimeoutMs    int               `yaml:"timeout_ms"`
	Required     bool              `yaml:"required"`
}

type AggregationResult struct {
	Combined    map[string]interface{}
	StatusCode  int
	Headers     map[string][]string
	Errors      []error
	ResponseMap map[string]*EndpointResponse
}

type EndpointResponse struct {
	Name       string
	StatusCode int
	Body       []byte
	Headers    http.Header
	Duration   time.Duration
	Error      error
}

func NewAggregator(logger logging.Logger, timeout time.Duration, maxBodySize int64) *Aggregator {
	if maxBodySize <= 0 {
		maxBodySize = 10 * 1024 * 1024 // 10MB default
	}

	return &Aggregator{
		client: &http.Client{
			Timeout: timeout,
		},
		logger:      logger,
		maxBodySize: maxBodySize,
	}
}

func (a *Aggregator) Aggregate(ctx context.Context, config *AggregationConfig, templateData map[string]interface{}) (*AggregationResult, error) {
	if len(config.Endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints configured")
	}

	// Set up result
	result := &AggregationResult{
		Combined:    make(map[string]interface{}),
		StatusCode:  http.StatusOK,
		Headers:     make(map[string][]string),
		ResponseMap: make(map[string]*EndpointResponse),
	}

	// Set up context with timeout if configured
	var cancel context.CancelFunc
	if config.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	// Make requests to all endpoints
	if config.ParallelCalls {
		a.aggregateParallel(ctx, config, templateData, result)
	} else {
		a.aggregateSequential(ctx, config, templateData, result)
	}

	// Check for required endpoint failures
	for _, endpoint := range config.Endpoints {
		if endpoint.Required {
			resp := result.ResponseMap[endpoint.Name]
			if resp == nil || resp.Error != nil {
				result.StatusCode = http.StatusBadGateway
				result.Errors = append(result.Errors, fmt.Errorf("required endpoint %s failed", endpoint.Name))
				return result, nil
			}
		}
	}

	// Merge results according to configured method
	err := a.mergeResults(config, result)
	if err != nil {
		result.Errors = append(result.Errors, err)
		result.StatusCode = http.StatusInternalServerError
	}

	return result, nil
}

func (a *Aggregator) aggregateParallel(ctx context.Context, config *AggregationConfig, templateData map[string]interface{}, result *AggregationResult) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, endpoint := range config.Endpoints {
		wg.Add(1)
		go func(ep EndpointConfig) {
			defer wg.Done()

			resp, err := a.callEndpoint(ctx, ep, templateData, config.Headers)

			mu.Lock()
			result.ResponseMap[ep.Name] = resp
			if err != nil {
				result.Errors = append(result.Errors, err)
			}
			mu.Unlock()
		}(endpoint)
	}

	wg.Wait()
}
