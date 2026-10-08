-- +goose Up
CREATE TABLE IF NOT EXISTS telegram_commands (
    id SERIAL PRIMARY KEY,
    command TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    response TEXT NOT NULL DEFAULT '',
    is_menu BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- seed default commands
INSERT INTO telegram_commands (command, label, description, response, is_menu, sort_order) VALUES
 ('start',    'Menu Utama',  'Tampilkan menu utama',        'Selamat datang di Inventaris Kantor Bot!\nPilih menu di bawah:', true, 1),
 ('bantu',    'Bantuan',     'Daftar perintah',             'Perintah tersedia:\n/masuk SKU jumlah — catat masuk\n/cari kata — cari barang\n/stok — stok menipis', false, 2),
 ('masuk',    'Barang Masuk','Catat barang masuk',          'Format: /masuk <SKU> <jumlah>\nContoh: /masuk PERMED-2026-001 10', false, 3),
 ('cari',     'Cari Barang', 'Cari barang berdasarkan kata','Format: /cari <kata kunci>\nContoh: /cari tensimeter', false, 4),
 ('stok',     'Stok Menipis','Rekap barang stok menipis',   'Menampilkan daftar barang dengan stok di bawah batas minimum.', false, 5)
ON CONFLICT (command) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS telegram_commands;