package service

import (
	"context"
	"errors"
	"testing"

	"trade/internal/domain"
)

var validPriceChange = UpdateListingPriceInput{ListingID: 1, SellerID: "vasya", Price: 150}

func TestUpdateListingPrice_Updates(t *testing.T) {
	repo := repoWithSword()

	l, err := New(repo).UpdateListingPrice(context.Background(), "game-1", "upd-1", validPriceChange)
	if err != nil {
		t.Fatal(err)
	}

	if l.Price != 150 || l.Status != domain.ListingActive {
		t.Fatalf("unexpected result: %+v", l)
	}
	if repo.listings[0].Price != 150 {
		t.Fatalf("price in repo = %d, want 150", repo.listings[0].Price)
	}
}

func TestUpdateListingPrice_ReplayReturnsSameResult(t *testing.T) {
	repo := repoWithSword()
	svc := New(repo)

	first, err := svc.UpdateListingPrice(context.Background(), "game-1", "upd-1", validPriceChange)
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.UpdateListingPrice(context.Background(), "game-1", "upd-1", validPriceChange)
	if err != nil {
		t.Fatal(err)
	}

	if first != second {
		t.Fatalf("replay = %+v, want %+v", second, first)
	}
	if repo.changeCalls != 1 {
		t.Fatalf("update executed %d times, want 1", repo.changeCalls)
	}
}

func TestUpdateListingPrice_ConcurrentRequestsWithSameKey(t *testing.T) {
	repo := repoWithSword()
	svc := New(repo)

	// оба запроса не нашли ключ; первый успел поменять цену, второй упирается в занятый ключ
	var winner domain.Listing
	repo.beforeChange = func() {
		w, err := svc.UpdateListingPrice(context.Background(), "game-1", "upd-1", validPriceChange)
		if err != nil {
			t.Fatal(err)
		}
		winner = w
	}

	got, err := svc.UpdateListingPrice(context.Background(), "game-1", "upd-1", validPriceChange)
	if err != nil {
		t.Fatalf("request with the same key must get winner's result, got %v", err)
	}
	if got != winner {
		t.Fatalf("got %+v, want %+v", got, winner)
	}
}

func TestUpdateListingPrice_BuyWithOldPriceFails(t *testing.T) {
	svc := New(repoWithSword())

	if _, err := svc.UpdateListingPrice(context.Background(), "game-1", "upd-1", validPriceChange); err != nil {
		t.Fatal(err)
	}
	// покупатель видел старую цену 100, игра списала 100
	_, err := svc.BuyListing(context.Background(), "game-1", "buy-1", validPurchase)

	if !errors.Is(err, domain.ErrPriceMismatch) {
		t.Fatalf("err = %v, want ErrPriceMismatch", err)
	}
}

func TestUpdateListingPrice_BusinessErrors(t *testing.T) {
	cases := []struct {
		name   string
		gameID string
		input  func(in *UpdateListingPriceInput)
		want   error
	}{
		{"missing listing", "game-1", func(in *UpdateListingPriceInput) { in.ListingID = 999 }, domain.ErrNotFound},
		{"listing of another game", "game-2", func(*UpdateListingPriceInput) {}, domain.ErrNotFound},
		{"not owner", "game-1", func(in *UpdateListingPriceInput) { in.SellerID = "petya" }, domain.ErrNotListingOwner},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := repoWithSword()
			in := validPriceChange
			tc.input(&in)

			_, err := New(repo).UpdateListingPrice(context.Background(), tc.gameID, "upd-1", in)

			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if repo.listings[0].Price != 100 {
				t.Fatal("price must not change after failed update")
			}
			if len(repo.records) != 0 {
				t.Fatal("failed update must not be stored")
			}
		})
	}
}

func TestUpdateListingPrice_ClosedListing(t *testing.T) {
	svc := New(repoWithSword())

	if _, err := svc.CancelListing(context.Background(), "game-1", "cancel-1", validCancel); err != nil {
		t.Fatal(err)
	}
	_, err := svc.UpdateListingPrice(context.Background(), "game-1", "upd-1", validPriceChange)

	if !errors.Is(err, domain.ErrListingNotActive) {
		t.Fatalf("err = %v, want ErrListingNotActive", err)
	}
}

func TestUpdateListingPrice_Validation(t *testing.T) {
	cases := []struct {
		name  string
		key   string
		input func(in *UpdateListingPriceInput)
		field string
	}{
		{"no key", "", func(*UpdateListingPriceInput) {}, "Idempotency-Key"},
		{"zero id", "k", func(in *UpdateListingPriceInput) { in.ListingID = 0 }, "id"},
		{"no seller", "k", func(in *UpdateListingPriceInput) { in.SellerID = "" }, "seller_id"},
		{"zero price", "k", func(in *UpdateListingPriceInput) { in.Price = 0 }, "price"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := repoWithSword()
			in := validPriceChange
			tc.input(&in)

			_, err := New(repo).UpdateListingPrice(context.Background(), "game-1", tc.key, in)

			var vErr *domain.ValidationError
			if !errors.As(err, &vErr) || vErr.Field != tc.field {
				t.Fatalf("err = %v, want ValidationError on %q", err, tc.field)
			}
			if repo.changeCalls != 0 {
				t.Fatal("repo must not be called on invalid input")
			}
		})
	}
}
