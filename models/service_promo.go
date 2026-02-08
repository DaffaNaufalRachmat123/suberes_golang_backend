package models

import "time"

type ServicePromo struct {
	BaseModel

	ServiceID        int       `gorm:"column:service_id" json:"service_id"`
	PromoName        string    `gorm:"column:promo_name" json:"promo_name"`
	PromoCount       int       `gorm:"column:promo_count" json:"promo_count"`
	PromoCategory    string    `gorm:"type:enum('Discount','Cashback','Free Service');column:promo_category" json:"promo_category"`
	PromoPrice       int       `gorm:"column:promo_price" json:"promo_price"`
	PromoDescription string    `gorm:"type:text;column:promo_description" json:"promo_description"`
	PromoStartDate   time.Time `gorm:"column:promo_start_date" json:"promo_start_date"`
	PromoEndDate     time.Time `gorm:"column:promo_end_date" json:"promo_end_date"`
	PromoImage       string    `gorm:"type:text;column:promo_image" json:"promo_image"`
}

func (ServicePromo) TableName() string {
	return "service_promos"
}
