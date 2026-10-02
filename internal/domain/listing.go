package domain

import "time"

type ListingStatus string

const (
	ListingActive    ListingStatus = "active"
	ListingSold      ListingStatus = "sold"
	ListingCancelled ListingStatus = "cancelled"
)

type Listing struct {
	ID        int64
	GameID    string
	SellerID  string
	ItemCode  string
	Quantity  int
	Price     int64
	Status    ListingStatus
	BuyerID   string
	CreatedAt time.Time
}

type ListingFilter struct {
	GameID   string
	ItemCode string
	SellerID string
	BeforeID int64
	Limit    int
}

type Purchase struct {
	ListingID int64
	BuyerID   string
	Price     int64
}

type Cancellation struct {
	ListingID int64
	SellerID  string
}

type PriceChange struct {
	ListingID int64
	SellerID  string
	Price     int64
}

func (l Listing) CanBeBought(p Purchase) error {
	switch {
	case l.Status != ListingActive:
		return ErrListingNotActive
	case l.SellerID == p.BuyerID:
		return ErrOwnListing
	case l.Price != p.Price:
		return ErrPriceMismatch
	}
	return nil
}

func (l Listing) CanBeCancelled(c Cancellation) error {
	return l.canBeChangedBySeller(c.SellerID)
}

func (l Listing) CanChangePrice(c PriceChange) error {
	return l.canBeChangedBySeller(c.SellerID)
}

func (l Listing) canBeChangedBySeller(sellerID string) error {
	switch {
	case l.Status != ListingActive:
		return ErrListingNotActive
	case l.SellerID != sellerID:
		return ErrNotListingOwner
	}
	return nil
}
