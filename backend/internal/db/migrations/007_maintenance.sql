-- +goose Up
CREATE TABLE IF NOT EXISTS item_maintenances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
    service_type TEXT NOT NULL DEFAULT 'Servis Rutin',
    service_date DATE NOT NULL DEFAULT CURRENT_DATE,
    vendor_or_technician TEXT,
    cost BIGINT DEFAULT 0,
    description TEXT,
    next_service_date DATE,
    photo_url TEXT,
    performed_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_maint_item_id ON item_maintenances(item_id);
CREATE INDEX IF NOT EXISTS idx_maint_next_date ON item_maintenances(next_service_date);

-- +goose Down
DROP TABLE IF EXISTS item_maintenances;
