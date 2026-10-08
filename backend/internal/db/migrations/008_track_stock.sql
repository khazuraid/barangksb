-- +goose Up
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS track_stock BOOLEAN NOT NULL DEFAULT TRUE;

-- +goose Down
ALTER TABLE inventory_items DROP COLUMN IF EXISTS track_stock;
