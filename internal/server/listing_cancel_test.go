package server

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"trade/internal/domain"
	"trade/internal/service"
)

const validCancelBody = `{"seller_id":"vasya"}`

func TestCancelListing_OK(t *testing.T) {
	svc := &fakeService{}

	rec := doPost(newTestHandler(svc), "/v1/listings/7/cancel", validCancelBody, "good-key", "cancel-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	want := service.CancelListingInput{ListingID: 7, SellerID: "vasya"}
	if svc.lastCancel != want || svc.lastKey != "cancel-1" || svc.lastGameID != "game-1" {
		t.Fatalf("service called with game=%q key=%q input=%+v", svc.lastGameID, svc.lastKey, svc.lastCancel)
	}

	var body cancelListingResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	wantBody := cancelListingResponse{
		ListingID: 7,
		SellerID:  "vasya",
		Item:      itemResponse{Code: "wolf_pelt", Quantity: 50},
	}
	if body != wantBody {
		t.Fatalf("body = %+v, want %+v", body, wantBody)
	}
}

func TestCancelListing_Errors(t *testing.T) {
	cases := []struct {
		name   string
		target string
		body   string
		apiKey string
		svcErr error
		status int
		code   string
	}{
		{"no api key", "/v1/listings/7/cancel", validCancelBody, "", nil, http.StatusUnauthorized, "unauthorized"},
		{"non-numeric id", "/v1/listings/abc/cancel", validCancelBody, "good-key", nil, http.StatusBadRequest, "validation_error"},
		{"broken json", "/v1/listings/7/cancel", `{"seller_id":`, "good-key", nil, http.StatusBadRequest, "validation_error"},
		{"not found", "/v1/listings/7/cancel", validCancelBody, "good-key", domain.ErrNotFound, http.StatusNotFound, "not_found"},
		{"not owner", "/v1/listings/7/cancel", validCancelBody, "good-key", domain.ErrNotListingOwner, http.StatusForbidden, "not_owner"},
		{"already closed", "/v1/listings/7/cancel", validCancelBody, "good-key", domain.ErrListingNotActive, http.StatusConflict, "listing_not_active"},
		{"key reused", "/v1/listings/7/cancel", validCancelBody, "good-key", domain.ErrIdempotencyKeyReused, http.StatusUnprocessableEntity, "idempotency_key_reused"},
		{"internal", "/v1/listings/7/cancel", validCancelBody, "good-key", io.ErrUnexpectedEOF, http.StatusInternalServerError, "internal_error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doPost(newTestHandler(&fakeService{err: tc.svcErr}), tc.target, tc.body, tc.apiKey, "cancel-1")

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
