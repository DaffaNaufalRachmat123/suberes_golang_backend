package models

import "time"

type ServicePromo struct {
	BaseModel
	ServiceID        int       `gorm:"type:integer" json:"service_id"`
	PromoName        string    `gorm:"type:varchar(255)" json:"promo_name"`
	PromoCount       int       `gorm:"type:integer" json:"promo_count"`
	PromoCategory    string    `gorm:"type:varchar(20);check:promo_category IN ('Discount','Cashback','Free Service')" json:"promo_category"`
	PromoPrice       int       `gorm:"type:integer" json:"promo_price"`
	PromoDescription string    `gorm:"type:text" json:"promo_description"`
	PromoStartDate   time.Time `gorm:"type:timestamp" json:"promo_start_date"`
	PromoEndDate     time.Time `gorm:"type:timestamp" json:"promo_end_date"`
	PromoImage       string    `gorm:"type:text" json:"promo_image"`

	// Associations
	Service *Service `gorm:"foreignKey:ServiceID;references:ID" json:"service,omitempty"`
}

func (ServicePromo) TableName() string {
	return "service_promos"
}
