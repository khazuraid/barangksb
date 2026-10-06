package service

// Reorganisasi: satu service layer (Service struct) dipakai web handler,
// Telegram bot, dan scheduler. Query SQL hanya di sini.

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"inventariskantor/internal/models"
)

type Service struct {
	Pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service { return &Service{Pool: pool} }

// ---------- Items ----------

type ItemFilter struct {
	Query   string
	Cat     string
	Loc     string
	Page    int
	PerPage int
}

func (f *ItemFilter) offset() int {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 || f.PerPage > 200 {
		f.PerPage = 25
	}
	return (f.Page - 1) * f.PerPage
}

func (s *Service) ListItems(ctx context.Context, f ItemFilter) ([]models.InventoryItem, int, error) {
	where := " WHERE TRUE"
	args := []any{}
	if f.Query != "" {
		n := len(args) + 1
		where += fmt.Sprintf(` AND (name ILIKE $%d OR sku ILIKE $%d OR location ILIKE $%d)`, n, n, n)
		args = append(args, "%"+f.Query+"%")
	}
	if f.Cat != "" {
		where += fmt.Sprintf(` AND category = $%d`, len(args)+1)
		args = append(args, f.Cat)
	}
	if f.Loc != "" {
		where += fmt.Sprintf(` AND location = $%d`, len(args)+1)
		args = append(args, f.Loc)
	}

	var total int
	if err := s.Pool.QueryRow(ctx, `SELECT count(*) FROM inventory_items`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sql := `SELECT * FROM inventory_items` + where + ` ORDER BY name LIMIT ` +
		strconv.Itoa(f.PerPage) + ` OFFSET ` + strconv.Itoa(f.offset())
	rows, err := s.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.InventoryItem])
	return items, total, err
}

