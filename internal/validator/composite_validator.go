package validator

import (
	"context"
	"fmt"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type CompositeValidator struct {
	validators  []Validator
	logger      logging.Logger
	stopOnFirst bool
}

type CompositeValidationResult struct {
	Valid   bool                        `json:"valid"`
	Errors  []string                    `json:"errors,omitempty"`
	Message string                      `json:"message,omitempty"`
	Results map[string]ValidationResult `json:"results,omitempty"`
}

func NewCompositeValidator(validators []Validator, logger logging.Logger, stopOnFirst bool) *CompositeValidator {
	return &CompositeValidator{
		validators:  validators,
		logger:      logger,
		stopOnFirst: stopOnFirst,
	}
}

func (v *CompositeValidator) Validate(ctx context.Context, data interface{}) (ValidationResult, error) {
	result := &CompositeValidationResult{
		Valid:   true,
		Errors:  []string{},
		Results: make(map[string]ValidationResult),
	}

	validatorNames := ctx.Value("validator_names")
	var names []string
	if validatorNames != nil {
		names = validatorNames.([]string)
	}

	for i, validator := range v.validators {
		var validatorName string
		if i < len(names) {
			validatorName = names[i]
		} else {
			validatorName = fmt.Sprintf("validator_%d", i)
		}

		valResult, err := validator.Validate(ctx, data)
		if err != nil {
			v.logger.Warn("Validator error",
				"validator", validatorName,
				"error", err)
			continue
		}

		result.Results[validatorName] = valResult

		if !valResult.IsValid() {
			result.Valid = false
			result.Errors = append(result.Errors, valResult.GetErrors()...)

			if v.stopOnFirst {
				break
			}
		}
	}

	if !result.Valid {
		result.Message = "Validation failed"
	}

	return result, nil
}

func (r *CompositeValidationResult) IsValid() bool {
	return r.Valid
}

func (r *CompositeValidationResult) GetErrors() []string {
	return r.Errors
}

func (r *CompositeValidationResult) GetMessage() string {
	return r.Message
}

func (v *CompositeValidator) AddValidator(validator Validator) {
	v.validators = append(v.validators, validator)
}

func (v *CompositeValidator) SetStopOnFirst(stopOnFirst bool) {
	v.stopOnFirst = stopOnFirst
}

func (v *CompositeValidator) GetValidatorCount() int {
	return len(v.validators)
}
