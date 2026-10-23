-- +goose Up
CREATE TABLE IF NOT EXISTS label_templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    width_mm NUMERIC(6,2) NOT NULL DEFAULT 64,
    height_mm NUMERIC(6,2) NOT NULL DEFAULT 32,
    layout_mode TEXT NOT NULL DEFAULT 'table',
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_label_templates_updated ON label_templates(updated_at DESC);

-- +goose Down
DROP TABLE IF EXISTS label_templates;
