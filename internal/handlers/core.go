package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Core exposes domain operations for non-HTTP callers (Telegram bot, cron).
type Core struct {
	Pool *pgxpool.Pool
}

type StockInResult struct {
	SKU, Name, Unit string
	Previous, New   int32
}

// RecordStockIn performs the same atomic stock-in as the web form.
func (c *Core) RecordStockIn(ctx context.Context, sku string, qty int32, receivedBy, notes string) (*StockInResult, error) {
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var name, unit string
	var prev, next int32
	err = tx.QueryRow(ctx,
		`UPDATE inventory_items SET current_stock = current_stock + $2, updated_at = now()
		 WHERE sku = $1 RETURNING name, unit, current_stock - $2, current_stock`,
		sku, qty).Scan(&name, &unit, &prev, &next)
	if err != nil {
		return nil, fmt.Errorf("SKU %q tidak ditemukan", sku)
	}

	var itemID string
	_ = tx.QueryRow(ctx, `SELECT id FROM inventory_items WHERE sku=$1`, sku).Scan(&itemID)

	_, err = tx.Exec(ctx, `INSERT INTO stock_transactions
		(type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock, received_by, notes)
		VALUES ('IN', $1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		itemID, sku, name, qty, unit, prev, next,
		nullable(receivedBy), nullable(notes))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &StockInResult{SKU: sku, Name: name, Unit: unit, Previous: prev, New: next}, nil
}

type pgxTx = pgx.Tx

type SearchResult struct {
	Name, SKU, Unit, Location string
	Current                   int32
}

// AttachPhoto menghubungkan URL foto ke transaksi terakhir SKU (fitur bot).
func (c *Core) AttachPhoto(ctx context.Context, sku, photoURL string) error {
	_, err := c.Pool.Exec(ctx, `UPDATE stock_transactions SET photo_url=$2
		WHERE id = (SELECT id FROM stock_transactions WHERE item_sku=$1 ORDER BY timestamp DESC LIMIT 1)`,
		sku, &photoURL)
	return err
}

// SearchItems di core.go — versi lama dengan signature berbeda dihilangkan.

func (c *Core) SearchItems(ctx context.Context, keyword string, n int) []SearchResult {
	rows, err := c.Pool.Query(ctx,
		`SELECT name, sku, unit, location, current_stock FROM inventory_items
		 WHERE name ILIKE $1 OR sku ILIKE $1 OR location ILIKE $1
		 ORDER BY name LIMIT $2`, "%"+keyword+"%", n)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []SearchResult
	for rows.Next() {
		var r SearchResult
		if rows.Scan(&r.Name, &r.SKU, &r.Unit, &r.Location, &r.Current) == nil {
			out = append(out, r)
		}
	}
	return out
}

// LowStock returns items where current_stock <= min_stock.
func (c *Core) LowStock(ctx context.Context) []SearchResult {
	rows, err := c.Pool.Query(ctx,
		`SELECT name, sku, unit, location, current_stock FROM inventory_items
		 WHERE current_stock <= min_stock ORDER BY current_stock`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []SearchResult
	for rows.Next() {
		var r SearchResult
		if rows.Scan(&r.Name, &r.SKU, &r.Unit, &r.Location, &r.Current) == nil {
			out = append(out, r)
		}
	}
	return out
}

func nullable(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}
