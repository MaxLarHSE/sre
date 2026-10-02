package repository

import (
	"context"
	"strconv"
	"strings"

	"trade/internal/domain"
)

func (r *Repository) ListActiveListings(ctx context.Context, f domain.ListingFilter) ([]domain.Listing, error) {
	var (
		where = []string{"game_id = $1", "status = 'active'"}
		args  = []any{f.GameID}
	)
	addArg := func(cond string, v any) {
		args = append(args, v)
		where = append(where, cond+" $"+strconv.Itoa(len(args)))
	}

	if f.ItemCode != "" {
		addArg("item_code =", f.ItemCode)
	}
	if f.SellerID != "" {
		addArg("seller_id =", f.SellerID)
	}
	if f.BeforeID > 0 {
		addArg("id <", f.BeforeID)
	}
	args = append(args, f.Limit)

	query := `SELECT ` + listingColumns + `
		FROM listings
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY id DESC
		LIMIT $` + strconv.Itoa(len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []domain.Listing
	for rows.Next() {
		l, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		listings = append(listings, l)
	}
	return listings, rows.Err()
}
