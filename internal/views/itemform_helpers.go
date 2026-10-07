package views

import (
	"github.com/a-h/templ"
)

// ItemFormData dipakai items.templ & ItemForm.
type ItemFormData struct {
	User UserInfo
	Item *ItemFormValues
	Cats []string
	Locs []string
}

type ItemFormValues struct {
	ID              string
	SKU, Name       string
	Category        string
	Location        string
	CurrentStock    int
	MinStock        int
	Unit            string
	PricePerUnit    int64
	Description     string
	Merk            string
	TypeModel       string
	SerialNumber    string
	ProcurementYear string
	ConditionStatus string
	FundingSource   string
	Distributor     string
	AklAkd          string
}

var units = []string{"Unit", "Pcs", "Set", "Box", "Rim", "Pack", "Dus", "Botol", "Roll"}
var conditions = []string{"Berfungsi", "Rusak Ringan", "Rusak Berat", "Perlu Kalibrasi"}

func itemVal(v *ItemFormValues, f func(*ItemFormValues) string) string {
	if v == nil {
		return ""
	}
	return f(v)
}

func itemAction(it *ItemFormValues) templ.SafeURL {
	if it == nil {
		return templ.URL("/items")
	}
	return templ.URL("/items/" + it.ID)
}
