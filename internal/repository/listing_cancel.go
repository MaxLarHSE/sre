package repository

import (
	"context"

	"trade/internal/domain"
)

func (r *Repository) CancelListing(ctx context.Context, gameID string, c domain.Cancellation, key, requestHash string) (domain.Listing, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Listing{}, err
	}

	defer tx.Rollback(ctx)

	l, err := lockListing(ctx, tx, gameID, c.ListingID)
	if err != nil {
		return domain.Listing{}, err
	}
	if err := l.CanBeCancelled(c); err != nil {
		return domain.Listing{}, err
	}

	cancelled, err := scanListing(tx.QueryRow(ctx,
		`UPDATE listings SET status = 'cancelled', closed_at = now()
		WHERE id = $1
		RETURNING `+listingColumns,
		l.ID,
	))
	if err != nil {
		return domain.Listing{}, err
	}

	if err := commitWithResult(ctx, tx, gameID, key, requestHash, cancelled); err != nil {
		return domain.Listing{}, err
	}
	return cancelled, nil
}
