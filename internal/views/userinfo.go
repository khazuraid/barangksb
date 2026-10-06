package views

import (
	"strings"
)

// Initial returns the first rune of name untuk avatar.
func (u UserInfo) Initial() string {
	n := strings.TrimSpace(u.Name)
	if n == "" {
		return "?"
	}
	return strings.ToUpper(n[:1])
}

// ActiveTitle maps path ke judul topbar.
func (u UserInfo) ActiveTitle() string {
	titles := map[string]string{
		"/":            "Dashboard",
		"/items":       "Daftar Barang",
		"/items/new":   "Tambah Barang",
		"/categories":  "Kategori Barang",
		"/locations":   "Lokasi / Ruangan",
		"/movement":    "Mutasi Barang",
		"/history":     "Riwayat Mutasi",
		"/barcode":     "QR Generator",
		"/adjust/bulk": "Stok Opname Massal",
		"/drivesync":   "Sinkron Google Drive",
		"/users":       "Kelola Pengguna",
		"/audit":       "Audit Log",
		"/password":    "Ganti Password",
	}
	if t, ok := titles[u.Active]; ok {
		return t
	}
	if strings.HasPrefix(u.Active, "/items/") {
		return "Detail Barang"
	}
	return "Inventaris Kantor"
}
