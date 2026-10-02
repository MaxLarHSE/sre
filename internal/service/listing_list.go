package service

import (
	"context"
	"encoding/base64"
	"errors"
	"strconv"

	"trade/internal/domain"
)

const (
	defaultListingsLimit = 20
	maxListingsLimit     = 100
)

type ListListingsParams struct {
	ItemCode string
	SellerID string
	Cursor   string
	Limit    int
}

type ListingsPage struct {
	Items      []domain.Listing
	NextCursor string
}

func (s *Service) ListListings(ctx context.Context, gameID string, p ListListingsParams) (ListingsPage, error) {
	limit := p.Limit
	if limit == 0 {
		limit = defaultListingsLimit
	}
	if limit < 1 || limit > maxListingsLimit {
		return ListingsPage{}, &domain.ValidationError{Field: "limit", Message: "must be between 1 and 100"}
	}

	var beforeID int64
	if p.Cursor != "" {
		id, err := decodeCursor(p.Cursor)
		if err != nil {
			return ListingsPage{}, &domain.ValidationError{Field: "cursor", Message: "invalid cursor"}
		}
		beforeID = id
	}

	items, err := s.repo.ListActiveListings(ctx, domain.ListingFilter{
		GameID:   gameID,
		ItemCode: p.ItemCode,
		SellerID: p.SellerID,
		BeforeID: beforeID,
		Limit:    limit + 1,
	})
	if err != nil {
		return ListingsPage{}, err
	}

	page := ListingsPage{Items: items}
	if len(items) > limit {
		page.Items = items[:limit]
		page.NextCursor = encodeCursor(items[limit-1].ID)
	}
	return page, nil
}

func encodeCursor(id int64) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.FormatInt(id, 10)))
}

func decodeCursor(cursor string) (int64, error) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return 0, err
	}
	id, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, errors.New("cursor id must be positive")
	}
	return id, nil
}
