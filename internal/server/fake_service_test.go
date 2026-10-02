package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"

	"trade/internal/domain"
	"trade/internal/service"
)

// fakeService запоминает аргументы последнего вызова и возвращает err, если он задан.
type fakeService struct {
	page service.ListingsPage
	err  error

	lastGameID string
	lastKey    string
	lastParams service.ListListingsParams
	lastInput  service.CreateListingInput
	lastBuy    service.BuyListingInput
	lastCancel service.CancelListingInput
	lastUpdate service.UpdateListingPriceInput
}

func (f *fakeService) Ready(context.Context) error { return nil }

func (f *fakeService) Authenticate(_ context.Context, key string) (string, error) {
	if key == "good-key" {
		return "game-1", nil
	}
	return "", domain.ErrUnauthorized
}

func (f *fakeService) ListListings(_ context.Context, gameID string, p service.ListListingsParams) (service.ListingsPage, error) {
	f.lastGameID, f.lastParams = gameID, p
	return f.page, f.err
}

func (f *fakeService) CreateListing(_ context.Context, gameID, key string, in service.CreateListingInput) (domain.Listing, error) {
	f.lastGameID, f.lastKey, f.lastInput = gameID, key, in
	if f.err != nil {
		return domain.Listing{}, f.err
	}
	return domain.Listing{
		ID: 42, GameID: gameID, SellerID: in.SellerID, ItemCode: in.ItemCode,
		Quantity: in.Quantity, Price: in.Price, Status: domain.ListingActive,
	}, nil
}

func (f *fakeService) BuyListing(_ context.Context, gameID, key string, in service.BuyListingInput) (domain.Listing, error) {
	f.lastGameID, f.lastKey, f.lastBuy = gameID, key, in
	if f.err != nil {
		return domain.Listing{}, f.err
	}
	return domain.Listing{
		ID: in.ListingID, GameID: gameID, SellerID: "vasya", ItemCode: "iron_sword",
		Quantity: 1, Price: in.Price, Status: domain.ListingSold, BuyerID: in.BuyerID,
	}, nil
}

func (f *fakeService) CancelListing(_ context.Context, gameID, key string, in service.CancelListingInput) (domain.Listing, error) {
	f.lastGameID, f.lastKey, f.lastCancel = gameID, key, in
	if f.err != nil {
		return domain.Listing{}, f.err
	}
	return domain.Listing{
		ID: in.ListingID, GameID: gameID, SellerID: in.SellerID, ItemCode: "wolf_pelt",
		Quantity: 50, Price: 200, Status: domain.ListingCancelled,
	}, nil
}

func (f *fakeService) UpdateListingPrice(_ context.Context, gameID, key string, in service.UpdateListingPriceInput) (domain.Listing, error) {
	f.lastGameID, f.lastKey, f.lastUpdate = gameID, key, in
	if f.err != nil {
		return domain.Listing{}, f.err
	}
	return domain.Listing{
		ID: in.ListingID, GameID: gameID, SellerID: in.SellerID, ItemCode: "iron_sword",
		Quantity: 1, Price: in.Price, Status: domain.ListingActive,
	}, nil
}

func newTestHandler(svc Service) http.Handler {
	s := &Server{svc: svc, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	return s.routes()
}

func doRequest(h http.Handler, method, target, body, apiKey, idemKey string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}
	if idemKey != "" {
		req.Header.Set("Idempotency-Key", idemKey)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func doGet(h http.Handler, target, apiKey string) *httptest.ResponseRecorder {
	return doRequest(h, http.MethodGet, target, "", apiKey, "")
}

func doPost(h http.Handler, target, body, apiKey, idemKey string) *httptest.ResponseRecorder {
	return doRequest(h, http.MethodPost, target, body, apiKey, idemKey)
}
