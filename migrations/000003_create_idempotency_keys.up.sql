CREATE TABLE idempotency_keys (
    game_id     UUID NOT NULL REFERENCES games (id),
    key         TEXT NOT NULL,
    status_code INT NOT NULL,
    response    JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (game_id, key)
);
