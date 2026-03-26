// Package helpers contains common utilities
package helpers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
)

type ctxKey string

const userIDKey ctxKey = "UserID"

var (
	validate   = validator.New()
	hashSecret = MustGetEnvVar("HMAC_HASH_SECRET")
)

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

// MustGetEnvVar returns value of env var by given nae, panics if env var is not found.
func MustGetEnvVar(name string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		panic(fmt.Sprintf("env var %s not set", name))
	}

	return value
}

// HashSecret returns a sha256 hash of the secret.
func HashSecret(secret string) string {
	hasher := hmac.New(sha256.New, []byte(hashSecret))

	hasher.Write([]byte(secret))

	return hex.EncodeToString(hasher.Sum(nil))
}
