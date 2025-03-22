package validator

import (
	"fmt"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type ValidatorFactory struct {
	logger   logging.Logger
	registry *ValidatorRegistry
}

type ValidatorConfig struct {
	RequestValidator *ValidationConfig        `yaml:"request_validator" json:"request_validator"`
	JWTValidator     *JWTConfig               `yaml:"jwt_validator" json:"jwt_validator"`
	SchemaValidator  *OpenAPIValidationConfig `yaml:"schema_validator" json:"schema_validator"`
	GRPCValidator    *GRPCValidationConfig    `yaml:"grpc_validator" json:"grpc_validator"`
	URLValidator     *URLValidationConfig     `yaml:"url_validator" json:"url_validator"`
}

func NewValidatorFactory(logger logging.Logger) *ValidatorFactory {
	return &ValidatorFactory{
		logger:   logger,
		registry: NewValidatorRegistry(),
	}
}

func (f *ValidatorFactory) CreateValidator(validatorType string, config interface{}) (Validator, error) {
	switch validatorType {
	case "request":
		requestConfig, ok := config.(*ValidationConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for request validator")
		}

		validator, err := NewRequestValidator(requestConfig, f.logger)
		if err != nil {
			return nil, err
		}

		return NewRequestValidatorAdapter(validator), nil

	case "jwt":
		jwtConfig, ok := config.(*JWTConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for JWT validator")
		}

		validator, err := NewJWTValidator(jwtConfig, f.logger)
		if err != nil {
			return nil, err
		}

		return NewJWTValidatorAdapter(validator), nil

	case "schema":
		schemaConfig, ok := config.(*OpenAPIValidationConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for schema validator")
		}

		validator, err := NewSchemaValidator(schemaConfig, f.logger)
		if err != nil {
			return nil, err
		}

		return NewSchemaValidatorAdapter(validator), nil

	case "grpc":
		grpcConfig, ok := config.(*GRPCValidationConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for gRPC validator")
		}

		validator, err := NewGRPCValidator(grpcConfig, f.logger)
		if err != nil {
			return nil, err
		}

		return NewGRPCValidatorAdapter(validator), nil

	case "url":
		urlConfig, ok := config.(*URLValidationConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for URL validator")
		}

		validator, err := NewURLValidator(urlConfig, f.logger)
		if err != nil {
			return nil, err
		}

		return NewURLValidatorAdapter(validator), nil

	default:
		return nil, fmt.Errorf("unknown validator type: %s", validatorType)
	}
}

func (f *ValidatorFactory) CreateCompositeValidator(configs map[string]interface{}, stopOnFirst bool) (*CompositeValidator, error) {
	validators := make([]Validator, 0, len(configs))

	for validatorType, config := range configs {
		validator, err := f.CreateValidator(validatorType, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create validator %s: %w", validatorType, err)
		}

		validators = append(validators, validator)
	}

	return NewCompositeValidator(validators, f.logger, stopOnFirst), nil
}

func (f *ValidatorFactory) CreateFromConfig(config *ValidatorConfig) (map[string]Validator, error) {
	validators := make(map[string]Validator)

	if config.RequestValidator != nil {
		validator, err := f.CreateValidator("request", config.RequestValidator)
		if err != nil {
			return nil, err
		}
		validators["request"] = validator
	}

	if config.JWTValidator != nil {
		validator, err := f.CreateValidator("jwt", config.JWTValidator)
		if err != nil {
			return nil, err
		}
		validators["jwt"] = validator
	}

	if config.SchemaValidator != nil {
		validator, err := f.CreateValidator("schema", config.SchemaValidator)
		if err != nil {
			return nil, err
		}
		validators["schema"] = validator
	}

	if config.GRPCValidator != nil {
		validator, err := f.CreateValidator("grpc", config.GRPCValidator)
		if err != nil {
			return nil, err
		}
		validators["grpc"] = validator
	}

	if config.URLValidator != nil {
		validator, err := f.CreateValidator("url", config.URLValidator)
		if err != nil {
			return nil, err
		}
		validators["url"] = validator
	}

	return validators, nil
}

func (f *ValidatorFactory) RegisterValidator(name string, validator Validator) {
	f.registry.Register(name, validator)
}

func (f *ValidatorFactory) GetValidator(name string) (Validator, bool) {
	return f.registry.Get(name)
}

func (f *ValidatorFactory) GetRegistry() *ValidatorRegistry {
	return f.registry
}
