package domain

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")

	ErrListingNotActive = errors.New("listing is not active")
	ErrOwnListing       = errors.New("cannot buy own listing")
	ErrPriceMismatch    = errors.New("listing price has changed")
	ErrNotListingOwner  = errors.New("listing belongs to another seller")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
