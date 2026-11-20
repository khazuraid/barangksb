package service

import "strings"

// SanitizeSpreadsheetCell neutralizes CSV/XLSX formula injection payloads
// (values starting with =, +, -, @, tab, or carriage return).
func SanitizeSpreadsheetCell(s string) string {
	if s == "" {
		return s
	}
	trimmed := strings.TrimLeft(s, " \t\r\n")
	if trimmed == "" {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@', '\t', '\r', '\n':
		return "'" + s
	}
	switch trimmed[0] {
	case '=', '+', '-', '@':
		return "'" + s
	}
	return s
}
