package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// UserRow untuk halaman kelola user (tanpa hash).
type UserRow struct {
	ID        string
	Name      string
	Email     string
	Role      string
	CreatedAt pgtype.Timestamptz
}

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Location struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// InventoryItem mirror urutan kolom inventory_items (untuk RowToStructByPos).
type InventoryItem struct {
	ID              pgtype.UUID
	Sku             string
	Name            string
	Category        string
	Location        string
	CurrentStock    int32
	MinStock        int32
	Unit            string
	PricePerUnit    pgtype.Int8
	Description     pgtype.Text
	PhotoUrl        pgtype.Text
	GeoLat          pgtype.Float8
	GeoLng          pgtype.Float8
	GeoAcc          pgtype.Float8
	GeoAt           pgtype.Timestamptz
	GeoName         pgtype.Text
	DriveFileLink   pgtype.Text
	Merk            pgtype.Text
	TypeModel       pgtype.Text
	SerialNumber    pgtype.Text
	ProcurementYear pgtype.Text
	ConditionStatus pgtype.Text
	FundingSource   pgtype.Text
	Distributor     pgtype.Text
	AklAkd          pgtype.Text
	IsAvailable     bool
	BarcodeFormat   pgtype.Text
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	LocationID      pgtype.Text
}

// StockTransaction mirror urutan kolom stock_transactions.
type StockTransaction struct {
	ID               pgtype.UUID
	Type             string
	ItemID           pgtype.UUID
	ItemSku          string
	ItemName         string
	Quantity         int32
	Unit             string
	PreviousStock    int32
	NewStock         int32
	SupplierOrSource pgtype.Text
	ReceivedBy       pgtype.Text
	InvoiceOrPoNum   pgtype.Text
	PhotoUrl         pgtype.Text
	GeoLat           pgtype.Float8
	GeoLng           pgtype.Float8
	GeoAcc           pgtype.Float8
	GeoAt            pgtype.Timestamptz
	GeoName          pgtype.Text
	DriveFileLink    pgtype.Text
	Merk             pgtype.Text
	TypeModel        pgtype.Text
	SerialNumber     pgtype.Text
	ProcurementYear  pgtype.Text
	ConditionStatus  pgtype.Text
	FundingSource    pgtype.Text
	Distributor      pgtype.Text
	AklAkd           pgtype.Text
	Notes            pgtype.Text
	SyncedToDrive    pgtype.Bool
	Timestamp        pgtype.Timestamptz
}
