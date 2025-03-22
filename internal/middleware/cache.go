package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/horizon-gateway/horizon/internal/cache"
	"github.com/horizon-gateway/horizon/internal/utils/logging"
)

// CacheMiddleware creates a middleware for response caching
func CacheMiddleware(cache cache.Cache, logger logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if caching is not configured
		if cache == nil {
			return c.Next()
		}

		// Convert Fiber context to http.Request for the cache
		httpReq := &http.Request{
			Method: c.Method(),
			URL:    c.Request().URI().QueryArgs(),
			Header: make(http.Header),
			Host:   c.Hostname(),
		}

		// Copy headers
		c.Request().Header.VisitAll(func(key, value []byte) {
			httpReq.Header.Add(string(key), string(value))
		})

		// Generate cache key
		key := cache.GenerateKey(httpReq)

		// Check for conditional requests
		ifNoneMatch := string(c.Request().Header.Peek("If-None-Match"))
		ifModifiedSince := string(c.Request().Header.Peek("If-Modified-Since"))

		// Try to get from cache
		if cachedResp, found := cache.Get(key); found {
			// For conditional requests
			if ifNoneMatch != "" && cachedResp.ETag != "" && ifNoneMatch == cachedResp.ETag {
				return c.Status(http.StatusNotModified).Send(nil)
			}

			if ifModifiedSince != "" && cachedResp.LastModified != "" {
				if ifModSince, err := http.ParseTime(ifModifiedSince); err == nil {
					if lastMod, err := http.ParseTime(cachedResp.LastModified); err == nil {
						if !lastMod.After(ifModSince) {
							return c.Status(http.StatusNotModified).Send(nil)
						}
					}
				}
			}

			// Set caching headers
			c.Response().Header.Set("X-Cache", "HIT")
			c.Response().Header.Set("Age", strconv.FormatInt(int64(time.Since(cachedResp.Created).Seconds()), 10))

			// Copy headers from cached response
			for key, values := range cachedResp.Headers {
				for _, value := range values {
					c.Set(key, value)
				}
			}

			// Send cached response
			return c.Status(cachedResp.Status).Send(cachedResp.Body)
		}

		// Cache miss - set header
		c.Response().Header.Set("X-Cache", "MISS")

		// Create a response capture
		original := c.Response().BodyWriter()
		buffer := new(bytes.Buffer)

		// Replace response writer
		c.Response().SetBodyWriter(io.MultiWriter(original, buffer))

		// Process request
		err := c.Next()
		if err != nil {
			return err
		}

		// Create http.Response for ShouldCache check
		httpResp := &http.Response{
			StatusCode:    c.Response().StatusCode(),
			Header:        make(http.Header),
			ContentLength: int64(buffer.Len()),
		}

		// Copy headers to http.Response
		c.Response().Header.VisitAll(func(key, value []byte) {
			httpResp.Header.Add(string(key), string(value))
		})

		// Check if response should be cached
		if !cache.ShouldCache(httpReq, httpResp) {
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
			Body:         buffer.Bytes(),
			Created:      time.Now(),
			Expires:      time.Now().Add(ttl),
			ETag:         etag,
			LastModified: lastModified,
		}

		// Store in cache
		if err := cache.Set(key, cachedResp); err != nil {
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
