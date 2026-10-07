package views

import (
	"time"
)

// AuditData untuk halaman audit log.
type AuditData struct {
	User  UserInfo
	Rows  []AuditRow
	Total int
	Pager PaginationData
}

type AuditRow struct {
	Table string
	Op    string
	RowID string
	At    time.Time
}
