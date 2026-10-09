package handler

import (
	"context"
	"encoding/csv"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BulkHandler struct {
	pool *pgxpool.Pool
	tg   TelegramNotifier
}

func NewBulkHandler(pool *pgxpool.Pool, tg TelegramNotifier) *BulkHandler {
	return &BulkHandler{pool: pool, tg: tg}
}

// POST /api/adjust/bulk (multipart: file CSV with sku,fisik)
func (h *BulkHandler) BulkAdjust(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file CSV wajib"})
		return
	}
	defer file.Close()

	rd := csv.NewReader(file)
	rd.TrimLeadingSpace = true
	rows, err := rd.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV tidak terbaca"})
		return
	}

	ok, fail := 0, 0
	errs := []string{}
	person := c.GetString("name")

	for i, row := range rows {
		if i == 0 && len(row) > 0 && strings.EqualFold(strings.TrimSpace(row[0]), "sku") {
			continue
		}
		if len(row) < 2 {
			fail++
			errs = append(errs, "baris "+itoa(i+1)+": kolom kurang")
			continue
		}
		sku := strings.TrimSpace(row[0])
		var physical int32
		_, err := fmtScanf(&physical, row[1])
		if err != nil {
			fail++
			errs = append(errs, "baris "+itoa(i+1)+": angka tidak valid")
			continue
		}

		err = h.adjustSingle(c.Request.Context(), sku, physical, person)
		if err != nil {
			fail++
			errs = append(errs, "baris "+itoa(i+1)+" ("+sku+"): "+err.Error())
			continue
		}
		ok++
	}

	c.JSON(http.StatusOK, gin.H{"ok": ok, "fail": fail, "errors": errs})
}

func (h *BulkHandler) adjustSingle(ctx context.Context, sku string, actual int32, person string) error {
	var current int32
	var name, unit string
	err := h.pool.QueryRow(ctx, `SELECT current_stock, name, unit FROM inventory_items WHERE sku=$1`, sku).
		Scan(&current, &name, &unit)
	if err != nil {
		return err
	}
	diff := actual - current
	if diff == 0 {
		return nil
	}
	txType := "ADJUST+"
	absDiff := diff
	if diff < 0 {
		txType = "ADJUST-"
		absDiff = -diff
	}
	h.pool.Exec(ctx, `UPDATE inventory_items SET current_stock=$2, updated_at=now() WHERE sku=$1`, sku, actual)
	h.pool.Exec(ctx, `INSERT INTO stock_transactions (type, item_id, item_sku, item_name, quantity, unit, previous_stock, new_stock, received_by, notes)
		VALUES ($1, (SELECT id FROM inventory_items WHERE sku=$2), $2, $3, $4, $5, $6, $7, $8, $9)`,
		txType, sku, name, absDiff, unit, current, actual, nullableS(person), nullableS("Opname massal"))

	if h.tg != nil {
		go h.tg.NotifyMovement(context.Background(), txType, sku, name, absDiff, unit, current, actual, person, "Opname massal")
	}
	return nil
}

func itoa(n int) string {
	return fmtSprintf("%d", n)
}

func fmtScanf(target *int32, s string) (int, error) {
	n, err := parseInt(s)
	*target = n
	return 0, err
}

func parseInt(s string) (int32, error) {
	s = strings.TrimSpace(s)
	n := int32(0)
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errInvalidNum
		}
		n = n*10 + int32(c-'0')
	}
	return n, nil
}

var errInvalidNum = httpError("invalid number")

func httpError(msg string) error { return &simpleErr{msg: msg} }

type simpleErr struct{ msg string }

func (e *simpleErr) Error() string { return e.msg }

var _ io.Reader = (*strings.Reader)(nil)
