package validator

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type URLValidator struct {
	logger    logging.Logger
	config    *URLValidationConfig
	patterns  map[string]*regexp.Regexp
	whiteList map[string]bool
	blackList map[string]bool
}

type URLValidationConfig struct {
	AllowedSchemes []string            `yaml:"allowed_schemes" json:"allowed_schemes"`
	AllowedDomains []string            `yaml:"allowed_domains" json:"allowed_domains"`
	BlockedDomains []string            `yaml:"blocked_domains" json:"blocked_domains"`
	PathPatterns   map[string]string   `yaml:"path_patterns" json:"path_patterns"`
	RequireParams  map[string][]string `yaml:"require_params" json:"require_params"`
	DisallowParams []string            `yaml:"disallow_params" json:"disallow_params"`
	MaxURLLength   int                 `yaml:"max_url_length" json:"max_url_length"`
}

type URLValidationResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Message string   `json:"message,omitempty"`
	URL     *url.URL `json:"url,omitempty"`
}

func NewURLValidator(config *URLValidationConfig, logger logging.Logger) (*URLValidator, error) {
	validator := &URLValidator{
		logger:    logger,
		patterns:  make(map[string]*regexp.Regexp),
		whiteList: make(map[string]bool),
		blackList: make(map[string]bool),
	}

	if config == nil {
		config = &URLValidationConfig{
			AllowedSchemes: []string{"http", "https"},
			MaxURLLength:   2048,
		}
	}

	validator.config = config

	for name, pattern := range config.PathPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid path pattern '%s': %w", pattern, err)
		}
		validator.patterns[name] = re
	}

	for _, domain := range config.AllowedDomains {
		validator.whiteList[strings.ToLower(domain)] = true
	}

	for _, domain := range config.BlockedDomains {
		validator.blackList[strings.ToLower(domain)] = true
	}

	return validator, nil
}

func (v *URLValidator) ValidateURL(urlStr string) *URLValidationResult {
	result := &URLValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	// Check URL length
	if v.config.MaxURLLength > 0 && len(urlStr) > v.config.MaxURLLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("URL exceeds maximum length of %d characters", v.config.MaxURLLength))
		result.Message = "URL validation failed"
		return result
	}

	// Parse the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid URL format: %s", err.Error()))
		result.Message = "URL validation failed"
		return result
	}

	result.URL = parsedURL

	// Check scheme
	if len(v.config.AllowedSchemes) > 0 {
		schemeAllowed := false
		for _, scheme := range v.config.AllowedSchemes {
			if strings.EqualFold(parsedURL.Scheme, scheme) {
				schemeAllowed = true
				break
			}
		}

		if !schemeAllowed {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("URL scheme '%s' is not allowed. Allowed schemes: %s",
				parsedURL.Scheme, strings.Join(v.config.AllowedSchemes, ", ")))
		}
	}

	// Check domain
	if len(v.config.AllowedDomains) > 0 && parsedURL.Host != "" {
		host := strings.ToLower(parsedURL.Host)
		domainAllowed := false

		for domain := range v.whiteList {
			if host == domain || strings.HasSuffix(host, "."+domain) {
				domainAllowed = true
				break
			}
		}

		if !domainAllowed {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("URL host '%s' is not in the allowed domains list", parsedURL.Host))
		}
	}

	// Check blocked domains
	if len(v.config.BlockedDomains) > 0 && parsedURL.Host != "" {
		host := strings.ToLower(parsedURL.Host)

		for domain := range v.blackList {
			if host == domain || strings.HasSuffix(host, "."+domain) {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("URL host '%s' is in the blocked domains list", parsedURL.Host))
				break
			}
		}
	}

	// Check path patterns
	for name, pattern := range v.patterns {
		if !pattern.MatchString(parsedURL.Path) {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("URL path does not match the required pattern '%s'", name))
		}
	}

	// Check required query parameters
	if len(v.config.RequireParams) > 0 {
		query := parsedURL.Query()

		for routeName, requiredParams := range v.config.RequireParams {
			// Only apply to matching routes
			if routeName != "*" && !strings.Contains(urlStr, routeName) {
				continue
			}

			for _, param := range requiredParams {
				if _, exists := query[param]; !exists {
					result.Valid = false
					result.Errors = append(result.Errors, fmt.Sprintf("Required query parameter '%s' is missing", param))
				}
			}
		}
	}

	// Check disallowed query parameters
	if len(v.config.DisallowParams) > 0 {
		query := parsedURL.Query()

		for _, param := range v.config.DisallowParams {
			if _, exists := query[param]; exists {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("Query parameter '%s' is not allowed", param))
			}
		}
	}

	if !result.Valid {
		result.Message = "URL validation failed"
	}

	return result
}

func (v *URLValidator) AddAllowedDomain(domain string) {
	v.whiteList[strings.ToLower(domain)] = true
}

func (v *URLValidator) AddBlockedDomain(domain string) {
	v.blackList[strings.ToLower(domain)] = true
}

func (v *URLValidator) AddPathPattern(name, pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid path pattern '%s': %w", pattern, err)
	}

	v.patterns[name] = re
	return nil
}

func (v *URLValidator) RemoveAllowedDomain(domain string) {
	delete(v.whiteList, strings.ToLower(domain))
}

func (v *URLValidator) RemoveBlockedDomain(domain string) {
	delete(v.blackList, strings.ToLower(domain))
}

func (v *URLValidator) RemovePathPattern(name string) {
	delete(v.patterns, name)
}
