package views


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
	Pager PaginationData
}

type DriveSyncData struct {
	User       UserInfo
	Configured bool
	LastRun    string
	Result     string
}

