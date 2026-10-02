-- хэш запроса: тот же ключ с другим телом запроса - ошибка клиента
ALTER TABLE idempotency_keys ADD COLUMN request_hash TEXT NOT NULL DEFAULT '';
ALTER TABLE idempotency_keys ALTER COLUMN request_hash DROP DEFAULT;

-- сохраняются только успешные результаты, HTTP-статус хранить не нужно
ALTER TABLE idempotency_keys DROP COLUMN status_code;
