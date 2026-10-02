package service

import (
	"context"
	"math"

	"trade/internal/domain"
)

type CreateListingInput struct {
	SellerID string
	ItemCode string
	Quantity int
	Price    int64
}

func (in CreateListingInput) validate() error {
	var quantityErr error
	if in.Quantity < 1 || in.Quantity > math.MaxInt32 {
		quantityErr = &domain.ValidationError{Field: "quantity", Message: "must be between 1 and 2147483647"}
	}
	return firstErr(
		checkCode("seller_id", in.SellerID),
		checkCode("item_code", in.ItemCode),
		quantityErr,
		checkPrice(in.Price),
	)
}

func (s *Service) CreateListing(ctx context.Context, gameID, idempotencyKey string, in CreateListingInput) (domain.Listing, error) {
	if err := firstErr(validateIdempotencyKey(idempotencyKey), in.validate()); err != nil {
		return domain.Listing{}, err
	}

	hash, err := requestHash("create_listing", in)
	if err != nil {
		return domain.Listing{}, err
	}

	if l, found, err := s.storedListing(ctx, gameID, idempotencyKey, hash); err != nil || found {
		return l, err
	}

	l, err := s.repo.CreateListing(ctx, domain.Listing{
		GameID:   gameID,
		SellerID: in.SellerID,
		ItemCode: in.ItemCode,
		Quantity: in.Quantity,
		Price:    in.Price,
	}, idempotencyKey, hash)
	return s.resolveConcurrent(ctx, gameID, idempotencyKey, hash, l, err)
}
