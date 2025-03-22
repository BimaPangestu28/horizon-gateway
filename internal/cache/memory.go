package cache

import (
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

// MemoryCache implements in-memory caching
type MemoryCache struct {
	config    *CacheConfig
	cache     map[string]*CachedResponse
	mu        sync.RWMutex
	ttl       time.Duration
	keyParams keyParameters
}

// keyParameters contains the parameters used for cache key generation
type keyParameters struct {
	template       string
	ignoreParams   map[string]bool
	varyHeaders    map[string]bool
	methodsToCache map[string]bool
	includeHost    bool
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(config *CacheConfig) (*MemoryCache, error) {
	// Parse TTL
	ttl, err := time.ParseDuration(config.TTL)
	if err != nil {
		return nil, err
	}

	// Create key parameters
	keyParams := keyParameters{
		template:       config.CacheKeyTemplate,
		ignoreParams:   make(map[string]bool),
		varyHeaders:    make(map[string]bool),
		methodsToCache: make(map[string]bool),
		includeHost:    config.IncludeHost,
	}

	// Initialize ignore parameters
	for _, param := range config.IgnoreQueryParams {
		keyParams.ignoreParams[param] = true
	}

	// Initialize vary headers
	for _, header := range config.VaryHeaders {
		keyParams.varyHeaders[header] = true
	}

	// Initialize methods to cache
	if len(config.Methods) == 0 {
		keyParams.methodsToCache["GET"] = true
	} else {
		for _, method := range config.Methods {
			keyParams.methodsToCache[method] = true
		}
	}

	cache := &MemoryCache{
		config:    config,
		cache:     make(map[string]*CachedResponse),
		ttl:       ttl,
		keyParams: keyParams,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache, nil
}

// Get retrieves a cached response
func (c *MemoryCache) Get(key string) (*CachedResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	resp, found := c.cache[key]
	if !found {
		return nil, false
	}

	// Check if response has expired
	if time.Now().After(resp.Expires) {
		return nil, false
	}

	return resp, true
}

// Set stores a response in the cache
func (c *MemoryCache) Set(key string, response *CachedResponse) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = response
	return nil
}

// Delete removes a cached response
func (c *MemoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, key)
	return nil
}

// Clear clears all cached responses
func (c *MemoryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*CachedResponse)
	return nil
}

// GenerateKey generates a cache key from a request
func (c *MemoryCache) GenerateKey(req *http.Request) string {
	// Start with template
	template := c.keyParams.template

	// Replace placeholders
	key := template

	// Replace method
	key = strings.ReplaceAll(key, "{method}", req.Method)

	// Replace path
	key = strings.ReplaceAll(key, "{path}", req.URL.Path)

	// Replace host if needed
	if c.keyParams.includeHost {
		key = strings.ReplaceAll(key, "{host}", req.Host)
	}

	// Replace query params
	if strings.Contains(key, "{query}") {
		query := c.processQueryParams(req.URL.Query())
		key = strings.ReplaceAll(key, "{query}", query)
	}

	// Add vary headers if any
	if len(c.keyParams.varyHeaders) > 0 {
		headerValues := make([]string, 0, len(c.keyParams.varyHeaders))

		for header := range c.keyParams.varyHeaders {
			value := req.Header.Get(header)
			if value != "" {
				headerValues = append(headerValues, header+"="+value)
			}
		}

		if len(headerValues) > 0 {
			key += ":" + strings.Join(headerValues, ":")
		}
	}

	return key
}

// processQueryParams builds a string from query parameters, excluding ignored ones
func (c *MemoryCache) processQueryParams(query url.Values) string {
	if len(query) == 0 {
		return ""
	}

	parts := make([]string, 0, len(query))

	for param, values := range query {
		// Skip ignored parameters
		if c.keyParams.ignoreParams[param] {
			continue
		}

		// Sort values for consistency
		sort.Strings(values)

		// Join multiple values with comma
		parts = append(parts, param+"="+strings.Join(values, ","))
	}

	// Sort for consistent order
	sort.Strings(parts)

	return strings.Join(parts, "&")
}

// ShouldCache determines if a request/response should be cached
func (c *MemoryCache) ShouldCache(req *http.Request, resp *http.Response) bool {
	// Check if method should be cached
	if !c.keyParams.methodsToCache[req.Method] {
		return false
	}

	// Check never cache list
	for _, path := range c.config.NeverCache {
		if pathMatch(req.URL.Path, path) {
			return false
		}
	}

	// Check always cache list
	for _, path := range c.config.AlwaysCache {
		if pathMatch(req.URL.Path, path) {
			return true
		}
	}

	// Check response status code (only cache 200 OK)
	if resp.StatusCode != http.StatusOK {
		return false
	}

	// Check Cache-Control headers if configured to respect them
	if c.config.RespectCacheControl {
		cacheControl := resp.Header.Get("Cache-Control")

		// Don't cache if Cache-Control: no-store, no-cache, or private
		if strings.Contains(cacheControl, "no-store") ||
			strings.Contains(cacheControl, "no-cache") ||
			strings.Contains(cacheControl, "private") {
			return false
		}
	}

	// Check max size if configured
	if c.config.MaxSize > 0 {
		contentLength := resp.ContentLength
		if contentLength > c.config.MaxSize {
			return false
		}
	}

	return true
}

// Close closes the cache
func (c *MemoryCache) Close() error {
	return nil
}

// pathMatch checks if a request path matches a pattern
func pathMatch(path, pattern string) bool {
	// Convert pattern to prefix match if it ends with a wildcard
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, prefix)
	}

	// Exact match for non-wildcard patterns
	return path == pattern
}

// cleanup periodically removes expired items
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(c.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()

		for key, resp := range c.cache {
			if now.After(resp.Expires) {
				delete(c.cache, key)
			}
		}
		c.mu.Unlock()
	}
}
