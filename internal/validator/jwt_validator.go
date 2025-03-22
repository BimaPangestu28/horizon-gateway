package validator

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/golang-jwt/jwt"
)

type JWTValidator struct {
	logger    logging.Logger
	keyFunc   jwt.Keyfunc
	config    *JWTConfig
	parseOpts []jwt.ParserOption
}

type JWTConfig struct {
	Issuer           string   `yaml:"issuer" json:"issuer"`
	Audience         []string `yaml:"audience" json:"audience"`
	RequiredClaims   []string `yaml:"required_claims" json:"required_claims"`
	SigningMethod    string   `yaml:"signing_method" json:"signing_method"`
	SigningKey       string   `yaml:"signing_key" json:"signing_key"`
	VerificationKey  string   `yaml:"verification_key" json:"verification_key"`
	ClockSkewSeconds int      `yaml:"clock_skew_seconds" json:"clock_skew_seconds"`
	KeyID            string   `yaml:"kid" json:"kid"`
}

type JWTValidationResult struct {
	Valid     bool
	Token     *jwt.Token
	Claims    jwt.MapClaims
	SubjectID string
	Roles     []string
	Error     error
}

func NewJWTValidator(config *JWTConfig, logger logging.Logger) (*JWTValidator, error) {
	if config == nil {
		return nil, errors.New("JWT configuration is required")
	}

	if config.SigningMethod == "" {
		config.SigningMethod = "HS256"
	}

	if config.ClockSkewSeconds == 0 {
		config.ClockSkewSeconds = 30
	}

	validator := &JWTValidator{
		logger: logger,
		config: config,
	}

	validator.keyFunc = func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != config.SigningMethod {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		if config.KeyID != "" && token.Header["kid"] != config.KeyID {
			return nil, fmt.Errorf("unexpected key ID")
		}

		switch config.SigningMethod {
		case "HS256", "HS384", "HS512":
			return []byte(config.SigningKey), nil
		case "RS256", "RS384", "RS512", "ES256", "ES384", "ES512":
			if config.VerificationKey == "" {
				return nil, errors.New("verification key is required for RSA and ECDSA")
			}
			return jwt.ParseRSAPublicKeyFromPEM([]byte(config.VerificationKey))
		default:
			return nil, fmt.Errorf("unsupported signing method: %s", config.SigningMethod)
		}
	}

	validator.parseOpts = []jwt.ParserOption{
		jwt.WithLeeway(time.Duration(config.ClockSkewSeconds) * time.Second),
	}

	return validator, nil
}

func (v *JWTValidator) ValidateToken(tokenString string) *JWTValidationResult {
	result := &JWTValidationResult{
		Valid: false,
	}

	if tokenString == "" {
		result.Error = errors.New("token is empty")
		return result
	}

	// Clean the token string (remove 'Bearer ' prefix if present)
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	// Parse the token
	parser := jwt.NewParser(v.parseOpts...)
	token, err := parser.Parse(tokenString, v.keyFunc)

	if err != nil {
		result.Error = err
		return result
	}

	// Get claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		result.Error = errors.New("failed to parse claims")
		return result
	}

	result.Token = token
	result.Claims = claims

	// Store token validity
	result.Valid = token.Valid

	if !token.Valid {
		result.Error = errors.New("token is not valid")
		return result
	}

	// Validate issuer if specified
	if v.config.Issuer != "" {
		if iss, ok := claims["iss"].(string); !ok || iss != v.config.Issuer {
			result.Valid = false
			result.Error = errors.New("invalid issuer")
			return result
		}
	}

	// Validate audience if specified
	if len(v.config.Audience) > 0 {
		validAudience := false
		if aud, ok := claims["aud"].(string); ok {
			for _, a := range v.config.Audience {
				if a == aud {
					validAudience = true
					break
				}
			}
		} else if auds, ok := claims["aud"].([]interface{}); ok {
			for _, a := range v.config.Audience {
				for _, aud := range auds {
					if audStr, ok := aud.(string); ok && audStr == a {
						validAudience = true
						break
					}
				}
				if validAudience {
					break
				}
			}
		}

		if !validAudience {
			result.Valid = false
			result.Error = errors.New("invalid audience")
			return result
		}
	}

	// Check required claims
	for _, claim := range v.config.RequiredClaims {
		if _, ok := claims[claim]; !ok {
			result.Valid = false
			result.Error = fmt.Errorf("missing required claim: %s", claim)
			return result
		}
	}

	// Extract subject ID if present
	if sub, ok := claims["sub"].(string); ok {
		result.SubjectID = sub
	}

	// Extract roles if present
	result.Roles = extractRoles(claims)

	return result
}

