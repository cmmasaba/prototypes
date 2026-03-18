// Package helpers contains common utilities
package helpers

import (
	"context"

	"github.com/go-playground/validator/v10"
)

type ctxKey string

const userIDKey ctxKey = "UserID"

var validate = validator.New()

// Validate returns nil if validation for the struct's exposed fields passes.
func Validate(v any) error {
	return validate.Struct(v)
}

// GetUserIDCtx returns userID from context, bool is false if userID is not found.
func GetUserIDCtx(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)

	return userID, ok
}

// SetUserIDCtx returns new context with userID value.
func SetUserIDCtx(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
