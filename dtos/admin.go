package dtos

type CreateAdminRequest struct {
	CompleteName    string `json:"complete_name" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	PhoneNumber     string `json:"phone_number" binding:"required"`
	CountryCode     string `json:"country_code" binding:"required"`
	UserType        string `json:"user_type" binding:"required,oneof=superadmin admin"`
	UserGender      string `json:"user_gender" binding:"required"`
	Alamat          string `json:"alamat" binding:"required"`
	DomisiliAddress string `json:"domisili_address" binding:"required"`
}

type UpdateAdminStatusRequest struct {
	IsActive string `json:"is_active" binding:"required,oneof=yes no"`
	Reason   string `json:"reason" binding:"required"`
	UserType string `json:"user_type" binding:"required,oneof=superadmin admin"`
}

type RemoveAdminRequest struct {
	Password string `json:"password" binding:"required"`
	Reason   string `json:"reason" binding:"required"`
}

type LoginAdminRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateFirebaseTokenRequest struct {
	FirebaseToken string `json:"firebase_token" binding:"required"`
}
