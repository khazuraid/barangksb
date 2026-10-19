-- +goose Up
CREATE TABLE IF NOT EXISTS item_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_number TEXT NOT NULL UNIQUE,
    item_id UUID REFERENCES inventory_items(id) ON DELETE SET NULL,
    item_sku TEXT,
    item_name TEXT,
    reporter_name TEXT NOT NULL,
    reporter_contact TEXT NOT NULL,
    report_type TEXT NOT NULL DEFAULT 'Kerusakan',
    description TEXT NOT NULL,
    location TEXT,
    photo_url TEXT,
    status TEXT NOT NULL DEFAULT 'PENDING',
    admin_notes TEXT,
    handled_by TEXT,
    handled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_reports_created_at ON item_reports(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reports_status ON item_reports(status);
CREATE INDEX IF NOT EXISTS idx_reports_item_id ON item_reports(item_id);
CREATE INDEX IF NOT EXISTS idx_reports_number ON item_reports(report_number);

-- +goose Down
DROP TABLE IF EXISTS item_reports;
