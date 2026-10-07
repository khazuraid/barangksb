package views

import (
	"fmt"
	"time"
)

func fmtTimeA(t time.Time) string {
	return t.In(time.FixedZone("WIB", 7*3600)).Format("02-01-2006 15:04:05")
}

var _ = fmt.Sprintf
