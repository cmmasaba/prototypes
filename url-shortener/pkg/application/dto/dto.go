// Package dto implements types used for data transfer
package dto

import "time"

type EmailPasswordUserInput struct {
	Email    string `json:"email"    validate:"required"`
	Password string `json:"password" validate:"required"` // nolint: gosec
}

type User struct {
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}

type ValidatePasswordInput struct {
	Password string `json:"password" validate:"required"` // nolint: gosec
}

type LoginInput struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"` // nolint: gosec
}

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`  // nolint: gosec
	RefreshToken string `json:"refreshToken"` // nolint: gosec
	ExpiresAt    string `json:"expiresAt"`
}

type RefreshToken struct {
	ID        int64
	UserID    int64
	Revoked   bool
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}
