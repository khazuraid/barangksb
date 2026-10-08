-- +goose Up
CREATE TABLE IF NOT EXISTS app_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- seed telegram defaults (chat IDs now managed via telegram_subscribers table)
INSERT INTO app_settings (key, value) VALUES
 ('telegram_bot_token', ''),
 ('telegram_alert_low_stock', 'true'),
 ('telegram_alert_daily_time', '07:00')
ON CONFLICT (key) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS app_settings;