-- +goose Up
CREATE TABLE IF NOT EXISTS telegram_message_log (
    id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    chat_title TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'sent',
    error TEXT,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tg_log_sent_at ON telegram_message_log(sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_tg_log_chat_id ON telegram_message_log(chat_id);

-- +goose Down
DROP TABLE IF EXISTS telegram_message_log;