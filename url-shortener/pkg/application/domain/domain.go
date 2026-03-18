// Package domain represents domain specific entities
package domain

import "time"

type Link struct {
	ID             int64
	ShortCode      string
	OriginalURL    string
	OwnershipToken string
	CreatedAt      time.Time
	ExpiresAt      *time.Time
}

type Click struct {
	ID         int64
	LinkID     int64
	ClickedAt  time.Time
	IPHash     *string
	Referrer   *string
	UserAgent  *string
	DeviceType *string
	Browser    *string
	Os         *string
	Country    *string
	City       *string
}

type User struct {
	ID              int64
	PublicID        string
	Email           string
	PasswordHash    *string
	OauthProvider   *string
	OauthProviderID *string
	CreatedAt       time.Time
}

type RefreshToken struct {
	ID        int64
	UserID    int64
	Token     string
	Revoked   bool
	ExpireAt  time.Time
	CreatedAt *time.Time
}
