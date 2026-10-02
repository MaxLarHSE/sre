package repository

import (
	"context"

	"trade/internal/domain"
)

func (r *Repository) CreateListing(ctx context.Context, l domain.Listing, key, requestHash string) (domain.Listing, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Listing{}, err
	}

	defer tx.Rollback(ctx)

	created, err := scanListing(tx.QueryRow(ctx,
		`INSERT INTO listings (game_id, seller_id, item_code, quantity, price)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+listingColumns,
		l.GameID, l.SellerID, l.ItemCode, l.Quantity, l.Price,
	))
	if err != nil {
		return domain.Listing{}, err
	}

	if err := commitWithResult(ctx, tx, l.GameID, key, requestHash, created); err != nil {
		return domain.Listing{}, err
	}
	return created, nil
}
