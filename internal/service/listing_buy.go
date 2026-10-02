package service

import (
	"context"

	"trade/internal/domain"
)

type BuyListingInput struct {
	ListingID int64
	BuyerID   string
	Price     int64
}

func (in BuyListingInput) validate() error {
	return firstErr(
		checkListingID(in.ListingID),
		checkCode("buyer_id", in.BuyerID),
		checkPrice(in.Price),
	)
}

func (s *Service) BuyListing(ctx context.Context, gameID, idempotencyKey string, in BuyListingInput) (domain.Listing, error) {
	if err := firstErr(validateIdempotencyKey(idempotencyKey), in.validate()); err != nil {
		return domain.Listing{}, err
	}

	hash, err := requestHash("buy_listing", in)
	if err != nil {
		return domain.Listing{}, err
	}

	if l, found, err := s.storedListing(ctx, gameID, idempotencyKey, hash); err != nil || found {
		return l, err
	}

	l, err := s.repo.BuyListing(ctx, gameID, domain.Purchase{
		ListingID: in.ListingID,
		BuyerID:   in.BuyerID,
		Price:     in.Price,
	}, idempotencyKey, hash)
	return s.resolveConcurrent(ctx, gameID, idempotencyKey, hash, l, err)
}
