package service

import (
	"context"
	"errors"
	"testing"

	"trade/internal/domain"
)

// repoWithSword - в game-1 выставлен меч vasya за 100.
func repoWithSword() *fakeRepo {
	return &fakeRepo{listings: []domain.Listing{{
		ID: 1, GameID: "game-1", SellerID: "vasya", ItemCode: "iron_sword",
		Quantity: 1, Price: 100, Status: domain.ListingActive,
	}}}
}

var validPurchase = BuyListingInput{ListingID: 1, BuyerID: "petya", Price: 100}

func TestBuyListing_Sells(t *testing.T) {
	repo := repoWithSword()

	l, err := New(repo).BuyListing(context.Background(), "game-1", "buy-1", validPurchase)
	if err != nil {
		t.Fatal(err)
	}

	if l.Status != domain.ListingSold || l.BuyerID != "petya" || l.SellerID != "vasya" || l.ItemCode != "iron_sword" || l.Price != 100 {
		t.Fatalf("unexpected result: %+v", l)
	}
	if repo.listings[0].Status != domain.ListingSold {
		t.Fatal("listing in repo must be sold")
	}
}

func TestBuyListing_ReplayReturnsSameResult(t *testing.T) {
	repo := repoWithSword()
	svc := New(repo)

	first, err := svc.BuyListing(context.Background(), "game-1", "buy-1", validPurchase)
	if err != nil {
		t.Fatal(err)
	}
	// ответ потерялся, сервер игры повторяет с тем же ключом: лот уже продан, но это не ошибка
	second, err := svc.BuyListing(context.Background(), "game-1", "buy-1", validPurchase)
	if err != nil {
		t.Fatalf("replay must succeed, got %v", err)
	}

	if first != second {
		t.Fatalf("replay = %+v, want %+v", second, first)
	}
	if repo.changeCalls != 1 {
		t.Fatalf("buy executed %d times, want 1", repo.changeCalls)
	}
}

func TestBuyListing_NewAttemptOnSoldListing(t *testing.T) {
	svc := New(repoWithSword())

	if _, err := svc.BuyListing(context.Background(), "game-1", "buy-1", validPurchase); err != nil {
		t.Fatal(err)
	}
	// второе нажатие "Купить" - новая попытка с новым ключом
	_, err := svc.BuyListing(context.Background(), "game-1", "buy-2", validPurchase)

	if !errors.Is(err, domain.ErrListingNotActive) {
		t.Fatalf("err = %v, want ErrListingNotActive", err)
	}
}

func TestBuyListing_ConcurrentBuyers(t *testing.T) {
	repo := repoWithSword()
	svc := New(repo)

	// пока petya ждёт блокировку лота, masha успевает купить его первой
	repo.beforeChange = func() {
		in := validPurchase
		in.BuyerID = "masha"
		if _, err := svc.BuyListing(context.Background(), "game-1", "masha-key", in); err != nil {
			t.Fatal(err)
		}
	}

	_, err := svc.BuyListing(context.Background(), "game-1", "petya-key", validPurchase)

	if !errors.Is(err, domain.ErrListingNotActive) {
		t.Fatalf("err = %v, want ErrListingNotActive", err)
	}
	if repo.listings[0].BuyerID != "masha" {
		t.Fatalf("buyer = %q, want masha", repo.listings[0].BuyerID)
	}
}

func TestBuyListing_ConcurrentRequestsWithSameKey(t *testing.T) {
	repo := repoWithSword()
	svc := New(repo)

	// оба запроса не нашли ключ; первый успел купить, второй видит лот уже проданным
	var winner domain.Listing
	repo.beforeChange = func() {
		w, err := svc.BuyListing(context.Background(), "game-1", "buy-1", validPurchase)
		if err != nil {
			t.Fatal(err)
		}
		winner = w
	}

	got, err := svc.BuyListing(context.Background(), "game-1", "buy-1", validPurchase)
	if err != nil {
		t.Fatalf("request with the same key must get winner's result, got %v", err)
	}
	if got != winner {
		t.Fatalf("got %+v, want %+v", got, winner)
	}
}

func TestBuyListing_KeyReusedWithDifferentRequest(t *testing.T) {
	svc := New(repoWithSword())

	if _, err := svc.BuyListing(context.Background(), "game-1", "buy-1", validPurchase); err != nil {
		t.Fatal(err)
	}
	other := validPurchase
	other.BuyerID = "masha"
	_, err := svc.BuyListing(context.Background(), "game-1", "buy-1", other)

	if !errors.Is(err, domain.ErrIdempotencyKeyReused) {
		t.Fatalf("err = %v, want ErrIdempotencyKeyReused", err)
	}
}

func TestBuyListing_KeyFromCreateCannotBeReusedForBuy(t *testing.T) {
	repo := &fakeRepo{}
	svc := New(repo)

	created, err := svc.CreateListing(context.Background(), "game-1", "same-key", validInput)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.BuyListing(context.Background(), "game-1", "same-key", BuyListingInput{ListingID: created.ID, BuyerID: "petya", Price: created.Price})

	if !errors.Is(err, domain.ErrIdempotencyKeyReused) {
		t.Fatalf("err = %v, want ErrIdempotencyKeyReused", err)
	}
}

func TestBuyListing_BusinessErrors(t *testing.T) {
	cases := []struct {
		name   string
		gameID string
		input  func(in *BuyListingInput)
		want   error
	}{
		{"missing listing", "game-1", func(in *BuyListingInput) { in.ListingID = 999 }, domain.ErrNotFound},
		{"listing of another game", "game-2", func(*BuyListingInput) {}, domain.ErrNotFound},
		{"own listing", "game-1", func(in *BuyListingInput) { in.BuyerID = "vasya" }, domain.ErrOwnListing},
		{"price changed", "game-1", func(in *BuyListingInput) { in.Price = 90 }, domain.ErrPriceMismatch},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := repoWithSword()
			in := validPurchase
			tc.input(&in)

			_, err := New(repo).BuyListing(context.Background(), tc.gameID, "buy-1", in)

			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if repo.listings[0].Status != domain.ListingActive {
				t.Fatal("listing must stay active after failed purchase")
			}
			if len(repo.records) != 0 {
				t.Fatal("failed purchase must not be stored")
			}
		})
	}
}

func TestBuyListing_Validation(t *testing.T) {
	cases := []struct {
		name  string
		key   string
		input func(in *BuyListingInput)
		field string
	}{
		{"no key", "", func(*BuyListingInput) {}, "Idempotency-Key"},
		{"zero id", "k", func(in *BuyListingInput) { in.ListingID = 0 }, "id"},
		{"no buyer", "k", func(in *BuyListingInput) { in.BuyerID = "" }, "buyer_id"},
		{"zero price", "k", func(in *BuyListingInput) { in.Price = 0 }, "price"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := repoWithSword()
			in := validPurchase
			tc.input(&in)

			_, err := New(repo).BuyListing(context.Background(), "game-1", tc.key, in)

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
