package server

import (
	"net/http"

	"trade/internal/service"
)

type cancelListingRequest struct {
	SellerID string `json:"seller_id"`
}

type cancelListingResponse struct {
	ListingID int64        `json:"listing_id"`
	SellerID  string       `json:"seller_id"`
	Item      itemResponse `json:"item"`
}

func (s *Server) cancelListing(w http.ResponseWriter, r *http.Request) {
	listingID, err := listingIDFromPath(r)
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	var req cancelListingRequest
	if err := readJSON(w, r, &req); err != nil {
		s.handleError(w, r, err)
		return
	}

	listing, err := s.svc.CancelListing(r.Context(), gameIDFrom(r.Context()), r.Header.Get("Idempotency-Key"), service.CancelListingInput{
		ListingID: listingID,
		SellerID:  req.SellerID,
	})
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, cancelListingResponse{
		ListingID: listing.ID,
		SellerID:  listing.SellerID,
		Item:      toItemResponse(listing),
	})
}
