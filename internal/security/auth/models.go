package auth

import (
	"context"
	"net/http"
)

// AuthType represents the supported authentication types
type AuthType string

const (
	// AuthTypeAPIKey represents API key authentication
	AuthTypeAPIKey AuthType = "api_key"

	// AuthTypeJWT represents JWT validation
	AuthTypeJWT AuthType = "jwt"

	// AuthTypeOAuth2 represents OAuth2 authentication
	AuthTypeOAuth2 AuthType = "oauth2"
)

// AuthConfig represents authentication configuration for a route
type AuthConfig struct {
	// Enabled indicates whether authentication is enabled
	Enabled bool `yaml:"enabled"`

	// Type is the authentication type (api_key, jwt, oauth2)
	Type AuthType `yaml:"type"`

	// APIKey configuration (when Type is api_key)
	APIKey *APIKeyConfig `yaml:"api_key,omitempty"`

	// JWT configuration (when Type is jwt)
	JWT *JWTConfig `yaml:"jwt,omitempty"`

	// OAuth2 configuration (when Type is oauth2)
	OAuth2 *OAuth2Config `yaml:"oauth2,omitempty"`
}

// APIKeyConfig defines API key authentication settings
type APIKeyConfig struct {
	// Header is the header name for the API key
	Header string `yaml:"header" default:"X-API-Key"`

	// Query is the query parameter name for the API key
	Query string `yaml:"query" default:"api_key"`

	// In specifies where to look for the API key (header, query, or both)
	In string `yaml:"in" default:"header"`

	// Keys is a list of valid API keys and their metadata
	Keys []APIKey `yaml:"keys,omitempty"`

	// KeysFile is a path to a file containing API keys
	KeysFile string `yaml:"keys_file,omitempty"`
}

// APIKey represents a single API key with metadata
type APIKey struct {
	// Key is the API key value
	Key string `yaml:"key"`

	// Name is a descriptive name for the key
	Name string `yaml:"name"`

	// Metadata contains additional information about the key
	Metadata map[string]string `yaml:"metadata,omitempty"`

	// Scopes defines the allowed scopes for this key
	Scopes []string `yaml:"scopes,omitempty"`

	// Expires is the expiration time for the key
	Expires string `yaml:"expires,omitempty"`
}

// JWTConfig defines JWT validation settings
type JWTConfig struct {
	// Secret is the HMAC secret key (for HS256, HS384, HS512)
	Secret string `yaml:"secret,omitempty"`

	// PublicKey is the RSA or ECDSA public key (for RS256, ES256, etc.)
	PublicKey string `yaml:"public_key,omitempty"`

	// PublicKeyFile is the path to the public key file
	PublicKeyFile string `yaml:"public_key_file,omitempty"`

	// Algorithm is the signing algorithm (HS256, RS256, etc.)
	Algorithm string `yaml:"algorithm" default:"HS256"`

	// Issuer is the expected issuer claim
	Issuer string `yaml:"issuer,omitempty"`

	// Audience is the expected audience claim
	Audience string `yaml:"audience,omitempty"`

	// ClaimsToHeaders maps JWT claims to request headers
	ClaimsToHeaders map[string]string `yaml:"claims_to_headers,omitempty"`
}

// OAuth2Config defines OAuth2 client settings
type OAuth2Config struct {
	// TokenIntrospectionURL is the URL for token introspection
	TokenIntrospectionURL string `yaml:"token_introspection_url"`

	// ClientID is the OAuth2 client ID
	ClientID string `yaml:"client_id,omitempty"`

	// ClientSecret is the OAuth2 client secret
	ClientSecret string `yaml:"client_secret,omitempty"`

	// Scopes defines the required scopes
	Scopes []string `yaml:"scopes,omitempty"`
}

// Authenticator interface defines authentication functionality
type Authenticator interface {
	// Authenticate validates the credentials and returns metadata or an error
	Authenticate(ctx context.Context, req *http.Request) (map[string]interface{}, error)

	// GetType returns the authentication type
	GetType() AuthType
}
