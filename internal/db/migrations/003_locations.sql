-- +goose Up
CREATE TABLE IF NOT EXISTS locations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

-- migration: pindahkan lokasi unik dari inventory_items ke master
INSERT INTO locations (id, name)
SELECT DISTINCT lower(regexp_replace(location, '[^a-zA-Z0-9]+', '-', 'g')), location
FROM inventory_items
WHERE location IS NOT NULL AND location <> ''
ON CONFLICT (id) DO NOTHING;

ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS location_id TEXT REFERENCES locations(id);

-- isi location_id berdasarkan nama yang cocok
UPDATE inventory_items i SET location_id = l.id
FROM locations l WHERE i.location = l.name AND i.location_id IS NULL;

-- +goose Down
ALTER TABLE inventory_items DROP COLUMN IF EXISTS location_id;
DROP TABLE IF EXISTS locations;
