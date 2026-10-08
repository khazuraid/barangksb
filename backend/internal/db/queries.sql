-- name: ListItems :many
SELECT id, sku, name, category, location, current_stock, min_stock, unit,
       price_per_unit, condition_status, is_available, created_at, updated_at
FROM inventory_items
WHERE
  (sqlc.arg('query')::text = '' OR
   name ILIKE '%' || sqlc.arg('query')::text || '%' OR
   sku ILIKE '%' || sqlc.arg('query')::text || '%' OR
   location ILIKE '%' || sqlc.arg('query')::text || '%')
  AND (sqlc.arg('category')::text = '' OR category = sqlc.arg('category')::text)
  AND (sqlc.arg('location')::text = '' OR location = sqlc.arg('location')::text)
ORDER BY name
LIMIT sqlc.arg('limit')::int OFFSET sqlc.arg('offset')::int;

-- name: CountItems :one
SELECT count(*) FROM inventory_items
WHERE
  (sqlc.arg('query')::text = '' OR
   name ILIKE '%' || sqlc.arg('query')::text || '%' OR
   sku ILIKE '%' || sqlc.arg('query')::text || '%' OR
   location ILIKE '%' || sqlc.arg('query')::text || '%')
  AND (sqlc.arg('category')::text = '' OR category = sqlc.arg('category')::text)
  AND (sqlc.arg('location')::text = '' OR location = sqlc.arg('location')::text);

-- name: GetItem :one
SELECT * FROM inventory_items WHERE id = $1;

-- name: GetItemBySKU :one
SELECT * FROM inventory_items WHERE sku = $1;

-- name: CreateItem :one
INSERT INTO inventory_items
  (sku, name, category, location, current_stock, min_stock, unit, price_per_unit,
   description, photo_url, merk, type_model, serial_number, procurement_year,
   condition_status, funding_source, distributor, akl_akd)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
RETURNING *;

-- name: UpdateItem :exec
UPDATE inventory_items SET
  sku=$2, name=$3, category=$4, location=$5, current_stock=$6, min_stock=$7,
  unit=$8, price_per_unit=$9, description=$10, photo_url=$11, merk=$12,
  type_model=$13, serial_number=$14, procurement_year=$15, condition_status=$16,
  funding_source=$17, distributor=$18, akl_akd=$19, updated_at=now()
WHERE id=$1;

-- name: DeleteItem :exec
DELETE FROM inventory_items WHERE id=$1;

-- name: IncrementStock :one
UPDATE inventory_items SET current_stock = current_stock + $2, updated_at = now()
WHERE sku = $1 RETURNING name, unit, current_stock - $2, current_stock;

-- name: DecrementStock :one
UPDATE inventory_items SET current_stock = current_stock - $2, updated_at = now()
WHERE sku = $1 AND current_stock >= $2
RETURNING name, unit, current_stock + $2, current_stock;

-- name: CreateTransaction :exec
INSERT INTO stock_transactions
  (type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock,
   received_by, notes, photo_url, geo_lat, geo_lng, geo_acc, geo_at, geo_name)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16);

-- name: ListTransactions :many
SELECT * FROM stock_transactions
WHERE
  (sqlc.arg('sku')::text = '' OR item_sku ILIKE '%' || sqlc.arg('sku')::text || '%')
  AND (sqlc.arg('type')::text = '' OR type = sqlc.arg('type')::text)
ORDER BY timestamp DESC
LIMIT sqlc.arg('limit')::int OFFSET sqlc.arg('offset')::int;

-- name: CountTransactions :one
SELECT count(*) FROM stock_transactions
WHERE
  (sqlc.arg('sku')::text = '' OR item_sku ILIKE '%' || sqlc.arg('sku')::text || '%')
  AND (sqlc.arg('type')::text = '' OR type = sqlc.arg('type')::text);

-- name: ListCategories :many
SELECT id, name FROM categories ORDER BY name;

-- name: CreateCategory :exec
INSERT INTO categories (id, name) VALUES ($1,$2)
ON CONFLICT (id) DO UPDATE SET name=$2;

-- name: RenameCategory :exec
UPDATE categories SET name=$2 WHERE id=$1;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id=$1;

-- name: ListLocations :many
SELECT id, name FROM locations ORDER BY name;

-- name: CreateLocation :exec
INSERT INTO locations (id, name) VALUES ($1,$2)
ON CONFLICT (id) DO UPDATE SET name=$2;

-- name: RenameLocation :exec
UPDATE locations SET name=$2 WHERE id=$1;

-- name: DeleteLocation :exec
DELETE FROM locations WHERE id=$1;

-- name: GetUserByEmail :one
SELECT id, name, email, role, password_hash FROM users WHERE email=$1;

-- name: ListUsers :many
SELECT id, name, email, role, created_at FROM users ORDER BY created_at;

-- name: CreateUser :exec
INSERT INTO users (name, email, password_hash, role)
VALUES ($1,$2,$3,$4)
ON CONFLICT (email) DO UPDATE SET name=$1, password_hash=$3, role=$4;

-- name: UpdateUserRole :exec
UPDATE users SET role=$2 WHERE email=$1;

-- name: DeleteUser :exec
DELETE FROM users WHERE email=$1;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash=$2 WHERE email=$1;

-- name: ListAuditLog :many
SELECT table_name, op, row_id, at FROM audit_log ORDER BY at DESC LIMIT $1;

-- name: CountAuditLog :one
SELECT count(*) FROM audit_log;

-- name: DashboardStats :one
SELECT count(*), COALESCE(sum(current_stock),0) FROM inventory_items;

-- name: DashboardLowStock :many
SELECT name, sku, current_stock, min_stock, unit, location
FROM inventory_items WHERE current_stock <= min_stock ORDER BY current_stock LIMIT 10;

-- name: DashboardRecentTx :many
SELECT timestamp, item_name, item_sku, quantity, unit, type
FROM stock_transactions ORDER BY timestamp DESC LIMIT 8;

-- name: DashboardStockByCat :many
SELECT category, COALESCE(sum(current_stock),0) FROM inventory_items GROUP BY category;

-- name: AllItems :many
SELECT * FROM inventory_items ORDER BY name;

-- name: AllTransactions :many
SELECT * FROM stock_transactions ORDER BY timestamp DESC;
