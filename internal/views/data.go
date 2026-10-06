package views

import (
	"inventariskantor/internal/models"
)

type UserInfo struct {
	Name   string
	Role   string
	Active string // path aktif untuk highlight navbar
}

type ItemsData struct {
	User  UserInfo
	Items []ItemRow
	Query string
	Cat   string
	Loc   string
	Cats  []string
	Locs  []string
	Pager Pager
}

type ItemRow struct {
	ID           string
	SKU, Name    string
	Category     string
	Location     string
	CurrentStock int
	MinStock     int
	Unit         string
	Condition    string
}

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

type MovementData struct {
	User  UserInfo
	Cats  []string
	Locs  []string
	SKU   string
	Items []ItemRow
	Tab   string
}

type CategoriesData struct {
	User UserInfo
	Cats []models.Category
}

type HistoryData struct {
	User    UserInfo
	TX      []TXRow
	Pager   Pager
	FSKU    string
	FType   string
	FFrom   string
	FTo     string
	Types   []string
}

type TXRow struct {
	Time, Type, SKU, Name, Unit, ReceivedBy, Distributor, PONumber, Condition, Geo string
	Quantity                                                                      int
}

type BarcodeData struct {
	User  UserInfo
	Items []ItemRow
}

type DriveSyncData struct {
	User       UserInfo
	Configured bool
	LastRun    string
	Result     string
}
