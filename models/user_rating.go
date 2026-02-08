package models

type UserRating struct {
	BaseModel // Pastikan struct BaseModel juga sudah sesuai tag-nya jika ada field di dalamnya

	OrderID      string  `gorm:"column:order_id" json:"order_id"`
	CustomerID   string  `gorm:"column:customer_id" json:"customer_id"`
	MitraID      string  `gorm:"column:mitra_id" json:"mitra_id"`
	ServiceID    int     `gorm:"column:service_id" json:"service_id"`
	SubServiceID int     `gorm:"column:sub_service_id" json:"sub_service_id"`
	Rating       float64 `gorm:"column:rating" json:"rating"`
	Comment      string  `gorm:"column:comment" json:"comment"`
	RatingType   string  `gorm:"column:rating_type" json:"rating_type"`
}

// Opsional: Jika ingin custom nama tabel
func (UserRating) TableName() string {
	return "user_ratings"
}
