package models

type MitraSearchQuery struct {
	Latitude             float64
	Longitude            float64
	IsActive              string
	IsBusy                string
	UserRating1           float64
	UserRating2           float64
	Distance              int
	IsAutoBid             string
	MinutesSubServices    int
	UserGender            string
	GrossAmountCompany    float64
	Page                  int
	Limit                 int
	CustomerID            string
	SubPaymentID          int
	OrderType             string
	ServiceDuration       int
	CustomerTimezoneCode  string
	CustomerTimeOrder     string
	JsonOrderTimes        string
	IsWithTime            bool
	InitialRange          int
	MaxRange              int
}

type MitraSearchResult struct {
	CountOrder       int     `gorm:"column:count_order"`
	CountOrderRepeat int     `gorm:"column:count_order_repeat"`
	ID               string  `gorm:"column:id"`
	FirebaseToken    string  `gorm:"column:firebase_token"`
	IsAutoBid        string  `gorm:"column:is_auto_bid"`
	AccountBalance   float64 `gorm:"column:account_balance"`
	TotalHutang      float64 `gorm:"column:total_hutang"`
	Distance         float64 `gorm:"column:distance"`
	MoneyLeft        float64 `gorm:"column:money_left"`
}
