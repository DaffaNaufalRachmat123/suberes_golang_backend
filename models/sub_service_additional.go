package models

type SubServiceAdditional struct {
	BaseModel

	SubServiceID   int     `gorm:"column:sub_service_id" json:"sub_service_id"`
	Title          string  `gorm:"column:title" json:"title"`
	BaseAmount     float64 `gorm:"column:base_amount" json:"base_amount"`
	Amount         float64 `gorm:"column:amount" json:"amount"`
	AdditionalType string  `gorm:"column:additional_type" json:"additional_type"`
}

func (SubServiceAdditional) TableName() string {
	return "sub_service_additional"
}
