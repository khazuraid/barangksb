-- +goose Up
ALTER TABLE locations ADD COLUMN IF NOT EXISTS code TEXT DEFAULT '';

ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS size TEXT DEFAULT '';
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS material TEXT DEFAULT '';
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS item_code TEXT DEFAULT '';
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS register_number TEXT DEFAULT '';
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS location_code TEXT DEFAULT '';

-- +goose Down
ALTER TABLE inventory_items DROP COLUMN IF EXISTS location_code;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS register_number;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS item_code;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS material;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS size;

ALTER TABLE locations DROP COLUMN IF EXISTS code;
