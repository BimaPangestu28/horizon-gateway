package auth

import (
	"fmt"
)

// NewAuthenticator creates an authenticator based on configuration
func NewAuthenticator(config *AuthConfig) (Authenticator, error) {
	if !config.Enabled {
		return nil, nil
	}

	// Create appropriate authenticator based on type
	switch config.Type {
	case AuthTypeAPIKey:
		return NewAPIKeyAuthenticator(config.APIKey)

	case AuthTypeJWT:
		return NewJWTAuthenticator(config.JWT)

	case AuthTypeOAuth2:
		// TODO: Implement OAuth2 authenticator
		return nil, fmt.Errorf("OAuth2 authentication not implemented yet")

	default:
		return nil, fmt.Errorf("unsupported authentication type: %s", config.Type)
	}
}
