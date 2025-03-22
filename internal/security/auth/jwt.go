package auth

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrMissingJWT       = errors.New("missing JWT token")
	ErrInvalidJWT       = errors.New("invalid JWT token")
	ErrJWTExpired       = errors.New("JWT token expired")
	ErrInvalidSignature = errors.New("invalid JWT signature")
	ErrInvalidClaims    = errors.New("invalid JWT claims")
)

// JWTAuthenticator implements JWT authentication
type JWTAuthenticator struct {
	config    *JWTConfig
	verifyKey interface{} // HMAC secret or RSA/ECDSA public key
}

// NewJWTAuthenticator creates a new JWT authenticator
func NewJWTAuthenticator(config *JWTConfig) (*JWTAuthenticator, error) {
	auth := &JWTAuthenticator{
		config: config,
	}

	// Load verification key based on algorithm
	if err := auth.loadVerificationKey(); err != nil {
		return nil, err
	}

	return auth, nil
}

// loadVerificationKey loads the key used to verify JWT signatures
func (a *JWTAuthenticator) loadVerificationKey() error {
	alg := a.config.Algorithm

	// For HMAC algorithms (HS256, HS384, HS512)
	if strings.HasPrefix(alg, "HS") {
		if a.config.Secret == "" {
			return errors.New("HMAC secret key is required for " + alg)
		}
		a.verifyKey = []byte(a.config.Secret)
		return nil
	}

	// For RSA algorithms (RS256, RS384, RS512)
	if strings.HasPrefix(alg, "RS") {
		// Load public key from file if specified
		if a.config.PublicKeyFile != "" {
			keyData, err := ioutil.ReadFile(a.config.PublicKeyFile)
			if err != nil {
				return fmt.Errorf("reading public key file: %w", err)
			}
			publicKey, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
			if err != nil {
				return fmt.Errorf("parsing RSA public key: %w", err)
			}
			a.verifyKey = publicKey
			return nil
		}

		// Parse public key from config
		if a.config.PublicKey != "" {
			publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(a.config.PublicKey))
			if err != nil {
				return fmt.Errorf("parsing RSA public key: %w", err)
			}
			a.verifyKey = publicKey
			return nil
		}

		return errors.New("RSA public key is required for " + alg)
	}

	// Add support for other algorithms (ES256, etc.) as needed

	return fmt.Errorf("unsupported algorithm: %s", alg)
}

// GetType returns the authentication type
func (a *JWTAuthenticator) GetType() AuthType {
	return AuthTypeJWT
}

// Authenticate validates the JWT in the request
func (a *JWTAuthenticator) Authenticate(ctx context.Context, req *http.Request) (map[string]interface{}, error) {
	// Extract JWT from request
	tokenString := a.extractJWT(req)
	if tokenString == "" {
		return nil, ErrMissingJWT
	}

	// Parse JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing algorithm
		if jwt.GetSigningMethod(a.config.Algorithm) != token.Method {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Return verification key
		return a.verifyKey, nil
	})

	// Handle parsing errors
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrJWTExpired
			}
			if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
				return nil, ErrInvalidSignature
			}
		}
		return nil, fmt.Errorf("validating JWT: %w", err)
	}

	// Validate claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Validate issuer if specified
		if a.config.Issuer != "" {
			if iss, ok := claims["iss"].(string); !ok || iss != a.config.Issuer {
				return nil, fmt.Errorf("invalid issuer: %v", claims["iss"])
			}
		}

		// Validate audience if specified
		if a.config.Audience != "" {
			if aud, ok := claims["aud"].(string); !ok || aud != a.config.Audience {
				return nil, fmt.Errorf("invalid audience: %v", claims["aud"])
			}
		}

		// Return claims as metadata
		metadata := make(map[string]interface{})
		for key, value := range claims {
			metadata[key] = value
		}

		return metadata, nil
	}

	return nil, ErrInvalidClaims
}

// extractJWT extracts the JWT from the request
func (a *JWTAuthenticator) extractJWT(req *http.Request) string {
	// Check Authorization header
	authHeader := req.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

// AddClaimsToHeaders adds JWT claims to request headers (exported version)
func (a *JWTAuthenticator) AddClaimsToHeaders(req *http.Request, claims map[string]interface{}) {
	a.addClaimsToHeaders(req, claims)
}

// addClaimsToHeaders adds JWT claims to request headers
func (a *JWTAuthenticator) addClaimsToHeaders(req *http.Request, claims map[string]interface{}) {
	for claimKey, headerName := range a.config.ClaimsToHeaders {
		if value, ok := claims[claimKey]; ok {
			// Convert value to string
			var strValue string
			switch v := value.(type) {
			case string:
				strValue = v
			case float64:
				strValue = fmt.Sprintf("%v", v)
			case bool:
				strValue = fmt.Sprintf("%v", v)
			default:
				continue
			}

			req.Header.Set(headerName, strValue)
		}
	}
}

// GetClaimsToHeaders returns the claims to headers mapping
func (a *JWTAuthenticator) GetClaimsToHeaders() map[string]string {
	return a.config.ClaimsToHeaders
}
