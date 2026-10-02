CREATE TYPE listing_status AS ENUM ('active', 'sold', 'cancelled');

CREATE TABLE listings (
    id         BIGSERIAL PRIMARY KEY,
    game_id    UUID NOT NULL REFERENCES games (id),
    seller_id  TEXT NOT NULL,
    item_code  TEXT NOT NULL,
    quantity   INT NOT NULL CHECK (quantity > 0),
    price      BIGINT NOT NULL CHECK (price > 0),
    status     listing_status NOT NULL DEFAULT 'active',
    buyer_id   TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    closed_at  TIMESTAMPTZ,

    -- покупатель есть только у проданного лота
    CONSTRAINT listings_buyer_chk CHECK ((status = 'sold') = (buyer_id IS NOT NULL)),
    -- закрытым считается любой лот, кроме активного
    CONSTRAINT listings_closed_chk CHECK ((status = 'active') = (closed_at IS NULL))
);

-- витрина: активные лоты игры с фильтром по item_code и курсором по id
CREATE INDEX listings_showcase_idx ON listings (game_id, status, item_code, id);
-- лоты конкретного продавца
CREATE INDEX listings_seller_idx ON listings (game_id, seller_id, status);
