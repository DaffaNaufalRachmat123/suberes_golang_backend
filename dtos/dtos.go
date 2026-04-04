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
type ServiceRequest struct {
	ParentID           int64  `json:"parent_id" binding:"required"`
	ServiceName        string `json:"service_name" binding:"required"`
	ServiceDescription string `json:"service_description" binding:"required"`
	ServiceType        string `json:"service_type" binding:"required"`
	ServiceCategory    string `json:"service_category" binding:"required"`
}
type ServiceUpdateRequest struct {
	ID                 int    `json:"id" binding:"required"`
	ServiceName        string `json:"service_name" binding:"required"`
	ServiceDescription string `json:"service_description" binding:"required"`
	ServiceCategory    string `json:"service_category" binding:"required,oneof=Cleaning Disinfectant Fogging Borongan Lainnya"`
	ServiceType        string `json:"service_type" binding:"required,oneof=Durasi 'Luas Ruangan' Single"`
}
type ServiceDeleteRequest struct {
	Password string `json:"password" binding:"required"`
}

// PaymentCreateRequest is used for POST /create and PUT /update/image/:id (multipart).
// Valid types: tunai, virtual account, ewallet, balance.
type PaymentCreateRequest struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	Desc  string `json:"desc"`
}

// PaymentUpdateRequest is used for PUT /update/:id (JSON body).
// Valid types: tunai, virtual account, transfer, balance.
type PaymentUpdateRequest struct {
	Title string `json:"title" binding:"required"`
	Type  string `json:"type" binding:"required"`
	Desc  string `json:"desc" binding:"required"`
}
