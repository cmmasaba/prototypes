package repository

import (
	"context"
	"time"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

// SaveClickData saves the click data for a link to the database.
func (r *Repository) SaveClickData(ctx context.Context, data domain.Click) (*domain.Click, error) {
	click, err := r.db.SaveNewClick(ctx, sqlc.SaveNewClickParams{
		ClickedAt:  pgtype.Timestamptz{Time: data.ClickedAt, Valid: true},
		LinkID:     data.LinkID,
		IpHash:     *data.IPHash,
		Referrer:   stringToPgtypeText(data.Referrer),
		UserAgent:  stringToPgtypeText(data.UserAgent),
		DeviceType: stringToPgtypeText(data.DeviceType),
		Browser:    stringToPgtypeText(data.Browser),
		Os:         stringToPgtypeText(data.Os),
		Country:    stringToPgtypeText(data.Country),
		City:       stringToPgtypeText(data.City),
	})
	if err != nil {
		return nil, err
	}

	return &domain.Click{
		ID:         click.ID,
		LinkID:     click.LinkID,
		ClickedAt:  click.ClickedAt.Time,
		IPHash:     &click.IpHash,
		Referrer:   pgtypeTextToString(click.Referrer),
		UserAgent:  pgtypeTextToString(click.UserAgent),
		DeviceType: pgtypeTextToString(click.DeviceType),
		Browser:    pgtypeTextToString(click.Browser),
		Os:         pgtypeTextToString(click.Os),
		Country:    pgtypeTextToString(click.Country),
		City:       pgtypeTextToString(click.City),
	}, nil
}

// GetClicksByLinkIDAndClickedAt queries clicks for a given link based on the given time.
func (r *Repository) GetClicksByLinkIDAndClickedAt(ctx context.Context, linkID int64, clickedAt time.Time) ([]domain.Click, error) {
	result, err := r.db.GetClicksByLinkIDAndClickedAt(ctx, sqlc.GetClicksByLinkIDAndClickedAtParams{
		LinkID:    linkID,
		ClickedAt: pgtype.Timestamptz{Time: clickedAt, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	var clicks []domain.Click

	for _, item := range result {
		clicks = append(clicks, domain.Click{
			ID:         item.ID,
			ClickedAt:  item.ClickedAt.Time,
			LinkID:     item.LinkID,
			Referrer:   pgtypeTextToString(item.Referrer),
			IPHash:     &item.IpHash,
			UserAgent:  pgtypeTextToString(item.UserAgent),
			DeviceType: pgtypeTextToString(item.DeviceType),
			Os:         pgtypeTextToString(item.Os),
			Country:    pgtypeTextToString(item.Country),
			City:       pgtypeTextToString(item.City),
			Browser:    pgtypeTextToString(item.Browser),
		})
	}

	return clicks, nil
}

// GetClicksByLinkIDAndCountry queries clicks for a given link based on the given country.
func (r *Repository) GetClicksByLinkIDAndCountry(ctx context.Context, linkID int64, country *string) ([]domain.Click, error) {
	result, err := r.db.GetClicksByLinkIDAndCountry(ctx, sqlc.GetClicksByLinkIDAndCountryParams{
		LinkID:  linkID,
		Country: stringToPgtypeText(country),
	})
	if err != nil {
		return nil, err
	}

	var clicks []domain.Click

	for _, item := range result {
		clicks = append(clicks, domain.Click{
			ID:         item.ID,
			ClickedAt:  item.ClickedAt.Time,
			LinkID:     item.LinkID,
			Referrer:   pgtypeTextToString(item.Referrer),
			IPHash:     &item.IpHash,
			UserAgent:  pgtypeTextToString(item.UserAgent),
			DeviceType: pgtypeTextToString(item.DeviceType),
			Os:         pgtypeTextToString(item.Os),
			Country:    pgtypeTextToString(item.Country),
			City:       pgtypeTextToString(item.City),
			Browser:    pgtypeTextToString(item.Browser),
		})
	}

	return clicks, nil
}
