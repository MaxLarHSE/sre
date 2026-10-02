package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"trade/internal/domain"
)

const maxIdempotencyKeyLen = 255

func validateIdempotencyKey(key string) error {
	if key == "" {
		return &domain.ValidationError{Field: "Idempotency-Key", Message: "header is required"}
	}
	if len(key) > maxIdempotencyKeyLen {
		return &domain.ValidationError{Field: "Idempotency-Key", Message: "must be at most 255 characters"}
	}
	return nil
}

func requestHash(operation string, input any) (string, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(append([]byte(operation+"\n"), body...))
	return hex.EncodeToString(sum[:]), nil
}

func (s *Service) storedListing(ctx context.Context, gameID, key, hash string) (l domain.Listing, found bool, err error) {
	rec, err := s.repo.GetIdempotencyRecord(ctx, gameID, key)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.Listing{}, false, nil
	}
	if err != nil {
		return domain.Listing{}, false, err
	}

	if rec.RequestHash != hash {
		return domain.Listing{}, true, domain.ErrIdempotencyKeyReused
	}
	if err := json.Unmarshal(rec.Response, &l); err != nil {
		return domain.Listing{}, true, fmt.Errorf("unmarshal stored listing: %w", err)
	}
	return l, true, nil
}

func (s *Service) resolveConcurrent(ctx context.Context, gameID, key, hash string, l domain.Listing, err error) (domain.Listing, error) {
	if errors.Is(err, domain.ErrIdempotencyKeyExists) || errors.Is(err, domain.ErrListingNotActive) {
		if stored, found, sErr := s.storedListing(ctx, gameID, key, hash); sErr != nil || found {
			return stored, sErr
		}
	}
	return l, err
}
