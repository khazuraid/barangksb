package views

import (
	"fmt"
	"time"
)

type AuditData struct {
	User  UserInfo
	Rows  []AuditRow
	Total int
}

type AuditRow struct {
	Table string
	Op    string
	RowID string
	At    time.Time
}

func fmtTimeA(t time.Time) string {
	return t.In(time.FixedZone("WIB", 7*3600)).Format("02-01-2006 15:04:05")
}

var _ = fmt.Sprintf
