package server

import (
	"net/http"
	"strconv"

	"trade/internal/domain"
	"trade/internal/service"
)

type listingsPageResponse struct {
	Items      []listingResponse `json:"items"`
	NextCursor *string           `json:"next_cursor"`
}

func (s *Server) listListings(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	params := service.ListListingsParams{
		ItemCode: q.Get("item_code"),
		SellerID: q.Get("seller_id"),
		Cursor:   q.Get("cursor"),
	}
	if v := q.Get("limit"); v != "" {
		limit, err := strconv.Atoi(v)
		if err != nil || limit < 1 {
			s.handleError(w, r, &domain.ValidationError{Field: "limit", Message: "must be an integer between 1 and 100"})
			return
		}
		params.Limit = limit
	}

	page, err := s.svc.ListListings(r.Context(), gameIDFrom(r.Context()), params)
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	resp := listingsPageResponse{Items: make([]listingResponse, 0, len(page.Items))}
	for _, l := range page.Items {
		resp.Items = append(resp.Items, toListingResponse(l))
	}
	if page.NextCursor != "" {
		resp.NextCursor = &page.NextCursor
	}

	writeJSON(w, http.StatusOK, resp)
}
