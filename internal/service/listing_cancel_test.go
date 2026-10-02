package service

import (
	"context"
	"errors"
	"testing"

	"trade/internal/domain"
)

var validCancel = CancelListingInput{ListingID: 1, SellerID: "vasya"}

func TestCancelListing_Cancels(t *testing.T) {
	repo := repoWithSword()

	l, err := New(repo).CancelListing(context.Background(), "game-1", "cancel-1", validCancel)
	if err != nil {
		t.Fatal(err)
	}

	if l.Status != domain.ListingCancelled || l.SellerID != "vasya" || l.ItemCode != "iron_sword" || l.Quantity != 1 {
		t.Fatalf("unexpected result: %+v", l)
	}
	if repo.listings[0].Status != domain.ListingCancelled {
		t.Fatal("listing in repo must be cancelled")
	}
}

func TestCancelListing_ReplayReturnsSameResult(t *testing.T) {
	repo := repoWithSword()
	svc := New(repo)

	first, err := svc.CancelListing(context.Background(), "game-1", "cancel-1", validCancel)
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.CancelListing(context.Background(), "game-1", "cancel-1", validCancel)
	if err != nil {
		t.Fatalf("replay must succeed, got %v", err)
	}

	if first != second {
		t.Fatalf("replay = %+v, want %+v", second, first)
	}
	if repo.changeCalls != 1 {
		t.Fatalf("cancel executed %d times, want 1", repo.changeCalls)
	}
}

func TestCancelListing_ConcurrentRequestsWithSameKey(t *testing.T) {
	repo := repoWithSword()
	svc := New(repo)

	// оба запроса не нашли ключ; первый успел снять лот, второй видит его уже снятым
	var winner domain.Listing
	repo.beforeChange = func() {
		w, err := svc.CancelListing(context.Background(), "game-1", "cancel-1", validCancel)
		if err != nil {
			t.Fatal(err)
		}
		winner = w
	}

	got, err := svc.CancelListing(context.Background(), "game-1", "cancel-1", validCancel)
	if err != nil {
		t.Fatalf("request with the same key must get winner's result, got %v", err)
	}
	if got != winner {
		t.Fatalf("got %+v, want %+v", got, winner)
	}
}

func TestCancelListing_BoughtWhileCancelling(t *testing.T) {
	repo := repoWithSword()
	svc := New(repo)

	// продавец снимает лот, но покупатель успевает купить его первым
	repo.beforeChange = func() {
		if _, err := svc.BuyListing(context.Background(), "game-1", "buy-1", validPurchase); err != nil {
			t.Fatal(err)
		}
	}

	_, err := svc.CancelListing(context.Background(), "game-1", "cancel-1", validCancel)

	if !errors.Is(err, domain.ErrListingNotActive) {
		t.Fatalf("err = %v, want ErrListingNotActive", err)
	}
	if repo.listings[0].Status != domain.ListingSold {
		t.Fatalf("status = %q, want sold", repo.listings[0].Status)
	}
}

func TestCancelListing_BusinessErrors(t *testing.T) {
	cases := []struct {
		name   string
		gameID string
		input  func(in *CancelListingInput)
		want   error
	}{
		{"missing listing", "game-1", func(in *CancelListingInput) { in.ListingID = 999 }, domain.ErrNotFound},
		{"listing of another game", "game-2", func(*CancelListingInput) {}, domain.ErrNotFound},
		{"not owner", "game-1", func(in *CancelListingInput) { in.SellerID = "petya" }, domain.ErrNotListingOwner},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := repoWithSword()
			in := validCancel
			tc.input(&in)

			_, err := New(repo).CancelListing(context.Background(), tc.gameID, "cancel-1", in)

			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if repo.listings[0].Status != domain.ListingActive {
				t.Fatal("listing must stay active after failed cancel")
			}
			if len(repo.records) != 0 {
				t.Fatal("failed cancel must not be stored")
			}
		})
	}
}

func TestCancelListing_AlreadySold(t *testing.T) {
	svc := New(repoWithSword())

	if _, err := svc.BuyListing(context.Background(), "game-1", "buy-1", validPurchase); err != nil {
		t.Fatal(err)
	}
	_, err := svc.CancelListing(context.Background(), "game-1", "cancel-1", validCancel)

	if !errors.Is(err, domain.ErrListingNotActive) {
		t.Fatalf("err = %v, want ErrListingNotActive", err)
	}
}

func TestCancelListing_Validation(t *testing.T) {
	cases := []struct {
		name  string
		key   string
		input func(in *CancelListingInput)
		field string
	}{
		{"no key", "", func(*CancelListingInput) {}, "Idempotency-Key"},
		{"zero id", "k", func(in *CancelListingInput) { in.ListingID = 0 }, "id"},
		{"no seller", "k", func(in *CancelListingInput) { in.SellerID = "" }, "seller_id"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := repoWithSword()
			in := validCancel
			tc.input(&in)

			_, err := New(repo).CancelListing(context.Background(), "game-1", tc.key, in)

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
