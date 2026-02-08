package models

type OrderOffers struct {
	BaseModel // Asumsi punya ID, CreatedAt, UpdatedAt

	TempID     string `gorm:"column:temp_id" json:"temp_id"`
	OrderID    string `gorm:"type:uuid;column:order_id" json:"order_id"`
	CustomerID string `gorm:"size:36;column:customer_id" json:"customer_id"`
	MitraID    string `gorm:"size:36;column:mitra_id" json:"mitra_id"`
}

func (OrderOffers) TableName() string {
	return "order_offers"
}
