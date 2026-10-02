ALTER TABLE idempotency_keys ADD COLUMN status_code INT NOT NULL DEFAULT 200;
ALTER TABLE idempotency_keys ALTER COLUMN status_code DROP DEFAULT;

ALTER TABLE idempotency_keys DROP COLUMN request_hash;
