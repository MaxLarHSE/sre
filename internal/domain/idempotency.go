package domain

import "errors"

var (
	ErrIdempotencyKeyExists = errors.New("idempotency key already exists")

	ErrIdempotencyKeyReused = errors.New("idempotency key reused with different request")
)

type IdempotencyRecord struct {
	GameID      string
	Key         string
	RequestHash string
	Response    []byte
}
