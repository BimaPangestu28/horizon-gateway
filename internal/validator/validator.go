package validator

import (
	"context"
	"errors"
	"net/http"
)

// Validator interface defines validation functionality
type Validator interface {
	Validate(ctx context.Context, data interface{}) (*ValidationResult, error)
}

// ValidationRegistry stores registered validators
type ValidatorRegistry struct {
	validators map[string]Validator
}

func NewValidatorRegistry() *ValidatorRegistry {
	return &ValidatorRegistry{
		validators: make(map[string]Validator),
	}
}

func (r *ValidatorRegistry) Register(name string, validator Validator) {
	r.validators[name] = validator
}

func (r *ValidatorRegistry) Get(name string) (Validator, bool) {
	validator, exists := r.validators[name]
	return validator, exists
}

func (r *ValidatorRegistry) GetAll() map[string]Validator {
	validators := make(map[string]Validator)
	for name, validator := range r.validators {
		validators[name] = validator
	}
	return validators
}

// Request validator adapter
type RequestValidatorAdapter struct {
	validator *RequestValidator
}

func NewRequestValidatorAdapter(validator *RequestValidator) *RequestValidatorAdapter {
	return &RequestValidatorAdapter{
		validator: validator,
	}
}

func (a *RequestValidatorAdapter) Validate(ctx context.Context, data interface{}) (*ValidationResult, error) {
	if req, ok := data.(*http.Request); ok {
		routeName := ctx.Value("route_name").(string)
		return a.validator.ValidateRequest(req, routeName)
	}
	return nil, ErrInvalidDataType
}

// JWT validator adapter
type JWTValidatorAdapter struct {
	validator *JWTValidator
}

func NewJWTValidatorAdapter(validator *JWTValidator) *JWTValidatorAdapter {
	return &JWTValidatorAdapter{
		validator: validator,
	}
}

func (a *JWTValidatorAdapter) Validate(ctx context.Context, data interface{}) (*ValidationResult, error) {
	if tokenString, ok := data.(string); ok {
		jwtResult := a.validator.ValidateToken(tokenString)
		// Convert JWTValidationResult to ValidationResult
		result := &ValidationResult{
			Valid: jwtResult.Valid,
		}
		if jwtResult.Error != nil {
			result.Errors = []string{jwtResult.Error.Error()}
			result.Message = jwtResult.Error.Error()
		}
		return result, nil
	}
	return nil, ErrInvalidDataType
}

// Schema validator adapter
type SchemaValidatorAdapter struct {
	validator *SchemaValidator
}

func NewSchemaValidatorAdapter(validator *SchemaValidator) *SchemaValidatorAdapter {
	return &SchemaValidatorAdapter{
		validator: validator,
	}
}

func (a *SchemaValidatorAdapter) Validate(ctx context.Context, data interface{}) (*ValidationResult, error) {
	if req, ok := data.(*http.Request); ok {
		schemaResult, err := a.validator.ValidateRequest(req)
		if err != nil {
			return nil, err
		}
		// Convert SchemaValidationResult to ValidationResult
		result := &ValidationResult{
			Valid:   schemaResult.Valid,
			Errors:  schemaResult.Errors,
			Message: schemaResult.Message,
		}
		return result, nil
	}
	return nil, ErrInvalidDataType
}

// GRPC validator adapter
type GRPCValidatorAdapter struct {
	validator *GRPCValidator
}

func NewGRPCValidatorAdapter(validator *GRPCValidator) *GRPCValidatorAdapter {
	return &GRPCValidatorAdapter{
		validator: validator,
	}
}

func (a *GRPCValidatorAdapter) Validate(ctx context.Context, data interface{}) (*ValidationResult, error) {
	if bytes, ok := data.([]byte); ok {
		msgType := ctx.Value("message_type").(string)
		grpcResult, err := a.validator.ValidateMessage(msgType, bytes)
		if err != nil {
			return nil, err
		}
		// Convert GRPCValidationResult to ValidationResult
		result := &ValidationResult{
			Valid:   grpcResult.Valid,
			Errors:  grpcResult.Errors,
			Message: grpcResult.Message,
		}
		return result, nil
	}
	return nil, ErrInvalidDataType
}

// URL validator adapter
type URLValidatorAdapter struct {
	validator *URLValidator
}

func NewURLValidatorAdapter(validator *URLValidator) *URLValidatorAdapter {
	return &URLValidatorAdapter{
		validator: validator,
	}
}

func (a *URLValidatorAdapter) Validate(ctx context.Context, data interface{}) (*ValidationResult, error) {
	if urlStr, ok := data.(string); ok {
		urlResult := a.validator.ValidateURL(urlStr)
		// Convert URLValidationResult to ValidationResult
		result := &ValidationResult{
			Valid:   urlResult.Valid,
			Errors:  urlResult.Errors,
			Message: urlResult.Message,
		}
		return result, nil
	}
	return nil, ErrInvalidDataType
}

var (
	ErrInvalidDataType = errors.New("invalid data type for validator")
)
