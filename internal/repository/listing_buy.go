package repository

import (
	"context"

	"trade/internal/domain"
)

func (r *Repository) BuyListing(ctx context.Context, gameID string, p domain.Purchase, key, requestHash string) (domain.Listing, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Listing{}, err
	}

	defer tx.Rollback(ctx)

	l, err := lockListing(ctx, tx, gameID, p.ListingID)
	if err != nil {
		return domain.Listing{}, err
	}
	if err := l.CanBeBought(p); err != nil {
		return domain.Listing{}, err
	}

	sold, err := scanListing(tx.QueryRow(ctx,
		`UPDATE listings SET status = 'sold', buyer_id = $2, closed_at = now()
		WHERE id = $1
		RETURNING `+listingColumns,
		l.ID, p.BuyerID,
	))
	if err != nil {
		return domain.Listing{}, err
	}

	if err := commitWithResult(ctx, tx, gameID, key, requestHash, sold); err != nil {
		return domain.Listing{}, err
	}
	return sold, nil
}