func extractRoles(claims jwt.MapClaims) []string {
	var roles []string

	// Check various common role claim formats
	if rolesVal, ok := claims["roles"]; ok {
		roles = extractRolesFromAny(rolesVal)
	} else if rolesVal, ok := claims["role"]; ok {
		roles = extractRolesFromAny(rolesVal)
	} else if rolesVal, ok := claims["groups"]; ok {
		roles = extractRolesFromAny(rolesVal)
	} else if scope, ok := claims["scope"].(string); ok {
		// Extract roles from OAuth2 scope claim
		for _, s := range strings.Fields(scope) {
			if strings.HasPrefix(s, "role:") {
				roles = append(roles, strings.TrimPrefix(s, "role:"))
			}
		}
	}

	return roles
}

func extractRolesFromAny(rolesVal interface{}) []string {
	var roles []string

	switch v := rolesVal.(type) {
	case string:
		for _, role := range strings.Split(v, ",") {
			role = strings.TrimSpace(role)
			if role != "" {
				roles = append(roles, role)
			}
		}
	case []interface{}:
		for _, roleVal := range v {
			if role, ok := roleVal.(string); ok {
				role = strings.TrimSpace(role)
				if role != "" {
					roles = append(roles, role)
				}
			}
		}
	case []string:
		for _, role := range v {
			role = strings.TrimSpace(role)
			if role != "" {
				roles = append(roles, role)
			}
		}
	}

	return roles
}

func (v *JWTValidator) GenerateToken(claims map[string]interface{}, exp time.Duration) (string, error) {
	tokenClaims := jwt.MapClaims{}

	// Add standard claims
	now := time.Now()
	tokenClaims["iat"] = now.Unix()

	if exp > 0 {
		tokenClaims["exp"] = now.Add(exp).Unix()
	}

	if v.config.Issuer != "" {
		tokenClaims["iss"] = v.config.Issuer
	}

	if len(v.config.Audience) > 0 {
		tokenClaims["aud"] = v.config.Audience[0]
	}

	// Add custom claims
	for k, v := range claims {
		tokenClaims[k] = v
	}

	// Create token with claims
	var signingMethod jwt.SigningMethod

	switch v.config.SigningMethod {
	case "HS256":
		signingMethod = jwt.SigningMethodHS256
	case "HS384":
		signingMethod = jwt.SigningMethodHS384
	case "HS512":
		signingMethod = jwt.SigningMethodHS512
	case "RS256":
		signingMethod = jwt.SigningMethodRS256
	case "RS384":
		signingMethod = jwt.SigningMethodRS384
	case "RS512":
		signingMethod = jwt.SigningMethodRS512
	case "ES256":
		signingMethod = jwt.SigningMethodES256
	case "ES384":
		signingMethod = jwt.SigningMethodES384
	case "ES512":
		signingMethod = jwt.SigningMethodES512
	default:
		return "", fmt.Errorf("unsupported signing method: %s", v.config.SigningMethod)
	}

	token := jwt.NewWithClaims(signingMethod, tokenClaims)

	if v.config.KeyID != "" {
		token.Header["kid"] = v.config.KeyID
	}

	// Sign the token
	var signingKey interface{}

	switch v.config.SigningMethod {
	case "HS256", "HS384", "HS512":
		signingKey = []byte(v.config.SigningKey)
	case "RS256", "RS384", "RS512", "ES256", "ES384", "ES512":
		var err error
		signingKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(v.config.SigningKey))
		if err != nil {
			return "", fmt.Errorf("failed to parse RSA private key: %w", err)
		}
	}

	return token.SignedString(signingKey)
}
