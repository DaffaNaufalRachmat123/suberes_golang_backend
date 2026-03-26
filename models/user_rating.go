package models

type UserRating struct {
	BaseModel
	OrderID        string  `gorm:"type:varchar(36)" json:"order_id"`
	CustomerID     string  `gorm:"type:varchar(36)" json:"customer_id"`
	MitraID        string  `gorm:"type:varchar(36)" json:"mitra_id"`
	LayananID      int     `gorm:"type:integer" json:"layanan_id"`
	ServiceID      int     `gorm:"type:integer" json:"service_id"`
	SubServiceID   int     `gorm:"type:integer" json:"sub_service_id"`
	Rating         float64 `gorm:"type:float" json:"rating"`
	Comment        string  `gorm:"type:varchar(255)" json:"comment"`
	RatingType     string  `gorm:"type:varchar(50);check:rating_type IN ('customer to mitra','mitra to customer')" json:"rating_type"`

	// Associations
	OrderTransaction *OrderTransaction `gorm:"foreignKey:OrderID;references:ID" json:"order_transaction,omitempty"`
	Customer         *User             `gorm:"foreignKey:CustomerID;references:ID" json:"customer,omitempty"`
	Mitra            *User             `gorm:"foreignKey:MitraID;references:ID" json:"mitra,omitempty"`
	LayananService   *LayananService   `gorm:"foreignKey:LayananID;references:ID" json:"layanan_service,omitempty"`
	Service          *Service          `gorm:"foreignKey:ServiceID;references:ID" json:"service,omitempty"`
	SubService       *SubService       `gorm:"foreignKey:SubServiceID;references:ID" json:"sub_service,omitempty"`
}

func (UserRating) TableName() string {
	return "users_ratings"
}
