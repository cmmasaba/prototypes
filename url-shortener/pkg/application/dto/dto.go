// Package dto implements types used for data transfer
package dto

import "time"

type EmailPasswordUserInput struct {
	Email    string `json:"email"    validate:"required"`
	Password string `json:"password" validate:"required"` // nolint: gosec
}

type UserOutput struct {
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type ValidatePasswordInput struct {
	Password string `json:"password" validate:"required"` // nolint: gosec
}
