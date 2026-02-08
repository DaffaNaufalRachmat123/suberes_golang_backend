package dtos

// BannerRequest adalah representasi JSON input dari Frontend.
// Tag 'binding' berfungsi sebagai validator (pengganti Joi).
type BannerRequest struct {
	CreatorID   string `json:"creator_id" binding:"required"`
	CreatorName string `json:"creator_name" binding:"required"`
	BannerTitle string `json:"banner_title" binding:"required"`
	BannerBody  string `json:"banner_body" binding:"required"`

	// Validasi enum: hanya boleh promo, coupon, visi misi, atau other
	BannerType string `json:"banner_type" binding:"required,oneof=promo coupon 'visi misi' other"`

	// Validasi: hanya boleh string "0" atau "1"
	IsBroadcast string `json:"is_broadcast" binding:"required,oneof=0 1"`
}
type LayananServiceRequest struct {
	LayananTitle       string `json:"layanan_title" binding:"required"`
	LayananDescription string `json:"layanan_description" binding:"required"`
	IsActive           string `json:"is_active" binding:"required,oneof=0 1"`
}
