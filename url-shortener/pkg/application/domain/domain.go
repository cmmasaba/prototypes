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
