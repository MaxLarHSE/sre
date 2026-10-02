package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"trade/internal/domain"
	"trade/internal/service"
)

const validBody = `{"seller_id":"vasya","item_code":"iron_sword","quantity":1,"price":100}`

func TestCreateListing_Created(t *testing.T) {
	svc := &fakeService{}

	rec := doPost(newTestHandler(svc), "/v1/listings", validBody, "good-key", "key-1")

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	wantInput := service.CreateListingInput{SellerID: "vasya", ItemCode: "iron_sword", Quantity: 1, Price: 100}
	if svc.lastInput != wantInput || svc.lastKey != "key-1" || svc.lastGameID != "game-1" {
		t.Fatalf("service called with game=%q key=%q input=%+v", svc.lastGameID, svc.lastKey, svc.lastInput)
	}

	var body listingResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.ID != 42 || body.Status != "active" || body.SellerID != "vasya" {
		t.Fatalf("unexpected body: %s", rec.Body)
	}
}

func TestCreateListing_Errors(t *testing.T) {
	cases := []struct {
		name   string
		body   string
		apiKey string
		svcErr error
		status int
		code   string
	}{
		{"no api key", validBody, "", nil, http.StatusUnauthorized, "unauthorized"},
		{"broken json", `{"seller_id":`, "good-key", nil, http.StatusBadRequest, "validation_error"},
		{"wrong type", `{"seller_id":"vasya","item_code":"x","quantity":"one","price":100}`, "good-key", nil, http.StatusBadRequest, "validation_error"},
		{"unknown field", `{"seller_id":"vasya","item_code":"x","quantity":1,"price":100,"buyer_id":"petya"}`, "good-key", nil, http.StatusBadRequest, "validation_error"},
		{"two objects", validBody + validBody, "good-key", nil, http.StatusBadRequest, "validation_error"},
		{"too large", `{"seller_id":"` + strings.Repeat("a", maxBodyBytes) + `"}`, "good-key", nil, http.StatusBadRequest, "validation_error"},
		{"service validation", validBody, "good-key", &domain.ValidationError{Field: "price", Message: "must be positive"}, http.StatusBadRequest, "validation_error"},
		{"key reused", validBody, "good-key", domain.ErrIdempotencyKeyReused, http.StatusUnprocessableEntity, "idempotency_key_reused"},
		{"internal", validBody, "good-key", io.ErrUnexpectedEOF, http.StatusInternalServerError, "internal_error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doPost(newTestHandler(&fakeService{err: tc.svcErr}), "/v1/listings", tc.body, tc.apiKey, "key-1")

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
