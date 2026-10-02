package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

	"trade/internal/domain"
)

const listingColumns = `id, game_id::text, seller_id, item_code, quantity, price, status::text, coalesce(buyer_id, ''), created_at`

func scanListing(row pgx.Row) (domain.Listing, error) {
	var l domain.Listing
	err := row.Scan(&l.ID, &l.GameID, &l.SellerID, &l.ItemCode, &l.Quantity, &l.Price, &l.Status, &l.BuyerID, &l.CreatedAt)
	return l, err
}

func lockListing(ctx context.Context, tx pgx.Tx, gameID string, listingID int64) (domain.Listing, error) {
	l, err := scanListing(tx.QueryRow(ctx,
		`SELECT `+listingColumns+` FROM listings WHERE id = $1 AND game_id = $2 FOR UPDATE`,
		listingID, gameID,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Listing{}, domain.ErrNotFound
	}
	return l, err
}

func commitWithResult(ctx context.Context, tx pgx.Tx, gameID, key, requestHash string, l domain.Listing) error {
	response, err := json.Marshal(l)
	if err != nil {
		return err
	}

	err = saveIdempotencyRecord(ctx, tx, domain.IdempotencyRecord{
		GameID:      gameID,
		Key:         key,
		RequestHash: requestHash,
		Response:    response,
	})
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
