package service

import (
	"unicode/utf8"

	"trade/internal/domain"
)

const maxCodeLen = 64

func checkCode(field, value string) error {
	if value == "" || utf8.RuneCountInString(value) > maxCodeLen {
		return &domain.ValidationError{Field: field, Message: "must be 1-64 characters"}
	}
	return nil
}

func checkPrice(price int64) error {
	if price < 1 {
		return &domain.ValidationError{Field: "price", Message: "must be positive"}
	}
	return nil
}

func checkListingID(id int64) error {
	if id < 1 {
		return &domain.ValidationError{Field: "id", Message: "must be a positive integer"}
	}
	return nil
}

func firstErr(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
