package views

import (
	"github.com/a-h/templ"
)

// SidebarItem satu entri menu.
type SidebarItem struct {
	Href  string
	Icon  string
	Label string
	Admin bool
}

// sidebarGroup: grup menu, Admin=true berarti hanya untuk role admin.
type sidebarGroup struct {
	Label string
	Items []SidebarItem
	Admin bool
}

var sidebarGroups = []sidebarGroup{
	{"Menu Utama", []SidebarItem{
		{"/", "ti-dashboard", "Dashboard", false},
		{"/items", "ti-box-seam", "Barang", false},
		{"/categories", "ti-category-2", "Kategori", false},
		{"/locations", "ti-map-pin", "Lokasi", false},
	}, false},
	{"Transaksi", []SidebarItem{
		{"/movement", "ti-arrows-exchange", "Mutasi Barang", false},
		{"/history", "ti-clock-hour-4", "Riwayat", false},
	}, false},
	{"Alat", []SidebarItem{
		{"/barcode", "ti-qrcode", "QR Generator", false},
		{"/adjust/bulk", "ti-clipboard-list", "Opname Massal", false},
		{"/drivesync", "ti-cloud-download", "Sinkron Drive", false},
	}, false},
	{"Admin", []SidebarItem{
		{"/users", "ti-users-group", "Pengguna", true},
		{"/audit", "ti-shield-check", "Audit Log", true},
	}, true},
}

var _ = templ.URL
