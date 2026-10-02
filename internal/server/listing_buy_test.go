package server

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"trade/internal/domain"
	"trade/internal/service"
)

const validBuyBody = `{"buyer_id":"petya","price":100}`

func TestBuyListing_OK(t *testing.T) {
	svc := &fakeService{}

	rec := doPost(newTestHandler(svc), "/v1/listings/7/buy", validBuyBody, "good-key", "buy-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	want := service.BuyListingInput{ListingID: 7, BuyerID: "petya", Price: 100}
	if svc.lastBuy != want || svc.lastKey != "buy-1" || svc.lastGameID != "game-1" {
		t.Fatalf("service called with game=%q key=%q input=%+v", svc.lastGameID, svc.lastKey, svc.lastBuy)
	}

	var body buyListingResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	wantBody := buyListingResponse{
		ListingID: 7,
		BuyerID:   "petya",
		Item:      itemResponse{Code: "iron_sword", Quantity: 1},
		SellerID:  "vasya",
		Amount:    100,
	}
	if body != wantBody {
		t.Fatalf("body = %+v, want %+v", body, wantBody)
	}
}

func TestBuyListing_Errors(t *testing.T) {
	cases := []struct {
		name   string
		target string
		body   string
		svcErr error
		status int
		code   string
	}{
		{"non-numeric id", "/v1/listings/abc/buy", validBuyBody, nil, http.StatusBadRequest, "validation_error"},
		{"unknown field", "/v1/listings/7/buy", `{"buyer_id":"petya","price":100,"seller_id":"x"}`, nil, http.StatusBadRequest, "validation_error"},
		{"not found", "/v1/listings/7/buy", validBuyBody, domain.ErrNotFound, http.StatusNotFound, "not_found"},
		{"already sold", "/v1/listings/7/buy", validBuyBody, domain.ErrListingNotActive, http.StatusConflict, "listing_not_active"},
		{"own listing", "/v1/listings/7/buy", validBuyBody, domain.ErrOwnListing, http.StatusConflict, "own_listing"},
		{"price changed", "/v1/listings/7/buy", validBuyBody, domain.ErrPriceMismatch, http.StatusConflict, "price_mismatch"},
		{"key reused", "/v1/listings/7/buy", validBuyBody, domain.ErrIdempotencyKeyReused, http.StatusUnprocessableEntity, "idempotency_key_reused"},
		{"internal", "/v1/listings/7/buy", validBuyBody, io.ErrUnexpectedEOF, http.StatusInternalServerError, "internal_error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doPost(newTestHandler(&fakeService{err: tc.svcErr}), tc.target, tc.body, "good-key", "buy-1")

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tc.status, rec.Body)
			}
			var body errorBody
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Error.Code != tc.code {
				t.Fatalf("code = %q, want %q", body.Error.Code, tc.code)
			}
		})
	}
}

func TestBuyListing_RequiresAPIKey(t *testing.T) {
	rec := doPost(newTestHandler(&fakeService{}), "/v1/listings/7/buy", validBuyBody, "", "buy-1")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}
