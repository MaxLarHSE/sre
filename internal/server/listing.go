package server

import (
	"net/http"
	"strconv"
	"time"

	"trade/internal/domain"
)

type listingResponse struct {
	ID        int64     `json:"id"`
	SellerID  string    `json:"seller_id"`
	ItemCode  string    `json:"item_code"`
	Quantity  int       `json:"quantity"`
	Price     int64     `json:"price"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func toListingResponse(l domain.Listing) listingResponse {
	return listingResponse{
		ID:        l.ID,
		SellerID:  l.SellerID,
		ItemCode:  l.ItemCode,
		Quantity:  l.Quantity,
		Price:     l.Price,
		Status:    string(l.Status),
		CreatedAt: l.CreatedAt,
	}
}

type itemResponse struct {
	Code     string `json:"code"`
	Quantity int    `json:"quantity"`
}

func toItemResponse(l domain.Listing) itemResponse {
	return itemResponse{Code: l.ItemCode, Quantity: l.Quantity}
}

func listingIDFromPath(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return 0, &domain.ValidationError{Field: "id", Message: "must be a positive integer"}
	}
	return id, nil
}
