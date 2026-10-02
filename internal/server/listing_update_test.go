package server

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"trade/internal/domain"
	"trade/internal/service"
)

const validUpdateBody = `{"seller_id":"vasya","price":150}`

func TestUpdateListing_OK(t *testing.T) {
	svc := &fakeService{}

	rec := doRequest(newTestHandler(svc), http.MethodPatch, "/v1/listings/7", validUpdateBody, "good-key", "upd-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	want := service.UpdateListingPriceInput{ListingID: 7, SellerID: "vasya", Price: 150}
	if svc.lastUpdate != want || svc.lastKey != "upd-1" || svc.lastGameID != "game-1" {
		t.Fatalf("service called with game=%q key=%q input=%+v", svc.lastGameID, svc.lastKey, svc.lastUpdate)
	}

	var body listingResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.ID != 7 || body.Price != 150 || body.Status != "active" {
		t.Fatalf("unexpected body: %s", rec.Body)
	}
}

func TestUpdateListing_Errors(t *testing.T) {
	cases := []struct {
		name   string
		target string
		body   string
		apiKey string
		svcErr error
		status int
		code   string
	}{
		{"no api key", "/v1/listings/7", validUpdateBody, "", nil, http.StatusUnauthorized, "unauthorized"},
		{"non-numeric id", "/v1/listings/abc", validUpdateBody, "good-key", nil, http.StatusBadRequest, "validation_error"},
		{"unknown field", "/v1/listings/7", `{"seller_id":"vasya","price":150,"quantity":3}`, "good-key", nil, http.StatusBadRequest, "validation_error"},
		{"not found", "/v1/listings/7", validUpdateBody, "good-key", domain.ErrNotFound, http.StatusNotFound, "not_found"},
		{"not owner", "/v1/listings/7", validUpdateBody, "good-key", domain.ErrNotListingOwner, http.StatusForbidden, "not_owner"},
		{"already closed", "/v1/listings/7", validUpdateBody, "good-key", domain.ErrListingNotActive, http.StatusConflict, "listing_not_active"},
		{"key reused", "/v1/listings/7", validUpdateBody, "good-key", domain.ErrIdempotencyKeyReused, http.StatusUnprocessableEntity, "idempotency_key_reused"},
		{"internal", "/v1/listings/7", validUpdateBody, "good-key", io.ErrUnexpectedEOF, http.StatusInternalServerError, "internal_error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doRequest(newTestHandler(&fakeService{err: tc.svcErr}), http.MethodPatch, tc.target, tc.body, tc.apiKey, "upd-1")

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
