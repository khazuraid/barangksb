package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"inventariskantor/internal/views"
)

// BulkStockAdjust memproses CSV opname: kolom sku,fisik (2 kolom).
// Return: jumlah OK, jumlah gagal.
func (s *Service) BulkStockAdjust(ctx context.Context, r io.Reader, by string) (ok, fail int, errs []string) {
	rd := csv.NewReader(r)
	rd.TrimLeadingSpace = true
	rows, err := rd.ReadAll()
	if err != nil {
		return 0, 0, []string{"CSV tidak terbaca: " + err.Error()}
	}
	for i, row := range rows {
		if i == 0 && len(row) > 0 && strings.EqualFold(strings.TrimSpace(row[0]), "sku") {
			continue // header
		}
		if len(row) < 2 {
			fail++
			errs = append(errs, fmt.Sprintf("baris %d: kolom kurang", i+1))
			continue
		}
		sku := strings.TrimSpace(row[0])
		var physical int
		if _, err := fmt.Sscanf(strings.TrimSpace(row[1]), "%d", &physical); err != nil {
			fail++
			errs = append(errs, fmt.Sprintf("baris %d: angka tidak valid", i+1))
			continue
		}
		_, err := s.AdjustStock(ctx, sku, int32(physical), by, "opname massal")
		if err != nil {
			fail++
			errs = append(errs, fmt.Sprintf("baris %d (%s): %v", i+1, sku, err))
			continue
		}
		ok++
	}
	return ok, fail, errs
}

// ListAuditLog returns recent audit entries (F4).
func (s *Service) ListAuditLog(ctx context.Context, limit int) ([]views.AuditRow, error) {
	if limit < 1 || limit > 500 {
		limit = 100
	}
	rows, err := s.Pool.Query(ctx, `SELECT table_name, op, row_id, at FROM audit_log ORDER BY at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []views.AuditRow
	for rows.Next() {
		var a views.AuditRow
		if err := rows.Scan(&a.Table, &a.Op, &a.RowID, &a.At); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// AttachPhoto menghubungkan URL foto ke transaksi terakhir SKU (fitur bot).
func (s *Service) AttachPhoto(ctx context.Context, sku, photoURL string) error {
	_, err := s.Pool.Exec(ctx, `UPDATE stock_transactions SET photo_url=$2
		WHERE id = (SELECT id FROM stock_transactions WHERE item_sku=$1 ORDER BY timestamp DESC LIMIT 1)`,
		sku, photoURL)
	return err
}


