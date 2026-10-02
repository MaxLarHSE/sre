package server

import (
	"net/http"

	"trade/internal/service"
)

type createListingRequest struct {
	SellerID string `json:"seller_id"`
	ItemCode string `json:"item_code"`
	Quantity int    `json:"quantity"`
	Price    int64  `json:"price"`
}

func (s *Server) createListing(w http.ResponseWriter, r *http.Request) {
	var req createListingRequest
	if err := readJSON(w, r, &req); err != nil {
		s.handleError(w, r, err)
		return
	}

	listing, err := s.svc.CreateListing(r.Context(), gameIDFrom(r.Context()), r.Header.Get("Idempotency-Key"), service.CreateListingInput{
		SellerID: req.SellerID,
		ItemCode: req.ItemCode,
		Quantity: req.Quantity,
		Price:    req.Price,
	})
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, toListingResponse(listing))
}
