package server

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"trade/internal/domain"
	"trade/internal/service"
)

func TestListListings_OK(t *testing.T) {
	svc := &fakeService{page: service.ListingsPage{
		Items: []domain.Listing{{
			ID: 7, SellerID: "vasya", ItemCode: "iron_sword", Quantity: 1, Price: 100,
			Status: domain.ListingActive, CreatedAt: time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC),
		}},
		NextCursor: "Nw",
	}}

	rec := doGet(newTestHandler(svc), "/v1/listings?item_code=iron_sword&seller_id=vasya&limit=1&cursor=OA", "good-key")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	want := service.ListListingsParams{ItemCode: "iron_sword", SellerID: "vasya", Cursor: "OA", Limit: 1}
	if svc.lastParams != want || svc.lastGameID != "game-1" {
		t.Fatalf("service called with game=%q params=%+v", svc.lastGameID, svc.lastParams)
	}

	var body listingsPageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Items) != 1 || body.Items[0].ID != 7 || body.NextCursor == nil || *body.NextCursor != "Nw" {
		t.Fatalf("unexpected body: %s", rec.Body)
	}
}

func TestListListings_EmptyPage(t *testing.T) {
	rec := doGet(newTestHandler(&fakeService{}), "/v1/listings", "good-key")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if got, want := rec.Body.String(), "{\"items\":[],\"next_cursor\":null}\n"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestListListings_Errors(t *testing.T) {
	cases := []struct {
		name   string
		target string
		apiKey string
		svcErr error
		status int
		code   string
	}{
		{"no api key", "/v1/listings", "", nil, http.StatusUnauthorized, "unauthorized"},
		{"bad api key", "/v1/listings", "bad-key", nil, http.StatusUnauthorized, "unauthorized"},
		{"non-numeric limit", "/v1/listings?limit=abc", "good-key", nil, http.StatusBadRequest, "validation_error"},
		{"zero limit", "/v1/listings?limit=0", "good-key", nil, http.StatusBadRequest, "validation_error"},
		{"service validation", "/v1/listings", "good-key", &domain.ValidationError{Field: "cursor", Message: "invalid"}, http.StatusBadRequest, "validation_error"},
		{"internal", "/v1/listings", "good-key", io.ErrUnexpectedEOF, http.StatusInternalServerError, "internal_error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doGet(newTestHandler(&fakeService{err: tc.svcErr}), tc.target, tc.apiKey)

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
