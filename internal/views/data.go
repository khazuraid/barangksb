package views

import (
	"inventariskantor/internal/models"
)

type UserInfo struct {
	Name   string
	Role   string
	Active string // path aktif untuk highlight navbar
}





type MovementData struct {
	User  UserInfo
	Cats  []string
	Locs  []string
	SKU   string
	Items []ItemRow
	Tab   string
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

type ItemsData struct {
	User  UserInfo
	Items []ItemRow
	Query string
	Cat   string
	Loc   string
	Cats  []string
	Locs  []string
	Pager PaginationData
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

type LocationsData struct {
	User UserInfo
	Locs []models.Location
}
