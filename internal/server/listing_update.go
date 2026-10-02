package server

import (
	"net/http"

	"trade/internal/service"
)

type updateListingRequest struct {
	SellerID string `json:"seller_id"`
	Price    int64  `json:"price"`
}

func (s *Server) updateListing(w http.ResponseWriter, r *http.Request) {
	listingID, err := listingIDFromPath(r)
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	var req updateListingRequest
	if err := readJSON(w, r, &req); err != nil {
		s.handleError(w, r, err)
		return
	}

	listing, err := s.svc.UpdateListingPrice(r.Context(), gameIDFrom(r.Context()), r.Header.Get("Idempotency-Key"), service.UpdateListingPriceInput{
		ListingID: listingID,
		SellerID:  req.SellerID,
		Price:     req.Price,
	})
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toListingResponse(listing))
}