func (s *Service) GetItem(ctx context.Context, id string) (models.InventoryItem, error) {
	rows, err := s.Pool.Query(ctx, `SELECT * FROM inventory_items WHERE id=$1`, id)
	if err != nil {
		return models.InventoryItem{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByPos[models.InventoryItem])
}

func (s *Service) GetItemBySKU(ctx context.Context, sku string) (models.InventoryItem, error) {
	rows, err := s.Pool.Query(ctx, `SELECT * FROM inventory_items WHERE sku=$1`, sku)
	if err != nil {
		return models.InventoryItem{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByPos[models.InventoryItem])
}

func (s *Service) AllItems(ctx context.Context) ([]models.InventoryItem, error) {
	rows, err := s.Pool.Query(ctx, `SELECT * FROM inventory_items ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[models.InventoryItem])
}

type ItemInput struct {
	SKU, Name, Category, Location, Unit string
	CurrentStock, MinStock              int32
	PricePerUnit                        int64
	Description, PhotoURL               string
	Merk, TypeModel, SerialNumber       string
	ProcurementYear, ConditionStatus    string
	FundingSource, Distributor, AklAkd  string
}

func (s *Service) CreateItem(ctx context.Context, in ItemInput) (models.InventoryItem, error) {
	sku := strings.TrimSpace(in.SKU)
	if sku == "" {
		g, err := s.nextSKU(ctx, in.Category)
		if err != nil {
			return models.InventoryItem{}, err
		}
		sku = g
	}
	rows, err := s.Pool.Query(ctx, `INSERT INTO inventory_items
		(sku, name, category, location, current_stock, min_stock, unit, price_per_unit,
		 description, photo_url, merk, type_model, serial_number, procurement_year,
		 condition_status, funding_source, distributor, akl_akd, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18, now())
		RETURNING *`,
		sku, in.Name, in.Category, in.Location, in.CurrentStock, in.MinStock, in.Unit,
		nullableI8(in.PricePerUnit), nullableS(in.Description), nullableS(in.PhotoURL),
		nullableS(in.Merk), nullableS(in.TypeModel), nullableS(in.SerialNumber),
		nullableS(in.ProcurementYear), nullableS(in.ConditionStatus),
		nullableS(in.FundingSource), nullableS(in.Distributor), nullableS(in.AklAkd))
	if err != nil {
		return models.InventoryItem{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByPos[models.InventoryItem])
}

func (s *Service) UpdateItem(ctx context.Context, id string, in ItemInput) error {
	_, err := s.Pool.Exec(ctx, `UPDATE inventory_items SET
		sku=$2, name=$3, category=$4, location=$5, current_stock=$6, min_stock=$7, unit=$8,
		price_per_unit=$9, description=$10, photo_url=$11, merk=$12, type_model=$13,
		serial_number=$14, procurement_year=$15, condition_status=$16, funding_source=$17,
		distributor=$18, akl_akd=$19, updated_at=now()
		WHERE id=$1`,
		id, in.SKU, in.Name, in.Category, in.Location, in.CurrentStock, in.MinStock, in.Unit,
		nullableI8(in.PricePerUnit), nullableS(in.Description), nullableS(in.PhotoURL),
		nullableS(in.Merk), nullableS(in.TypeModel), nullableS(in.SerialNumber),
		nullableS(in.ProcurementYear), nullableS(in.ConditionStatus),
		nullableS(in.FundingSource), nullableS(in.Distributor), nullableS(in.AklAkd))
	return err
}

func (s *Service) DeleteItem(ctx context.Context, id string) error {
	_, err := s.Pool.Exec(ctx, `DELETE FROM inventory_items WHERE id=$1`, id)
	return err
}

// nextSKU: <KODEKATEGORI>-<TAHUN>-<3digit> berurutan per kategori per tahun.
func (s *Service) nextSKU(ctx context.Context, category string) (string, error) {
	var slug string
	err := s.Pool.QueryRow(ctx, `SELECT id FROM categories WHERE name=$1`, category).Scan(&slug)
	if err != nil {
		slug = Slugify(category)
	}
	parts := strings.Split(strings.Trim(slug, "-"), "-")
	prefix := strings.ToUpper(parts[0])
	if len(prefix) < 2 && len(slug) >= 3 {
		prefix = strings.ToUpper(Slugify(category))[:3]
	}
	if len(prefix) > 4 {
		prefix = prefix[:4]
	}

	year := time.Now().Format("2006")
	var last string
	err = s.Pool.QueryRow(ctx,
		`SELECT sku FROM inventory_items WHERE sku LIKE $1 ORDER BY sku DESC LIMIT 1`,
		prefix+"-"+year+"-%").Scan(&last)
	serial := 1
	// format: PREFIX-YYYY-NNN → offset serial = len(prefix)+1+4+1 = len(prefix)+6
	if err == nil && len(last) > len(prefix)+6 {
		if n, e := strconv.Atoi(last[len(prefix)+6:]); e == nil {
			serial = n + 1
		}
	}
	return fmt.Sprintf("%s-%s-%03d", prefix, year, serial), nil
}

// ---------- Movement (IN / OUT / ADJUST) ----------

type StockInResult struct {
	SKU, Name, Unit string
	Previous, New   int32
}

func (s *Service) RecordStockIn(ctx context.Context, sku string, qty int32, receivedBy, notes string) (*StockInResult, error) {
	return s.movement(ctx, "IN", sku, qty, receivedBy, notes)
}

func (s *Service) RecordStockOut(ctx context.Context, sku string, qty int32, issuedTo, notes string) (*StockInResult, error) {
	return s.movement(ctx, "OUT", sku, qty, issuedTo, notes)
}

func (s *Service) AdjustStock(ctx context.Context, sku string, actualStock int32, by, notes string) (*StockInResult, error) {
	it, err := s.GetItemBySKU(ctx, sku)
	if err != nil {
		return nil, fmt.Errorf("SKU %q tidak ditemukan", sku)
	}
	diff := actualStock - it.CurrentStock
	if diff == 0 {
		return &StockInResult{SKU: sku, Name: it.Name, Unit: it.Unit,
			Previous: it.CurrentStock, New: actualStock}, nil
	}
	txType := "ADJUST+"
	if diff < 0 {
		txType = "ADJUST-"
	}
	return s.movement(ctx, txType, sku, abs32(diff), by, "Opname: "+notes)
}

func (s *Service) movement(ctx context.Context, txType, sku string, qty int32, person, notes string) (*StockInResult, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var name, unit string
	var prev, next int32
	q := `UPDATE inventory_items SET current_stock = current_stock + $2, updated_at = now()
	      WHERE sku = $1 RETURNING name, unit, current_stock - $2, current_stock`
	if txType == "OUT" {
		q = `UPDATE inventory_items SET current_stock = current_stock - $2, updated_at = now()
		     WHERE sku = $1 AND current_stock >= $2 RETURNING name, unit, current_stock + $2, current_stock`
	}
	err = tx.QueryRow(ctx, q, sku, qty).Scan(&name, &unit, &prev, &next)
	if err != nil {
		if txType == "OUT" {
			return nil, fmt.Errorf("stok tidak cukup atau SKU %q tidak ditemukan", sku)
		}
		return nil, fmt.Errorf("SKU %q tidak ditemukan", sku)
	}

	var itemID string
	_ = tx.QueryRow(ctx, `SELECT id FROM inventory_items WHERE sku=$1`, sku).Scan(&itemID)

	_, err = tx.Exec(ctx, `INSERT INTO stock_transactions
		(type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock, received_by, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		txType, itemID, sku, name, qty, unit, prev, next, nullableS(person), nullableS(notes))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &StockInResult{SKU: sku, Name: name, Unit: unit, Previous: prev, New: next}, nil
}

// ---------- Transactions ----------

type TXFilter struct {
	SKU     string
	Type    string
	From    string // YYYY-MM-DD
	To      string
	Page    int
	PerPage int
}

func (f *TXFilter) offset() int {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 || f.PerPage > 500 {
		f.PerPage = 50
	}
	return (f.Page - 1) * f.PerPage
}

func (s *Service) ListTransactions(ctx context.Context, f TXFilter) ([]models.StockTransaction, int, error) {
	where := " WHERE TRUE"
	args := []any{}
	if f.SKU != "" {
		where += fmt.Sprintf(` AND item_sku ILIKE $%d`, len(args)+1)
		args = append(args, "%"+f.SKU+"%")
	}
	if f.Type != "" {
		where += fmt.Sprintf(` AND type = $%d`, len(args)+1)
		args = append(args, f.Type)
	}
	if f.From != "" {
		where += fmt.Sprintf(` AND timestamp >= $%d`, len(args)+1)
		args = append(args, f.From+" 00:00:00")
	}
	if f.To != "" {
		where += fmt.Sprintf(` AND timestamp <= $%d`, len(args)+1)
		args = append(args, f.To+" 23:59:59")
	}

	var total int
	if err := s.Pool.QueryRow(ctx, `SELECT count(*) FROM stock_transactions`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sql := `SELECT * FROM stock_transactions` + where + ` ORDER BY timestamp DESC LIMIT ` +
		strconv.Itoa(f.PerPage) + ` OFFSET ` + strconv.Itoa(f.offset())
	rows, err := s.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	txs, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.StockTransaction])
	return txs, total, err
}

// ---------- Categories ----------

func (s *Service) ListCategories(ctx context.Context) ([]models.Category, error) {
	rows, err := s.Pool.Query(ctx, `SELECT id, name FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Service) CategoryNames(ctx context.Context) ([]string, error) {
	rows, err := s.Pool.Query(ctx, `SELECT name FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[string])
}

// ---------- Locations ----------

func (s *Service) ListLocations(ctx context.Context) ([]models.Location, error) {
	rows, err := s.Pool.Query(ctx, `SELECT id, name FROM locations ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Location
	for rows.Next() {
		var l models.Location
		if err := rows.Scan(&l.ID, &l.Name); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *Service) LocationNames(ctx context.Context) ([]string, error) {
	rows, err := s.Pool.Query(ctx, `SELECT name FROM locations ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[string])
}

// ---------- Dashboard ----------

type DashStats struct {
	TotalItems int
	TotalStock int
	LowStock   []SearchRow
	StockByCat map[string]int
	CatCount   int
}

type SearchRow struct {
	Name, SKU, Unit, Location string
	Current                   int32
}

func (s *Service) Dashboard(ctx context.Context) (*DashStats, error) {
	st := &DashStats{StockByCat: map[string]int{}}
	if err := s.Pool.QueryRow(ctx, `SELECT count(*), COALESCE(sum(current_stock),0) FROM inventory_items`).
		Scan(&st.TotalItems, &st.TotalStock); err != nil {
		return nil, err
	}
	rows, err := s.Pool.Query(ctx, `SELECT name, sku, unit, location, current_stock FROM inventory_items
		WHERE current_stock <= min_stock ORDER BY current_stock`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var r SearchRow
		if rows.Scan(&r.Name, &r.SKU, &r.Unit, &r.Location, &r.Current) == nil {
			st.LowStock = append(st.LowStock, r)
		}
	}
	rows.Close()

	crows, err := s.Pool.Query(ctx, `SELECT category, COALESCE(sum(current_stock),0) FROM inventory_items GROUP BY category`)
	if err != nil {
		return nil, err
	}
	for crows.Next() {
		var c string
		var n int
		if crows.Scan(&c, &n) == nil {
			st.StockByCat[c] = n
		}
	}
	crows.Close()
	st.CatCount = len(st.StockByCat)
	return st, nil
}

// AllTX returns all transactions (untuk PDF bulanan).
func (s *Service) AllTX(ctx context.Context) ([]models.StockTransaction, error) {
	rows, err := s.Pool.Query(ctx, `SELECT * FROM stock_transactions ORDER BY timestamp DESC`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[models.StockTransaction])
}

// AllTransactions sama dengan AllTX.
func (s *Service) AllTransactions(ctx context.Context) ([]models.StockTransaction, error) {
	return s.AllTX(ctx)
}

func nullableS(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func nullableI8(n int64) any {
	if n == 0 {
		return nil
	}
	return n
}

func abs32(n int32) int32 {
	if n < 0 {
		return -n
	}
	return n
}

// Slugify converts a name into a URL/ID-safe slug (dedup dashes).
func Slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prevDash := false
	for _, c := range s {
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			b.WriteRune(c)
			prevDash = false
		case c == ' ' || c == '-' || c == '/' || c == '&' || c == '(' || c == ')':
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

var _ = pgtype.Text{}
