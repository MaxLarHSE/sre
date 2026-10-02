package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"trade/internal/domain"
)

func (r *Repository) GameIDByAPIKey(ctx context.Context, apiKey string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx, `SELECT id::text FROM games WHERE api_key = $1`, apiKey).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return id, nil
}
