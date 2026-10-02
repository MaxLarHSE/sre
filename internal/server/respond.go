package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"trade/internal/domain"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorBody{Error: errorDetail{Code: code, Message: message}})
}

func (s *Server) handleError(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr *domain.ValidationError

	switch {
	case errors.As(err, &validationErr):
		writeError(w, http.StatusBadRequest, "validation_error", validationErr.Error())
	case errors.Is(err, domain.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing or invalid X-API-Key")
	case errors.Is(err, domain.ErrIdempotencyKeyReused):
		writeError(w, http.StatusUnprocessableEntity, "idempotency_key_reused", "Idempotency-Key was already used with a different request")
	case errors.Is(err, domain.ErrNotListingOwner):
		writeError(w, http.StatusForbidden, "not_owner", "listing belongs to another seller")
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, domain.ErrListingNotActive):
		writeError(w, http.StatusConflict, "listing_not_active", "listing is already sold or cancelled")
	case errors.Is(err, domain.ErrOwnListing):
		writeError(w, http.StatusConflict, "own_listing", "seller cannot buy own listing")
	case errors.Is(err, domain.ErrPriceMismatch):
		writeError(w, http.StatusConflict, "price_mismatch", "listing price has changed")
	case errors.Is(err, context.Canceled):
	default:
		s.logger.Error("internal error", "error", err, "path", r.URL.Path)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

const maxBodyBytes = 64 << 10

func readJSON(w http.ResponseWriter, r *http.Request, v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()

	if err := dec.Decode(v); err != nil {
		return &domain.ValidationError{Field: "body", Message: "invalid JSON: " + err.Error()}
	}
	if dec.More() {
		return &domain.ValidationError{Field: "body", Message: "must contain a single JSON object"}
	}
	return nil
}
