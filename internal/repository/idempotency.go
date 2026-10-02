package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"trade/internal/domain"
)

const uniqueViolation = "23505"

func (r *Repository) GetIdempotencyRecord(ctx context.Context, gameID, key string) (domain.IdempotencyRecord, error) {
	rec := domain.IdempotencyRecord{GameID: gameID, Key: key}
	err := r.pool.QueryRow(ctx,
		`SELECT request_hash, response FROM idempotency_keys WHERE game_id = $1 AND key = $2`,
		gameID, key,
	).Scan(&rec.RequestHash, &rec.Response)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.IdempotencyRecord{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.IdempotencyRecord{}, err
	}
	return rec, nil
}

func saveIdempotencyRecord(ctx context.Context, tx pgx.Tx, rec domain.IdempotencyRecord) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO idempotency_keys (game_id, key, request_hash, response) VALUES ($1, $2, $3, $4)`,
		rec.GameID, rec.Key, rec.RequestHash, rec.Response,
	)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return domain.ErrIdempotencyKeyExists
	}
	return err
}
