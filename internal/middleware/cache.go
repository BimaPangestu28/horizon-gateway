package middleware

import (
	"bytes"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/bimapangestu28/horizon/internal/cache"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	// Status is the HTTP status code
	Status int
	// Headers contains the response headers
	Headers http.Header
	// Body is the response body
	Body []byte
	// Created is when the response was cached
	Created time.Time
	// Expires is when the response expires
	Expires time.Time
	// ETag is the entity tag for the response (if any)
	ETag string
	// LastModified is the last modified time (if any)
	LastModified string
}

// CacheMiddleware creates a middleware for response caching
func CacheMiddleware(cacheStore cache.Cache, logger logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if caching is not configured
		if cacheStore == nil {
			return c.Next()
		}

		// Convert Fiber context to http.Request for the cache
		httpReq := &http.Request{
			Method: c.Method(),
			URL: &url.URL{
				Path:     c.Path(),
				RawQuery: string(c.Request().URI().QueryString()),
			},
			Header: make(http.Header),
			Host:   c.Hostname(),
		}

		// Copy headers
		c.Request().Header.VisitAll(func(key, value []byte) {
			httpReq.Header.Add(string(key), string(value))
		})

		// Generate cache key
		key := cacheStore.GenerateKey(httpReq)

		// Check for conditional requests
		ifNoneMatch := string(c.Request().Header.Peek("If-None-Match"))
		ifModifiedSince := string(c.Request().Header.Peek("If-Modified-Since"))

		// Try to get from cache
		if cachedResp, found := cacheStore.Get(key); found {
			// Convert from cache.CachedResponse to our local CachedResponse
			cached := &CachedResponse{
				Status:       cachedResp.Status,
				Headers:      cachedResp.Headers,
				Body:         cachedResp.Body,
				Created:      cachedResp.Created,
				Expires:      cachedResp.Expires,
				ETag:         cachedResp.ETag,
				LastModified: cachedResp.LastModified,
			}

			// For conditional requests
			if ifNoneMatch != "" && cached.ETag != "" && ifNoneMatch == cached.ETag {
				return c.Status(http.StatusNotModified).Send(nil)
			}

			if ifModifiedSince != "" && cached.LastModified != "" {
				if ifModSince, err := http.ParseTime(ifModifiedSince); err == nil {
					if lastMod, err := http.ParseTime(cached.LastModified); err == nil {
						if !lastMod.After(ifModSince) {
							return c.Status(http.StatusNotModified).Send(nil)
						}
					}
				}
			}

			// Set caching headers
			c.Response().Header.Set("X-Cache", "HIT")
			c.Response().Header.Set("Age", strconv.FormatInt(int64(time.Since(cached.Created).Seconds()), 10))

			// Copy headers from cached response
			for key, values := range cached.Headers {
				for _, value := range values {
					c.Set(key, value)
				}
			}

			// Send cached response
			return c.Status(cached.Status).Send(cached.Body)
		}

		// Cache miss - set header
		c.Response().Header.Set("X-Cache", "MISS")

		// Create a response capture
		responseBuffer := new(bytes.Buffer)

		// Process the request - but we need to intercept the response
		err := c.Next()

		// Check for errors
		if err != nil {
			return err
		}

		// Now we need to capture the response body
		// Get the body from the response
		responseBody := c.Response().Body()

		// Write the response to our buffer for caching
		responseBuffer.Write(responseBody)

		// Create http.Response for ShouldCache check
		httpResp := &http.Response{
			StatusCode:    c.Response().StatusCode(),
			Header:        make(http.Header),
			ContentLength: int64(len(responseBody)),
		}

		// Copy headers to http.Response
		c.Response().Header.VisitAll(func(key, value []byte) {
			httpResp.Header.Add(string(key), string(value))
		})

		// Check if response should be cached
		if !cacheStore.ShouldCache(httpReq, httpResp) {
			return nil
		}

		// Extract caching headers
		cacheControl := httpResp.Header.Get("Cache-Control")
		expires := httpResp.Header.Get("Expires")
		etag := httpResp.Header.Get("ETag")
		lastModified := httpResp.Header.Get("Last-Modified")

		// Determine TTL
		ttl := determineTTL(cacheControl, expires)

		// Create cached response
		cachedResp := &cache.CachedResponse{
			Status:       c.Response().StatusCode(),
			Headers:      httpResp.Header,
			Body:         responseBuffer.Bytes(),
			Created:      time.Now(),
			Expires:      time.Now().Add(ttl),
			ETag:         etag,
			LastModified: lastModified,
		}

		// Store in cache
		if err := cacheStore.Set(key, cachedResp); err != nil {
			logger.Warn("Failed to cache response",
				"error", err.Error(),
				"key", key,
			)
		}

		return nil
	}
}

// determineTTL calculates the TTL based on Cache-Control and Expires headers
func determineTTL(cacheControl, expires string) time.Duration {
	// Default TTL (1 minute)
	defaultTTL := 1 * time.Minute

	// Check Cache-Control: max-age
	if cacheControl != "" {
		parts := strings.Split(cacheControl, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "max-age=") {
				seconds, err := strconv.Atoi(strings.TrimPrefix(part, "max-age="))
				if err == nil && seconds > 0 {
					return time.Duration(seconds) * time.Second
				}
			}
		}
	}

	// Check Expires header
	if expires != "" {
		expiresTime, err := http.ParseTime(expires)
		if err == nil {
			ttl := expiresTime.Sub(time.Now())
			if ttl > 0 {
				return ttl
			}
		}
	}

	return defaultTTL
}
