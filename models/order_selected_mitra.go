package models

type OrderSelectedMitra struct {
	BaseModel

	OrderID     string `gorm:"size:36;column:order_id" json:"order_id"`
	MitraID     string `gorm:"size:36;column:mitra_id" json:"mitra_id"`
	OfferStatus string `gorm:"type:enum('SELECTED','CANCELED','RECEIVED','ACCEPTED','REJECTED');column:offer_status" json:"offer_status"`
}

func (OrderSelectedMitra) TableName() string {
	return "order_selected_mitras"
}
