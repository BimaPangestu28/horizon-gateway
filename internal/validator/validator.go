package validator

import (
	"context"
	"errors"
	"net/http"
)

type Validator interface {
	Validate(ctx context.Context, data interface{}) (ValidationResult, error)
}

type ValidatorRegistry struct {
	validators map[string]Validator
}

type ValidationResult interface {
	IsValid() bool
	GetErrors() []string
	GetMessage() string
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

type RequestValidatorAdapter struct {
	validator *RequestValidator
}

func NewRequestValidatorAdapter(validator *RequestValidator) *RequestValidatorAdapter {
	return &RequestValidatorAdapter{
		validator: validator,
	}
}

func (a *RequestValidatorAdapter) Validate(ctx context.Context, data interface{}) (ValidationResult, error) {
	if req, ok := data.(*http.Request); ok {
		routeName := ctx.Value("route_name").(string)
		return a.validator.ValidateRequest(req, routeName)
	}
	return nil, ErrInvalidDataType
}

type JWTValidatorAdapter struct {
	validator *JWTValidator
}

func NewJWTValidatorAdapter(validator *JWTValidator) *JWTValidatorAdapter {
	return &JWTValidatorAdapter{
		validator: validator,
	}
}

func (a *JWTValidatorAdapter) Validate(ctx context.Context, data interface{}) (ValidationResult, error) {
	if tokenString, ok := data.(string); ok {
		return a.validator.ValidateToken(tokenString), nil
	}
	return nil, ErrInvalidDataType
}

type SchemaValidatorAdapter struct {
	validator *SchemaValidator
}

func NewSchemaValidatorAdapter(validator *SchemaValidator) *SchemaValidatorAdapter {
	return &SchemaValidatorAdapter{
		validator: validator,
	}
}

func (a *SchemaValidatorAdapter) Validate(ctx context.Context, data interface{}) (ValidationResult, error) {
	if req, ok := data.(*http.Request); ok {
		return a.validator.ValidateRequest(req)
	}
	return nil, ErrInvalidDataType
}

type GRPCValidatorAdapter struct {
	validator *GRPCValidator
}

func NewGRPCValidatorAdapter(validator *GRPCValidator) *GRPCValidatorAdapter {
	return &GRPCValidatorAdapter{
		validator: validator,
	}
}

func (a *GRPCValidatorAdapter) Validate(ctx context.Context, data interface{}) (ValidationResult, error) {
	if bytes, ok := data.([]byte); ok {
		msgType := ctx.Value("message_type").(string)
		return a.validator.ValidateMessage(msgType, bytes)
	}
	return nil, ErrInvalidDataType
}

type URLValidatorAdapter struct {
	validator *URLValidator
}

func NewURLValidatorAdapter(validator *URLValidator) *URLValidatorAdapter {
	return &URLValidatorAdapter{
		validator: validator,
	}
}

func (a *URLValidatorAdapter) Validate(ctx context.Context, data interface{}) (ValidationResult, error) {
	if urlStr, ok := data.(string); ok {
		return a.validator.ValidateURL(urlStr), nil
	}
	return nil, ErrInvalidDataType
}

var (
	ErrInvalidDataType = errors.New("invalid data type for validator")
)

// Make ValidationResult implementations conform to the interface

func (r *ValidationResult) IsValid() bool {
	return r.Valid
}

func (r *ValidationResult) GetErrors() []string {
	return r.Errors
}

func (r *ValidationResult) GetMessage() string {
	return r.Message
}

func (r *JWTValidationResult) IsValid() bool {
	return r.Valid
}

func (r *JWTValidationResult) GetErrors() []string {
	if r.Error != nil {
		return []string{r.Error.Error()}
	}
	return nil
}

func (r *JWTValidationResult) GetMessage() string {
	if r.Error != nil {
		return r.Error.Error()
	}
	return ""
}

func (r *SchemaValidationResult) IsValid() bool {
	return r.Valid
}

func (r *SchemaValidationResult) GetErrors() []string {
	return r.Errors
}

func (r *SchemaValidationResult) GetMessage() string {
	return r.Message
}

func (r *GRPCValidationResult) IsValid() bool {
	return r.Valid
}

func (r *GRPCValidationResult) GetErrors() []string {
	return r.Errors
}

func (r *GRPCValidationResult) GetMessage() string {
	return r.Message
}

func (r *URLValidationResult) IsValid() bool {
	return r.Valid
}

func (r *URLValidationResult) GetErrors() []string {
	return r.Errors
}

func (r *URLValidationResult) GetMessage() string {
	return r.Message
}
