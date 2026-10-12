-- +goose Up
CREATE TABLE IF NOT EXISTS item_units (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
    unit_number INTEGER NOT NULL DEFAULT 1,
    register_number TEXT NOT NULL DEFAULT '',
    barcode TEXT NOT NULL DEFAULT '',
    serial_number TEXT DEFAULT '',
    condition_status TEXT NOT NULL DEFAULT 'Baik (B)',
    location TEXT NOT NULL DEFAULT '',
    location_code TEXT DEFAULT '',
    photo_url TEXT DEFAULT '',
    geo_lat DOUBLE PRECISION,
    geo_lng DOUBLE PRECISION,
    geo_acc DOUBLE PRECISION,
    geo_name TEXT DEFAULT '',
    notes TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_item_units_item_id ON item_units(item_id);
CREATE INDEX IF NOT EXISTS idx_item_units_barcode ON item_units(barcode);

-- +goose Down
DROP TABLE IF EXISTS item_units;
