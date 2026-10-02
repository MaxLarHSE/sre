package service

import (
	"context"
	"encoding/base64"
	"errors"
	"slices"
	"testing"

	"trade/internal/domain"
)

func listingsWithIDs(ids ...int64) []domain.Listing {
	out := make([]domain.Listing, 0, len(ids))
	for _, id := range ids {
		out = append(out, domain.Listing{ID: id})
	}
	return out
}

func listingIDs(ls []domain.Listing) []int64 {
	out := make([]int64, 0, len(ls))
	for _, l := range ls {
		out = append(out, l.ID)
	}
	return out
}

func TestListListings_PaginatesWithCursor(t *testing.T) {
	svc := New(&fakeRepo{listings: listingsWithIDs(5, 4, 3, 2, 1)})
	ctx := context.Background()

	first, err := svc.ListListings(ctx, "game-1", ListListingsParams{Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if got := listingIDs(first.Items); !slices.Equal(got, []int64{5, 4}) {
		t.Fatalf("first page = %v, want [5 4]", got)
	}
	if first.NextCursor == "" {
		t.Fatal("expected next cursor on first page")
	}

	second, err := svc.ListListings(ctx, "game-1", ListListingsParams{Limit: 2, Cursor: first.NextCursor})
	if err != nil {
		t.Fatal(err)
	}
	if got := listingIDs(second.Items); !slices.Equal(got, []int64{3, 2}) {
		t.Fatalf("second page = %v, want [3 2]", got)
	}

	last, err := svc.ListListings(ctx, "game-1", ListListingsParams{Limit: 2, Cursor: second.NextCursor})
	if err != nil {
		t.Fatal(err)
	}
	if got := listingIDs(last.Items); !slices.Equal(got, []int64{1}) {
		t.Fatalf("last page = %v, want [1]", got)
	}
	if last.NextCursor != "" {
		t.Fatalf("expected no cursor on last page, got %q", last.NextCursor)
	}
}

func TestListListings_DefaultLimit(t *testing.T) {
	repo := &fakeRepo{}
	if _, err := New(repo).ListListings(context.Background(), "game-1", ListListingsParams{}); err != nil {
		t.Fatal(err)
	}
	if repo.lastFilter.Limit != defaultListingsLimit+1 {
		t.Fatalf("repo limit = %d, want %d", repo.lastFilter.Limit, defaultListingsLimit+1)
	}
}

func TestListListings_Validation(t *testing.T) {
	cases := map[string]ListListingsParams{
		"limit too big":       {Limit: maxListingsLimit + 1},
		"negative limit":      {Limit: -1},
		"not base64 cursor":   {Cursor: "!!!"},
		"non-numeric cursor":  {Cursor: base64.RawURLEncoding.EncodeToString([]byte("abc"))},
		"non-positive cursor": {Cursor: encodeCursor(0)},
	}
	for name, params := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := New(&fakeRepo{}).ListListings(context.Background(), "game-1", params)
			var vErr *domain.ValidationError
			if !errors.As(err, &vErr) {
				t.Fatalf("err = %v, want ValidationError", err)
			}
		})
	}
}

func TestAuthenticate(t *testing.T) {
	svc := New(&fakeRepo{})

	if id, err := svc.Authenticate(context.Background(), "good-key"); err != nil || id != "game-1" {
		t.Fatalf("good key: id=%q err=%v", id, err)
	}
	for _, key := range []string{"", "bad-key"} {
		if _, err := svc.Authenticate(context.Background(), key); !errors.Is(err, domain.ErrUnauthorized) {
			t.Fatalf("key %q: err = %v, want ErrUnauthorized", key, err)
		}
	}
}
