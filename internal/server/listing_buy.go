package server

import (
	"net/http"

	"trade/internal/service"
)

type buyListingRequest struct {
	BuyerID string `json:"buyer_id"`
	Price   int64  `json:"price"`
}

type buyListingResponse struct {
	ListingID int64        `json:"listing_id"`
	BuyerID   string       `json:"buyer_id"`
	Item      itemResponse `json:"item"`
	SellerID  string       `json:"seller_id"`
	Amount    int64        `json:"amount"`
}

func (s *Server) buyListing(w http.ResponseWriter, r *http.Request) {
	listingID, err := listingIDFromPath(r)
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	var req buyListingRequest
	if err := readJSON(w, r, &req); err != nil {
		s.handleError(w, r, err)
		return
	}

	listing, err := s.svc.BuyListing(r.Context(), gameIDFrom(r.Context()), r.Header.Get("Idempotency-Key"), service.BuyListingInput{
		ListingID: listingID,
		BuyerID:   req.BuyerID,
		Price:     req.Price,
	})
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, buyListingResponse{
		ListingID: listing.ID,
		BuyerID:   listing.BuyerID,
		Item:      toItemResponse(listing),
		SellerID:  listing.SellerID,
		Amount:    listing.Price,
	})
}
