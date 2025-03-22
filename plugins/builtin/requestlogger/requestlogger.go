package requestlogger

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/bimapangestu28/horizon/plugins/interfaces"
)

type RequestLoggerPlugin struct {
	logger       logging.Logger
	config       Config
	initialized  bool
	requestCount int
}

type Config struct {
	LogHeaders     bool     `json:"log_headers"`
	LogBody        bool     `json:"log_body"`
	LogQueryParams bool     `json:"log_query_params"`
	ExcludeHeaders []string `json:"exclude_headers"`
	LogFormat      string   `json:"log_format"`
	LogLevel       string   `json:"log_level"`
}

var _ interfaces.Plugin = (*RequestLoggerPlugin)(nil)

func CreatePlugin() interfaces.Plugin {
	return &RequestLoggerPlugin{
		initialized:  false,
		requestCount: 0,
	}
}

func (p *RequestLoggerPlugin) Name() string {
	return "request-logger"
}

func (p *RequestLoggerPlugin) Version() string {
	return "1.0.0"
}

func (p *RequestLoggerPlugin) Description() string {
	return "Logs detailed information about incoming requests"
}

func (p *RequestLoggerPlugin) Author() string {
	return "Horizon Team"
}

func (p *RequestLoggerPlugin) Init(rawConfig map[string]interface{}) error {
	p.config = Config{
		LogHeaders:     true,
		LogBody:        false,
		LogQueryParams: true,
		ExcludeHeaders: []string{"authorization", "cookie"},
		LogFormat:      "json",
		LogLevel:       "info",
	}

	if rawConfig != nil {
		configBytes, err := json.Marshal(rawConfig)
		if err != nil {
			return fmt.Errorf("error marshaling config: %w", err)
		}

		if err := json.Unmarshal(configBytes, &p.config); err != nil {
			return fmt.Errorf("error unmarshaling config: %w", err)
		}
	}

	p.logger = logging.NewLogger()
	p.initialized = true

	return nil
}

func (p *RequestLoggerPlugin) Execute(phase interfaces.PluginPhase, data *interfaces.PluginData) error {
	if !p.initialized {
		return fmt.Errorf("plugin not initialized")
	}

	if phase != interfaces.PhaseRequest {
		return nil
	}

	p.requestCount++

	req := data.Request
	if req == nil {
		return fmt.Errorf("request data is nil")
	}

	fields := map[string]interface{}{
		"plugin":     p.Name(),
		"request_id": p.requestCount,
		"method":     req.Method,
		"path":       req.Path,
		"client_ip":  req.ClientIP,
		"timestamp":  time.Now().Format(time.RFC3339),
		"route_name": data.Route.Name,
		"status":     "received",
	}

	if p.config.LogQueryParams && req.QueryParams != nil {
		fields["query_params"] = req.QueryParams
	}

	if p.config.LogHeaders && req.Headers != nil {
		filteredHeaders := make(map[string][]string)
		for k, v := range req.Headers {
			shouldExclude := false
			for _, excluded := range p.config.ExcludeHeaders {
				if strings.EqualFold(k, excluded) {
					shouldExclude = true
					break
				}
			}

			if !shouldExclude {
				filteredHeaders[k] = v
			}
		}
		fields["headers"] = filteredHeaders
	}

	if p.config.LogBody && req.Body != nil && len(req.Body) > 0 {
		if !isTextContent(req.Headers) {
			fields["body_size"] = len(req.Body)
		} else {
			if len(req.Body) > 1024 {
				fields["body"] = string(req.Body[:1024]) + "... (truncated)"
			} else {
				fields["body"] = string(req.Body)
			}
		}
	}

	switch strings.ToLower(p.config.LogLevel) {
	case "debug":
		p.logger.Debug("Request received", fieldsToKeyValues(fields)...)
	case "info":
		p.logger.Info("Request received", fieldsToKeyValues(fields)...)
	case "warn":
		p.logger.Warn("Request received", fieldsToKeyValues(fields)...)
	case "error":
		p.logger.Error("Request received", fieldsToKeyValues(fields)...)
	default:
		p.logger.Info("Request received", fieldsToKeyValues(fields)...)
	}

	return nil
}

func (p *RequestLoggerPlugin) Shutdown() error {
	p.logger.Info("Shutting down request logger plugin",
		"name", p.Name(),
		"requests_processed", p.requestCount)
	return nil
}

func isTextContent(headers map[string][]string) bool {
	contentType := getHeader(headers, "Content-Type")
	if contentType == "" {
		return true
	}

	return strings.Contains(contentType, "text/") ||
		strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "application/xml") ||
		strings.Contains(contentType, "application/javascript")
}

func getHeader(headers map[string][]string, name string) string {
	for k, v := range headers {
		if strings.EqualFold(k, name) && len(v) > 0 {
			return v[0]
		}
	}
	return ""
}

func fieldsToKeyValues(fields map[string]interface{}) []interface{} {
	keyValues := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		keyValues = append(keyValues, k, v)
	}
	return keyValues
}
