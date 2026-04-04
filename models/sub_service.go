package models

type SubService struct {
	BaseModel
	ServiceID             uint    `gorm:"type:integer" json:"service_id"`
	SubPriceServiceTitle  string  `gorm:"type:varchar(255)" json:"sub_price_service_title"`
	SubPriceService       int64   `gorm:"type:bigint" json:"sub_price_service"`
	SubServiceDescription string  `gorm:"type:varchar(255)" json:"sub_service_description"`
	CompanyPercentage     float64 `gorm:"type:float" json:"company_percentage"`
	MinutesSubServices    int     `gorm:"type:integer" json:"minutes_sub_services"`
	Criteria              string  `gorm:"type:varchar(255)" json:"criteria"`
	IsRecommended         string  `gorm:"type:varchar(1);check:is_recommended IN ('0','1')" json:"is_recommended"`

	// Associations
	Service                 *Service                 `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	OrderTransactions       []OrderTransaction       `gorm:"foreignKey:SubServiceID" json:"order_transactions,omitempty"`
	SubServiceAdditionals   []SubServiceAdditional   `gorm:"foreignKey:SubServiceID" json:"sub_service_additionals"`
	OrderTransactionRepeats []OrderTransactionRepeat `gorm:"foreignKey:SubServiceID" json:"order_transaction_repeats,omitempty"`
	UserRatings             []UserRating             `gorm:"foreignKey:SubServiceID" json:"user_ratings,omitempty"`
	OrderChats              []OrderChat              `gorm:"foreignKey:SubServiceID" json:"order_chats,omitempty"`
	Notifications           []Notification           `gorm:"foreignKey:SubServiceID" json:"notifications,omitempty"`
}
