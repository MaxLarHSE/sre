package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"trade/internal/domain"
)

var validInput = CreateListingInput{SellerID: "vasya", ItemCode: "iron_sword", Quantity: 1, Price: 100}

func TestCreateListing_CreatesAndStoresResult(t *testing.T) {
	repo := &fakeRepo{}

	l, err := New(repo).CreateListing(context.Background(), "game-1", "key-1", validInput)
	if err != nil {
		t.Fatal(err)
	}

	if l.ID == 0 || l.GameID != "game-1" || l.SellerID != "vasya" || l.Status != domain.ListingActive {
		t.Fatalf("unexpected listing: %+v", l)
	}
	if len(repo.listings) != 1 {
		t.Fatalf("listings in repo = %d, want 1", len(repo.listings))
	}
	if _, ok := repo.records["game-1/key-1"]; !ok {
		t.Fatal("idempotency record was not saved")
	}
}

func TestCreateListing_ReplayReturnsSameListing(t *testing.T) {
	repo := &fakeRepo{}
	svc := New(repo)

	first, err := svc.CreateListing(context.Background(), "game-1", "key-1", validInput)
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.CreateListing(context.Background(), "game-1", "key-1", validInput)
	if err != nil {
		t.Fatal(err)
	}

	if first.ID != second.ID {
		t.Fatalf("replay returned listing %d, want %d", second.ID, first.ID)
	}
	if repo.createCalls != 1 || len(repo.listings) != 1 {
		t.Fatalf("create called %d times, %d listings; want 1 and 1", repo.createCalls, len(repo.listings))
	}
}

func TestCreateListing_SameKeyDifferentGames(t *testing.T) {
	repo := &fakeRepo{}
	svc := New(repo)

	if _, err := svc.CreateListing(context.Background(), "game-1", "key-1", validInput); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateListing(context.Background(), "game-2", "key-1", validInput); err != nil {
		t.Fatal(err)
	}

	if len(repo.listings) != 2 {
		t.Fatalf("listings = %d, want 2: keys of different games must not collide", len(repo.listings))
	}
}

func TestCreateListing_KeyReusedWithDifferentRequest(t *testing.T) {
	repo := &fakeRepo{}
	svc := New(repo)

	if _, err := svc.CreateListing(context.Background(), "game-1", "key-1", validInput); err != nil {
		t.Fatal(err)
	}

	other := validInput
	other.Price = 999
	_, err := svc.CreateListing(context.Background(), "game-1", "key-1", other)

	if !errors.Is(err, domain.ErrIdempotencyKeyReused) {
		t.Fatalf("err = %v, want ErrIdempotencyKeyReused", err)
	}
	if len(repo.listings) != 1 {
		t.Fatalf("listings = %d, want 1", len(repo.listings))
	}
}

func TestCreateListing_ConcurrentRequestWins(t *testing.T) {
	repo := &fakeRepo{}
	svc := New(repo)

	// наш запрос уже проверил ключ (его ещё нет), но параллельный запрос
	// с тем же ключом успевает создать лот раньше нас
	var winner domain.Listing
	repo.beforeCreate = func() {
		w, err := svc.CreateListing(context.Background(), "game-1", "key-1", validInput)
		if err != nil {
			t.Fatal(err)
		}
		winner = w
	}

	got, err := svc.CreateListing(context.Background(), "game-1", "key-1", validInput)
	if err != nil {
		t.Fatal(err)
	}

	if got.ID != winner.ID {
		t.Fatalf("got listing %d, want winner's %d", got.ID, winner.ID)
	}
	if len(repo.listings) != 1 {
		t.Fatalf("listings = %d, want 1: second listing must not be created", len(repo.listings))
	}
}

func TestCreateListing_FailureIsNotStored(t *testing.T) {
	repo := &fakeRepo{createErr: errors.New("db is down")}
	svc := New(repo)

	if _, err := svc.CreateListing(context.Background(), "game-1", "key-1", validInput); err == nil {
		t.Fatal("expected error")
	}
	if len(repo.records) != 0 {
		t.Fatal("failed request must not be stored")
	}

	// после восстановления повтор с тем же ключом выполняется заново
	repo.createErr = nil
	if _, err := svc.CreateListing(context.Background(), "game-1", "key-1", validInput); err != nil {
		t.Fatal(err)
	}
	if len(repo.listings) != 1 {
		t.Fatalf("listings = %d, want 1", len(repo.listings))
	}
}

func TestCreateListing_Validation(t *testing.T) {
	cases := []struct {
		name  string
		key   string
		input func(in *CreateListingInput)
		field string
	}{
		{"no key", "", func(*CreateListingInput) {}, "Idempotency-Key"},
		{"long key", strings.Repeat("k", 256), func(*CreateListingInput) {}, "Idempotency-Key"},
		{"no seller", "k", func(in *CreateListingInput) { in.SellerID = "" }, "seller_id"},
		{"long seller", "k", func(in *CreateListingInput) { in.SellerID = strings.Repeat("я", 65) }, "seller_id"},
		{"no item", "k", func(in *CreateListingInput) { in.ItemCode = "" }, "item_code"},
		{"zero quantity", "k", func(in *CreateListingInput) { in.Quantity = 0 }, "quantity"},
		{"huge quantity", "k", func(in *CreateListingInput) { in.Quantity = 1 << 31 }, "quantity"},
		{"zero price", "k", func(in *CreateListingInput) { in.Price = 0 }, "price"},
		{"negative price", "k", func(in *CreateListingInput) { in.Price = -5 }, "price"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := validInput
			tc.input(&in)
			repo := &fakeRepo{}

			_, err := New(repo).CreateListing(context.Background(), "game-1", tc.key, in)

			var vErr *domain.ValidationError
			if !errors.As(err, &vErr) || vErr.Field != tc.field {
				t.Fatalf("err = %v, want ValidationError on %q", err, tc.field)
			}
			if repo.createCalls != 0 {
				t.Fatal("repo must not be called on invalid input")
			}
		})
	}

	// 64 символа кириллицей - это 128 байт, но лимит в символах
	in := validInput
	in.SellerID = strings.Repeat("я", 64)
	if _, err := New(&fakeRepo{}).CreateListing(context.Background(), "game-1", "k", in); err != nil {
		t.Fatalf("64-char seller_id must be valid: %v", err)
	}
}
