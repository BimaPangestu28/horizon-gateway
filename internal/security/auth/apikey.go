package auth

import (
	"context"
	"errors"
	"net/http"
	"time"
)

var (
	ErrMissingAPIKey     = errors.New("missing API key")
	ErrInvalidAPIKey     = errors.New("invalid API key")
	ErrExpiredAPIKey     = errors.New("expired API key")
	ErrInsufficientScope = errors.New("insufficient scope")
)

// APIKeyAuthenticator implements API key authentication
type APIKeyAuthenticator struct {
	config   *APIKeyConfig
	keyMap   map[string]APIKey
	cacheTTL time.Duration
}

// NewAPIKeyAuthenticator creates a new API key authenticator
func NewAPIKeyAuthenticator(config *APIKeyConfig) (*APIKeyAuthenticator, error) {
	auth := &APIKeyAuthenticator{
		config:   config,
		keyMap:   make(map[string]APIKey),
		cacheTTL: 5 * time.Minute, // Default cache duration
	}

	// Load keys from configuration
	if err := auth.loadKeys(); err != nil {
		return nil, err
	}

	return auth, nil
}

// loadKeys loads API keys from configuration or file
func (a *APIKeyAuthenticator) loadKeys() error {
	// Load keys directly from config
	for _, key := range a.config.Keys {
		a.keyMap[key.Key] = key
	}

	// Load keys from file (if specified)
	if a.config.KeysFile != "" {
		// Implementation for loading keys from file
		// ...
	}

	return nil
}

// GetType returns the authentication type
func (a *APIKeyAuthenticator) GetType() AuthType {
	return AuthTypeAPIKey
}

// Authenticate validates the API key in the request
func (a *APIKeyAuthenticator) Authenticate(ctx context.Context, req *http.Request) (map[string]interface{}, error) {
	// Extract API key from request
	apiKey := a.extractAPIKey(req)
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	// Validate API key
	key, found := a.keyMap[apiKey]
	if !found {
		return nil, ErrInvalidAPIKey
	}

	// Check expiration
	if key.Expires != "" {
		expires, err := time.Parse(time.RFC3339, key.Expires)
		if err == nil && time.Now().After(expires) {
			return nil, ErrExpiredAPIKey
		}
	}

	// Return key metadata
	metadata := make(map[string]interface{})
	metadata["name"] = key.Name

	for k, v := range key.Metadata {
		metadata[k] = v
	}

	if len(key.Scopes) > 0 {
		metadata["scopes"] = key.Scopes
	}

	return metadata, nil
}

// extractAPIKey extracts the API key from the request based on configuration
func (a *APIKeyAuthenticator) extractAPIKey(req *http.Request) string {
	// Check header
	if a.config.In == "header" || a.config.In == "both" {
		if key := req.Header.Get(a.config.Header); key != "" {
			return key
		}
	}

	// Check query parameter
	if a.config.In == "query" || a.config.In == "both" {
		if key := req.URL.Query().Get(a.config.Query); key != "" {
			return key
		}
	}

	return ""
}
