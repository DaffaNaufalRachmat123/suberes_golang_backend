package models

type OrderRepeat struct {
	BaseModel
	ID uint `gorm:"primaryKey;column:id" json:"id"`

	OrderID      string `gorm:"type:uuid;column:order_id" json:"order_id"`
	CustomerID   string `gorm:"size:36;column:customer_id" json:"customer_id"`
	MitraID      string `gorm:"size:36;column:mitra_id" json:"mitra_id"`
	ServiceID    int    `gorm:"column:service_id" json:"service_id"`
	SubServiceID int    `gorm:"column:sub_service_id" json:"sub_service_id"`
}

func (OrderRepeat) TableName() string {
	return "order_repeats"
}
