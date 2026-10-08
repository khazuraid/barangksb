-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'petugas' CHECK (role IN ('admin','petugas')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS locations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS inventory_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    location TEXT NOT NULL,
    current_stock INTEGER NOT NULL DEFAULT 0,
    min_stock INTEGER NOT NULL DEFAULT 0,
    unit TEXT NOT NULL,
    price_per_unit BIGINT DEFAULT 0,
    description TEXT,
    photo_url TEXT,
    geo_lat DOUBLE PRECISION, geo_lng DOUBLE PRECISION, geo_acc DOUBLE PRECISION,
    geo_at TIMESTAMPTZ, geo_name TEXT,
    merk TEXT, type_model TEXT, serial_number TEXT,
    procurement_year TEXT,
    condition_status TEXT DEFAULT 'Berfungsi',
    funding_source TEXT, distributor TEXT, akl_akd TEXT,
    is_available BOOLEAN NOT NULL DEFAULT TRUE,
    barcode_format TEXT DEFAULT 'CODE128',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS stock_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type TEXT NOT NULL DEFAULT 'IN',
    item_id UUID NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
    item_sku TEXT NOT NULL, item_name TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit TEXT NOT NULL,
    previous_stock INTEGER NOT NULL, new_stock INTEGER NOT NULL,
    supplier_or_source TEXT, received_by TEXT, invoice_or_po_number TEXT,
    photo_url TEXT,
    geo_lat DOUBLE PRECISION, geo_lng DOUBLE PRECISION, geo_acc DOUBLE PRECISION,
    geo_at TIMESTAMPTZ, geo_name TEXT,
    merk TEXT, type_model TEXT, serial_number TEXT,
    procurement_year TEXT, condition_status TEXT,
    funding_source TEXT, distributor TEXT, akl_akd TEXT,
    notes TEXT,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tx_timestamp ON stock_transactions(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_items_category ON inventory_items(category);
CREATE INDEX IF NOT EXISTS idx_items_location ON inventory_items(location);

INSERT INTO categories(id, name) VALUES
 ('peralatan-medis','Peralatan Medis & Alkes (AKL/AKD)'),
 ('elektronik-it','Elektronik & IT Perkantoran'),
 ('furnitur','Furnitur & Perlengkapan Ruangan'),
 ('atk','ATK (Alat Tulis Kantor)'),
 ('alat-lab','Alat Laboratorium & Diagnostik'),
 ('pantri','Pantri & Fasilitas Umum'),
 ('kebersihan','Kebersihan & Sanitasi'),
 ('keamanan-k3','Keamanan & K3'),
 ('lainnya','Lainnya')
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS stock_transactions;
DROP TABLE IF EXISTS inventory_items;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;
