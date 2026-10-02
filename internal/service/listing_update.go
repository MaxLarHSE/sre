package service

import (
	"context"

	"trade/internal/domain"
)

type UpdateListingPriceInput struct {
	ListingID int64
	SellerID  string
	Price     int64
}

func (in UpdateListingPriceInput) validate() error {
	return firstErr(
		checkListingID(in.ListingID),
		checkCode("seller_id", in.SellerID),
		checkPrice(in.Price),
	)
}

func (s *Service) UpdateListingPrice(ctx context.Context, gameID, idempotencyKey string, in UpdateListingPriceInput) (domain.Listing, error) {
	if err := firstErr(validateIdempotencyKey(idempotencyKey), in.validate()); err != nil {
		return domain.Listing{}, err
	}

	hash, err := requestHash("update_listing_price", in)
	if err != nil {
		return domain.Listing{}, err
	}

	if l, found, err := s.storedListing(ctx, gameID, idempotencyKey, hash); err != nil || found {
		return l, err
	}

	l, err := s.repo.UpdateListingPrice(ctx, gameID, domain.PriceChange{
		ListingID: in.ListingID,
		SellerID:  in.SellerID,
		Price:     in.Price,
	}, idempotencyKey, hash)
	return s.resolveConcurrent(ctx, gameID, idempotencyKey, hash, l, err)
}
