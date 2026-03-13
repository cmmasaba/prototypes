// Package helpers contains common utilities
package helpers

import "github.com/go-playground/validator/v10"

var validate = validator.New()

// Validate returns nil if validation for the struct's exposed fields passes.
func Validate(v any) error {
	return validate.Struct(v)
}
