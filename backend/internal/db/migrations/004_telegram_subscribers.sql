-- +goose Up
CREATE TABLE IF NOT EXISTS telegram_subscribers (
    chat_id BIGINT PRIMARY KEY,
    type TEXT NOT NULL DEFAULT 'private',
    title TEXT NOT NULL DEFAULT '',
    username TEXT NOT NULL DEFAULT '',
    notify_in BOOLEAN NOT NULL DEFAULT TRUE,
    notify_out BOOLEAN NOT NULL DEFAULT TRUE,
    notify_adjust BOOLEAN NOT NULL DEFAULT FALSE,
    notify_low_stock BOOLEAN NOT NULL DEFAULT TRUE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS telegram_subscribers;