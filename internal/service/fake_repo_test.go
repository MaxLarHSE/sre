package service

import (
	"context"
	"encoding/json"
	"slices"

	"trade/internal/domain"
)

// fakeRepo - репозиторий в памяти. Все изменяющие методы атомарны, как и настоящие:
// при любой ошибке ничего не меняется.
type fakeRepo struct {
	listings   []domain.Listing // отсортированы по id DESC
	lastFilter domain.ListingFilter
	nextID     int64

	records map[string]domain.IdempotencyRecord // ключ - gameID + "/" + key

	createCalls int
	createErr   error
	// beforeCreate вызывается в начале CreateListing, чтобы смоделировать параллельный запрос
	beforeCreate func()

	// changeCalls - сколько раз вызывали BuyListing, CancelListing и UpdateListingPrice
	changeCalls int
	// beforeChange вызывается в начале этих методов, чтобы смоделировать параллельный запрос
	beforeChange func()
}

func (f *fakeRepo) Ping(context.Context) error { return nil }

func (f *fakeRepo) GameIDByAPIKey(_ context.Context, key string) (string, error) {
	if key == "good-key" {
		return "game-1", nil
	}
	return "", domain.ErrNotFound
}

func (f *fakeRepo) ListActiveListings(_ context.Context, filter domain.ListingFilter) ([]domain.Listing, error) {
	f.lastFilter = filter
	var out []domain.Listing
	for _, l := range f.listings {
		if filter.BeforeID > 0 && l.ID >= filter.BeforeID {
			continue
		}
		if len(out) == filter.Limit {
			break
		}
		out = append(out, l)
	}
	return out, nil
}

func (f *fakeRepo) CreateListing(_ context.Context, l domain.Listing, key, requestHash string) (domain.Listing, error) {
	if hook := f.beforeCreate; hook != nil {
		f.beforeCreate = nil
		hook()
	}

	f.createCalls++
	if f.createErr != nil {
		return domain.Listing{}, f.createErr
	}
	if f.hasRecord(l.GameID, key) {
		return domain.Listing{}, domain.ErrIdempotencyKeyExists
	}

	f.nextID++
	l.ID = f.nextID
	l.Status = domain.ListingActive

	f.saveRecord(l.GameID, key, requestHash, l)
	f.listings = append([]domain.Listing{l}, f.listings...)
	return l, nil
}

func (f *fakeRepo) BuyListing(_ context.Context, gameID string, p domain.Purchase, key, requestHash string) (domain.Listing, error) {
	return f.changeListing(gameID, p.ListingID, key, requestHash,
		func(l domain.Listing) error { return l.CanBeBought(p) },
		func(l *domain.Listing) { l.Status, l.BuyerID = domain.ListingSold, p.BuyerID },
	)
}

func (f *fakeRepo) CancelListing(_ context.Context, gameID string, c domain.Cancellation, key, requestHash string) (domain.Listing, error) {
	return f.changeListing(gameID, c.ListingID, key, requestHash,
		func(l domain.Listing) error { return l.CanBeCancelled(c) },
		func(l *domain.Listing) { l.Status = domain.ListingCancelled },
	)
}

func (f *fakeRepo) UpdateListingPrice(_ context.Context, gameID string, c domain.PriceChange, key, requestHash string) (domain.Listing, error) {
	return f.changeListing(gameID, c.ListingID, key, requestHash,
		func(l domain.Listing) error { return l.CanChangePrice(c) },
		func(l *domain.Listing) { l.Price = c.Price },
	)
}

// changeListing - общая часть изменяющих методов: найти лот, проверить, изменить, сохранить результат.
func (f *fakeRepo) changeListing(
	gameID string, listingID int64, key, requestHash string,
	check func(domain.Listing) error,
	apply func(*domain.Listing),
) (domain.Listing, error) {
	if hook := f.beforeChange; hook != nil {
		f.beforeChange = nil
		hook()
	}
	f.changeCalls++

	i := slices.IndexFunc(f.listings, func(l domain.Listing) bool { return l.ID == listingID && l.GameID == gameID })
	if i < 0 {
		return domain.Listing{}, domain.ErrNotFound
	}
	l := f.listings[i]
	if err := check(l); err != nil {
		return domain.Listing{}, err
	}
	if f.hasRecord(gameID, key) {
		return domain.Listing{}, domain.ErrIdempotencyKeyExists
	}

	apply(&l)
	f.saveRecord(gameID, key, requestHash, l)
	f.listings[i] = l
	return l, nil
}

func (f *fakeRepo) GetIdempotencyRecord(_ context.Context, gameID, key string) (domain.IdempotencyRecord, error) {
	rec, ok := f.records[gameID+"/"+key]
	if !ok {
		return domain.IdempotencyRecord{}, domain.ErrNotFound
	}
	return rec, nil
}

func (f *fakeRepo) hasRecord(gameID, key string) bool {
	_, ok := f.records[gameID+"/"+key]
	return ok
}

func (f *fakeRepo) saveRecord(gameID, key, requestHash string, l domain.Listing) {
	response, err := json.Marshal(l)
	if err != nil {
		panic(err)
	}
	if f.records == nil {
		f.records = map[string]domain.IdempotencyRecord{}
	}
	f.records[gameID+"/"+key] = domain.IdempotencyRecord{GameID: gameID, Key: key, RequestHash: requestHash, Response: response}
}
