package handlers

import (
	"net/url"

	"inventariskantor/internal/models"
	"inventariskantor/internal/views"
)

type modelsInventoryItem = models.InventoryItem

func toFormValues(it *models.InventoryItem) *views.ItemFormValues {
	return &views.ItemFormValues{
		ID:              it.ID.String(),
		SKU:             it.Sku,
		Name:            it.Name,
		Category:        it.Category,
		Location:        it.Location,
		CurrentStock:    int(it.CurrentStock),
		MinStock:        int(it.MinStock),
		Unit:            it.Unit,
		PricePerUnit:    it.PricePerUnit.Int64,
		Description:     optStr(it.Description),
		Merk:            optStr(it.Merk),
		TypeModel:       optStr(it.TypeModel),
		SerialNumber:    optStr(it.SerialNumber),
		ProcurementYear: optStr(it.ProcurementYear),
		ConditionStatus: optStrOr(it.ConditionStatus, "Berfungsi"),
		FundingSource:   optStr(it.FundingSource),
		Distributor:     optStr(it.Distributor),
		AklAkd:          optStr(it.AklAkd),
	}
}

var _ = url.Values{}
