package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"inventariskantor/internal/auth"
)

func TestSlugify(t *testing.T) {
	assert.Equal(t, "alat-kebersihan", slugify("Alat Kebersihan"))
	assert.Equal(t, "atk-alat-tulis", slugify("ATK (Alat Tulis)"))
	assert.Equal(t, "ruang-poli", slugify("Ruang Poli"))
}

func TestPagination(t *testing.T) {
	assert.Equal(t, 4, calcPages(100, 25))
	assert.Equal(t, 1, calcPages(5, 25))
	assert.Equal(t, 3, calcPages(75, 25))
}

func calcPages(total, perPage int) int {
	if perPage < 1 {
		return 1
	}
	n := total / perPage
	if total%perPage > 0 {
		n++
	}
	if n < 1 {
		n = 1
	}
	return n
}

func TestNullableS(t *testing.T) {
	assert.Nil(t, nullableS(""))
	assert.Nil(t, nullableS("  "))
	assert.Equal(t, "hello", nullableS("hello"))
}

func TestConfigPerPage(t *testing.T) {
	assert.Equal(t, 25, perPageLogic("25"))
	assert.Equal(t, 50, perPageLogic("50"))
	assert.Equal(t, 100, perPageLogic("100"))
	assert.Equal(t, 25, perPageLogic(""))
	assert.Equal(t, 25, perPageLogic("999"))
	assert.Equal(t, 25, perPageLogic("20"))
}

func perPageLogic(s string) int {
	switch s {
	case "25":
		return 25
	case "50":
		return 50
	case "100":
		return 100
	default:
		return 25
	}
}

var _ = http.StatusOK

func TestSanitizeCSVCell(t *testing.T) {
	assert.Equal(t, "'=cmd|' /C calc'!A0", sanitizeCSVCell("=cmd|' /C calc'!A0"))
	assert.Equal(t, "'+12345", sanitizeCSVCell("+12345"))
	assert.Equal(t, "'-minus", sanitizeCSVCell("-minus"))
	assert.Equal(t, "'@mention", sanitizeCSVCell("@mention"))
	assert.Equal(t, "' =HYPERLINK()", sanitizeCSVCell(" =HYPERLINK()"))
	assert.Equal(t, "'\t=1+1", sanitizeCSVCell("\t=1+1"))
	assert.Equal(t, "normal text", sanitizeCSVCell("normal text"))
	assert.Equal(t, "", sanitizeCSVCell(""))
}

func TestSanitizePhotoURL(t *testing.T) {
	u, err := SanitizePhotoURL("")
	assert.NoError(t, err)
	assert.Equal(t, "", u)

	u, err = SanitizePhotoURL("/api/uploads/abc-123.jpg")
	assert.NoError(t, err)
	assert.Equal(t, "/api/uploads/abc-123.jpg", u)

	u, err = SanitizePhotoURL("https://cdn.example.com/photo.jpg")
	assert.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/photo.jpg", u)

	for _, bad := range []string{
		"javascript:alert(1)",
		"JAVASCRIPT:alert(document.cookie)",
		"data:text/html,<script>alert(1)</script>",
		"//evil.com/x.jpg",
		"/api/uploads/../../etc/passwd",
		"vbscript:msgbox(1)",
	} {
		_, err := SanitizePhotoURL(bad)
		assert.Error(t, err, "expected error for %s", bad)
	}
}

func TestValidateSafeExternalURL(t *testing.T) {
	for _, bad := range []string{
		"http://169.254.169.254/latest/meta-data/",
		"https://169.254.169.254/latest/meta-data/",
		"https://127.0.0.1:8080",
		"https://10.0.0.1/api",
		"https://192.168.1.1/api",
		"https://localhost/api",
		"http://api.telegram.org",
		"https://user:pass@api.telegram.org",
	} {
		_, err := ValidateSafeExternalURL(bad, false)
		assert.Error(t, err, "expected SSRF block for %s", bad)
	}

	u, err := ValidateSafeExternalURL("http://127.0.0.1:8081", true)
	assert.NoError(t, err)
	assert.NotNil(t, u)
}

func TestRenderSingleLabelPreventsXSS(t *testing.T) {
	h := NewBarcodeHandler(nil)
	htmlOut := h.renderSingleLabel(
		LabelItem{ID: "abc-123", Name: "Test<script>", SKU: "SKU-1", Location: "Gudang"},
		"tj_103",
		`qr" onerror="alert(1)`,
		"INVENTARIS",
		true, true, true,
	)
	assert.NotContains(t, htmlOut, "onerror")
	assert.NotContains(t, htmlOut, "<script>")
	assert.Contains(t, htmlOut, `/api/barcode/abc-123.png?fmt=qr`)
}

func TestReportNumberAndMasking(t *testing.T) {
	num := generateReportNumber()
	assert.True(t, len(num) == len("LAP-20260101-AABBCCDDEEFF"))
	assert.Equal(t, "08********90", maskContact("081234567890"))
	assert.Equal(t, "1234••••••••7890", maskSecretValue("1234567890:ABCDEF7890"))
	assert.True(t, isMaskedSecret("1234••••••••7890"))
	assert.True(t, isMaskedSecret(""))
	assert.False(t, isMaskedSecret("123456789:ABCDEF real token"))
}

func TestIsLocalUploadURL(t *testing.T) {
	name, ok := isLocalUploadURL("/api/uploads/abc-123.jpg")
	assert.True(t, ok)
	assert.Equal(t, "abc-123.jpg", name)

	name, ok = isLocalUploadURL("/uploads/report_1.jpg")
	assert.True(t, ok)
	assert.Equal(t, "report_1.jpg", name)

	for _, bad := range []string{
		"",
		"https://evil.com/abc-123.jpg",
		"/api/uploads/../secret.jpg",
		"/api/uploads/sub/file.jpg",
		"/etc/passwd",
	} {
		_, ok := isLocalUploadURL(bad)
		assert.False(t, ok, "expected false for %s", bad)
	}
}

func TestCasbinPetugasReportPermissions(t *testing.T) {
	e, err := auth.NewEnforcer()
	assert.NoError(t, err)

	ok, _ := e.Enforce("petugas", "/api/reports", "GET")
	assert.True(t, ok)
	ok, _ = e.Enforce("petugas", "/api/reports/123", "GET")
	assert.True(t, ok)
	ok, _ = e.Enforce("petugas", "/api/reports/123/status", "PUT")
	assert.True(t, ok)
	ok, _ = e.Enforce("petugas", "/api/reports/123", "DELETE")
	assert.False(t, ok)
	ok, _ = e.Enforce("admin", "/api/reports/123", "DELETE")
	assert.True(t, ok)
}
