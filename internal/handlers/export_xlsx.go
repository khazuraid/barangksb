package handlers

import (
	"fmt"
	"net/http"

	"github.com/xuri/excelize/v2"
)

// ExportItemsXLSX streams master barang as .xlsx.
func (h *Handlers) ExportItemsXLSX(w http.ResponseWriter, r *http.Request) {
	items := h.allItems(r)
	f := excelize.NewFile()
	sheet := "Master_Barang"
	f.SetSheetName("Sheet1", sheet)

	for c, hd := range masterHeaders {
		cell, _ := excelize.CoordinatesToCellName(c+1, 1)
		f.SetCellValue(sheet, cell, hd)
	}
	for i, it := range items {
		row := i + 2
		vals := []any{
			boolAda(it.IsAvailable), optStr(it.SerialNumber), it.Name, it.Sku,
			dash(it.Merk), dash(it.TypeModel), dash(it.ProcurementYear),
			optStrOr(it.ConditionStatus, "Berfungsi"), it.Category, it.Location,
			it.CurrentStock, it.Unit, it.PricePerUnit.Int64,
			dash(it.FundingSource), dash(it.Distributor), dash(it.AklAkd),
			photoCell(it.PhotoUrl), geoCell(it.GeoLat, it.GeoLng),
			dash(it.Description), fmtWIB(it.UpdatedAt),
		}
		for c, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(c+1, row)
			f.SetCellValue(sheet, cell, v)
		}
	}
	// bold header
	style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetCellStyle(sheet, "A1", "T1", style)

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="master_barang.xlsx"`)
	if err := f.Write(w); err != nil {
		http.Error(w, fmt.Sprintf("xlsx: %v", err), http.StatusInternalServerError)
	}
}
