package repository

import (
	"context"
	"time"
)

// GetLinkByCode queries links by the given short code.
func (r *Repository) GetLinkByCode(ctx context.Context, code string) {}

// GetClicksByClickedAt queries clicks for a given link based on the given time.
func (r *Repository) GetClicksByClickedAt(ctx context.Context, linkID int64, clickedAt time.Time) {}

// GetClicksByCountry queries clicks for a given link based on the given country.
func (r *Repository) GetClicksByCountry(ctx context.Context, linkID int64, country *string) {}
