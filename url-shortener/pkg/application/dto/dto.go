// Package dto implements types used for data transfer
package dto

import "time"

type UserInput struct {
	Email           string    `json:"email" validate:"required"`
	Password        string    `json:"password" validate:"required"`
	OauthProvider   string    `json:"oauth_provider"`
	OauthProviderID string    `json:"oauth_provider_id"`
}

type UserOutput struct {
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}
