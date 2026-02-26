package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/cmmasaba/prototypes/pkg/application/domain"
)

// GetLinkByCode queries links by the given short code.
func (r *Repository) GetLinkByCode(ctx context.Context, code string) (*domain.Link, error) {
	params := map[string]string{
		"code": code,
	}

	stmt, ok := r.statements[searchLinksByCodeQuery]
	if !ok {
		return nil, fmt.Errorf(errPreparedStmtNotFound, searchLinksByCodeQuery)
	}

	var link Link

	err := stmt.GetContext(ctx, &link, params)
	if err != nil {
		return nil, err
	}

	return &domain.Link{
		ID:             link.ID,
		ShortCode:      link.ShortCode,
		OriginalURL:    link.OriginalURL,
		OwnershipToken: link.OwnershipToken,
		CreatedAt:      link.CreatedAt,
		ExpiresAt:      nullTimeToTime(link.ExpiresAt),
	}, nil
}

// GetClicksByClickedAt queries clicks for a given link based on the given time.
func (r *Repository) GetClicksByClickedAt(ctx context.Context, linkID int64, clickedAt time.Time) ([]domain.Click, error) {
	params := map[string]any{
		"link_id":    linkID,
		"clicked_at": clickedAt,
	}

	stmt, ok := r.statements[searchClicksByLinkIDAndClickedAtQuery]
	if !ok {
		return nil, fmt.Errorf(errPreparedStmtNotFound, searchClicksByLinkIDAndClickedAtQuery)
	}

	result := []Click{}

	err := stmt.SelectContext(ctx, &result, params)
	if err != nil {
		return nil, err
	}

	var clicks []domain.Click

	for _, item := range result {
		clicks = append(clicks, domain.Click{
			ID:         item.ID,
			ClickedAt:  item.ClickedAt,
			LinkID:     item.LinkID,
			Referrer:   nullStringToString(item.Referrer),
			IPHash:     nullStringToString(item.IPHash),
			UserAgent:  nullStringToString(item.UserAgent),
			DeviceType: nullStringToString(item.DeviceType),
			OS:         nullStringToString(item.OS),
			Country:    nullStringToString(item.Country),
			City:       nullStringToString(item.City),
			Browser:    nullStringToString(item.Browser),
		})
	}

	return clicks, nil
}

// GetClicksByCountry queries clicks for a given link based on the given country.
func (r *Repository) GetClicksByCountry(ctx context.Context, linkID int64, country *string) ([]domain.Click, error) {
	params := map[string]any{
		"link_id": linkID,
		"country": country,
	}

	stmt, ok := r.statements[searchClicksByLinkIDAndCountryQuery]
	if !ok {
		return nil, fmt.Errorf(errPreparedStmtNotFound, searchClicksByLinkIDAndCountryQuery)
	}

	result := []Click{}

	err := stmt.SelectContext(ctx, &result, params)
	if err != nil {
		return nil, err
	}

	var clicks []domain.Click

	for _, item := range result {
		clicks = append(clicks, domain.Click{
			ID:         item.ID,
			ClickedAt:  item.ClickedAt,
			LinkID:     item.LinkID,
			Referrer:   nullStringToString(item.Referrer),
			IPHash:     nullStringToString(item.IPHash),
			UserAgent:  nullStringToString(item.UserAgent),
			DeviceType: nullStringToString(item.DeviceType),
			OS:         nullStringToString(item.OS),
			Country:    nullStringToString(item.Country),
			City:       nullStringToString(item.City),
			Browser:    nullStringToString(item.Browser),
		})
	}

	return clicks, nil
}
