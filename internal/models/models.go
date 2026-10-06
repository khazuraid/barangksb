package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type InventoryItem struct {
	ID              pgtype.UUID        `json:"id"`
	Sku             string             `json:"sku"`
	Name            string             `json:"name"`
	Category        string             `json:"category"`
	Location        string             `json:"location"`
	CurrentStock    int32              `json:"current_stock"`
	MinStock        int32              `json:"min_stock"`
	Unit            string             `json:"unit"`
	PricePerUnit    pgtype.Int8        `json:"price_per_unit"`
	Description     pgtype.Text        `json:"description"`
	PhotoUrl        pgtype.Text        `json:"photo_url"`
	GeoLat          pgtype.Float8      `json:"geo_lat"`
	GeoLng          pgtype.Float8      `json:"geo_lng"`
	GeoAcc          pgtype.Float8      `json:"geo_acc"`
	GeoAt           pgtype.Timestamptz `json:"geo_at"`
	GeoName         pgtype.Text        `json:"geo_name"`
	DriveFileLink   pgtype.Text        `json:"drive_file_link"`
	Merk            pgtype.Text        `json:"merk"`
	TypeModel       pgtype.Text        `json:"type_model"`
	SerialNumber    pgtype.Text        `json:"serial_number"`
	ProcurementYear pgtype.Text        `json:"procurement_year"`
	ConditionStatus pgtype.Text        `json:"condition_status"`
	FundingSource   pgtype.Text        `json:"funding_source"`
	Distributor     pgtype.Text        `json:"distributor"`
	AklAkd          pgtype.Text        `json:"akl_akd"`
	IsAvailable     bool               `json:"is_available"`
	BarcodeFormat   pgtype.Text        `json:"barcode_format"`
	CreatedAt       pgtype.Timestamptz `json:"created_at"`
	UpdatedAt       pgtype.Timestamptz `json:"updated_at"`
}

type StockTransaction struct {
	ID               pgtype.UUID        `json:"id"`
	Type             string             `json:"type"`
	ItemID           pgtype.UUID        `json:"item_id"`
	ItemSku          string             `json:"item_sku"`
	ItemName         string             `json:"item_name"`
	Quantity         int32              `json:"quantity"`
	Unit             string             `json:"unit"`
	PreviousStock    int32              `json:"previous_stock"`
	NewStock         int32              `json:"new_stock"`
	SupplierOrSource pgtype.Text        `json:"supplier_or_source"`
	ReceivedBy       pgtype.Text        `json:"received_by"`
	InvoiceOrPoNum   pgtype.Text        `json:"invoice_or_po_number"`
	PhotoUrl         pgtype.Text        `json:"photo_url"`
	GeoLat           pgtype.Float8      `json:"geo_lat"`
	GeoLng           pgtype.Float8      `json:"geo_lng"`
	GeoAcc           pgtype.Float8      `json:"geo_acc"`
	GeoAt            pgtype.Timestamptz `json:"geo_at"`
	GeoName          pgtype.Text        `json:"geo_name"`
	DriveFileLink    pgtype.Text        `json:"drive_file_link"`
	Merk             pgtype.Text        `json:"merk"`
	TypeModel        pgtype.Text        `json:"type_model"`
	SerialNumber     pgtype.Text        `json:"serial_number"`
	ProcurementYear  pgtype.Text        `json:"procurement_year"`
	ConditionStatus  pgtype.Text        `json:"condition_status"`
	FundingSource    pgtype.Text        `json:"funding_source"`
	Distributor      pgtype.Text        `json:"distributor"`
	AklAkd           pgtype.Text        `json:"akl_akd"`
	Notes            pgtype.Text        `json:"notes"`
	SyncedToDrive    pgtype.Bool        `json:"synced_to_drive"`
	Timestamp        pgtype.Timestamptz `json:"timestamp"`
}

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
