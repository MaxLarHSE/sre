package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"trade/internal/domain"
	"trade/internal/service"
	"trade/internal/web"
)

type Service interface {
	Ready(ctx context.Context) error
	Authenticate(ctx context.Context, apiKey string) (string, error)
	ListListings(ctx context.Context, gameID string, p service.ListListingsParams) (service.ListingsPage, error)
	CreateListing(ctx context.Context, gameID, idempotencyKey string, in service.CreateListingInput) (domain.Listing, error)
	BuyListing(ctx context.Context, gameID, idempotencyKey string, in service.BuyListingInput) (domain.Listing, error)
	CancelListing(ctx context.Context, gameID, idempotencyKey string, in service.CancelListingInput) (domain.Listing, error)
	UpdateListingPrice(ctx context.Context, gameID, idempotencyKey string, in service.UpdateListingPriceInput) (domain.Listing, error)
}

type Server struct {
	svc    Service
	logger *slog.Logger
}

func New(addr string, svc Service, logger *slog.Logger) *http.Server {
	s := &Server{svc: svc, logger: logger}

	return &http.Server{
		Addr:              addr,
		Handler:           s.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /{$}", web.Handler())

	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)

	mux.Handle("GET /v1/listings", s.auth(http.HandlerFunc(s.listListings)))
	mux.Handle("POST /v1/listings", s.auth(http.HandlerFunc(s.createListing)))
	mux.Handle("PATCH /v1/listings/{id}", s.auth(http.HandlerFunc(s.updateListing)))
	mux.Handle("POST /v1/listings/{id}/buy", s.auth(http.HandlerFunc(s.buyListing)))
	mux.Handle("POST /v1/listings/{id}/cancel", s.auth(http.HandlerFunc(s.cancelListing)))

	return s.recoverer(s.accessLog(mux))
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.Ready(r.Context()); err != nil {
		s.logger.Error("not ready", "error", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}
