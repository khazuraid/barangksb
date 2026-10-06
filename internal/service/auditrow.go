package service

import (
	"time"
)

// AuditRow service-side copy (tanpa import views — hindari cycle).
type AuditRow struct {
	Table string
	Op    string
	RowID string
	At    time.Time
}
