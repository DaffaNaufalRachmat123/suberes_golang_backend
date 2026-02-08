package models

type SubService struct {
	BaseModel

	ServiceID             uint    `gorm:"column:service_id" json:"service_id"`
	SubPriceServiceTitle  string  `gorm:"column:sub_price_service_title" json:"sub_price_service_title"`
	SubPriceService       int     `gorm:"column:sub_price_service" json:"sub_price_service"`
	SubServiceDescription string  `gorm:"column:sub_service_description" json:"sub_service_description"`
	CompanyPercentage     float64 `gorm:"column:company_percentage" json:"company_percentage"`
	MinutesSubServices    int     `gorm:"column:minutes_sub_services" json:"minutes_sub_services"`
	Criteria              string  `gorm:"column:criteria" json:"criteria"`
	IsRecommended         string  `gorm:"column:is_recommended" json:"is_recommended"`
}
