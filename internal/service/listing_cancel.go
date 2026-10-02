package service

import (
	"context"

	"trade/internal/domain"
)

type CancelListingInput struct {
	ListingID int64
	SellerID  string
}

func (in CancelListingInput) validate() error {
	return firstErr(
		checkListingID(in.ListingID),
		checkCode("seller_id", in.SellerID),
	)
}

func (s *Service) CancelListing(ctx context.Context, gameID, idempotencyKey string, in CancelListingInput) (domain.Listing, error) {
	if err := firstErr(validateIdempotencyKey(idempotencyKey), in.validate()); err != nil {
		return domain.Listing{}, err
	}

	hash, err := requestHash("cancel_listing", in)
	if err != nil {
		return domain.Listing{}, err
	}

	if l, found, err := s.storedListing(ctx, gameID, idempotencyKey, hash); err != nil || found {
		return l, err
	}

	l, err := s.repo.CancelListing(ctx, gameID, domain.Cancellation{
		ListingID: in.ListingID,
		SellerID:  in.SellerID,
	}, idempotencyKey, hash)
	return s.resolveConcurrent(ctx, gameID, idempotencyKey, hash, l, err)
}
