package models

type Service struct {
	BaseModel
	ParentID              int
	ServiceName           string
	ServiceDescription    string
	ServiceImageThumbnail string
	ServiceCount          int
	ServiceType           string
	ServiceCategory       string
	IsResidental          string
	ServiceStatus         string
	IsActive              string
	MaxOrderCount         int
	PaymentTimeout        int

	SubServices []SubService `gorm:"foreignKey:ServiceID"`
}
