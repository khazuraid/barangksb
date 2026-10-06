package views

import (
	"fmt"
	"sort"
	"github.com/a-h/templ"
)

func fmtInt(n int) string { return fmt.Sprintf("%d", n) }

func fmtInt64(n int64) string { return fmt.Sprintf("%d", n) }

func sortedKeys(m map[string]int) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func barWidth(v int, m map[string]int) string {
	max := 1
	for _, x := range m {
		if x > max {
			max = x
		}
	}
	return fmt.Sprintf("width: %d%%", v*100/max)
}

func lowClass(cur, min int) string {
	if cur <= min {
		return "warn"
	}
	return ""
}

func itemAction(it *ItemFormValues) templ.SafeURL {
	if it == nil {
		return templ.URL("/items")
	}
	return templ.URL("/items/" + it.ID)
}
