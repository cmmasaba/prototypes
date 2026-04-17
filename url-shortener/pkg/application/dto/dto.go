// Package dto implements types used for data transfer
package dto

import "time"

const AnonymousUserID = "anonymous-user"

type EmailPasswordUserInput struct {
	Email    string `json:"email"    validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=128"` // nolint: gosec
}

type User struct {
	PublicID  string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type OTPRequiredResponse struct {
	User   User   `json:"user"`
	Status string `json:"status"`
}

type AuthResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token,omitempty"`  // nolint: gosec
	RefreshToken string `json:"refresh_token,omitempty"` // nolint: gosec
	ExpiresIn    int64  `json:"expires_in,omitempty"`
}

type ValidatePasswordInput struct {
	Password string `json:"password" validate:"required"` // nolint: gosec
}

type LoginInput struct {
	Email    string `json:"email"    validate:"required"`
	Password string `json:"password" validate:"required"` // nolint: gosec
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`  // nolint: gosec
	RefreshToken string `json:"refresh_token"` // nolint: gosec
	ExpiresIn    int64  `json:"expires_in"`
}

type RefreshToken struct {
	ID        int64
	UserID    int64
	Revoked   bool
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type RefreshAccessTokenInput struct {
	Token string `json:"refresh_token" validate:"required"` // nolint: gosec
}

type RefreshAccessTokenResponse struct {
	AccessToken  string `json:"access_token,omitempty"`  // nolint: gosec
	RefreshToken string `json:"refresh_token,omitempty"` // nolint: gosec
	ExpiresIn    int64  `json:"expires_in"`
}

type VerifyOTPInput struct {
	Purpose OTPPurpose `json:"purpose" validate:"required"`
	Value   string     `json:"value"   validate:"required"`
}

type RequestOTPInput struct {
	UserPublicID string     `json:"user_id" validate:"required"`
	Recipient    string     `json:"recipient" validate:"required"`
	Purpose      OTPPurpose `json:"purpose" validate:"required"`
}

type ShortenURLInput struct {
	URL       string    `json:"url" validate:"required"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ShortenURLResponse struct {
	ShortURL       string `json:"short_url"`
	OwnershipToken string `json:"ownership_token,omitempty"`
}
