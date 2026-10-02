package service

import (
	"context"
	"errors"

	"trade/internal/domain"
)

type Repository interface {
	Ping(ctx context.Context) error
	GameIDByAPIKey(ctx context.Context, apiKey string) (string, error)

	ListActiveListings(ctx context.Context, f domain.ListingFilter) ([]domain.Listing, error)

	CreateListing(ctx context.Context, l domain.Listing, idempotencyKey, requestHash string) (domain.Listing, error)

	BuyListing(ctx context.Context, gameID string, p domain.Purchase, idempotencyKey, requestHash string) (domain.Listing, error)
	CancelListing(ctx context.Context, gameID string, c domain.Cancellation, idempotencyKey, requestHash string) (domain.Listing, error)
	UpdateListingPrice(ctx context.Context, gameID string, c domain.PriceChange, idempotencyKey, requestHash string) (domain.Listing, error)

	GetIdempotencyRecord(ctx context.Context, gameID, key string) (domain.IdempotencyRecord, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Ready(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *Service) Authenticate(ctx context.Context, apiKey string) (string, error) {
	if apiKey == "" {
		return "", domain.ErrUnauthorized
	}

	gameID, err := s.repo.GameIDByAPIKey(ctx, apiKey)
	if errors.Is(err, domain.ErrNotFound) {
		return "", domain.ErrUnauthorized
	}
	return gameID, err
}
