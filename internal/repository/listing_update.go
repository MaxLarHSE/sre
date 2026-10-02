package repository

import (
	"context"

	"trade/internal/domain"
)

func (r *Repository) UpdateListingPrice(ctx context.Context, gameID string, c domain.PriceChange, key, requestHash string) (domain.Listing, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Listing{}, err
	}

	defer tx.Rollback(ctx)

	l, err := lockListing(ctx, tx, gameID, c.ListingID)
	if err != nil {
		return domain.Listing{}, err
	}
	if err := l.CanChangePrice(c); err != nil {
		return domain.Listing{}, err
	}

	updated, err := scanListing(tx.QueryRow(ctx,
		`UPDATE listings SET price = $2
		WHERE id = $1
		RETURNING `+listingColumns,
		l.ID, c.Price,
	))
	if err != nil {
		return domain.Listing{}, err
	}

	if err := commitWithResult(ctx, tx, gameID, key, requestHash, updated); err != nil {
		return domain.Listing{}, err
	}
	return updated, nil
}
