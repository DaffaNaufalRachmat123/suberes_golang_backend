package dtos

type LoginMitraRequest struct {
	Email    string `json:"email" form:"email" binding:"required,email"`
	Password string `json:"password" form:"password" binding:"required"`
}

type RegisterMitraRequest struct {
	Email                       string `json:"email" form:"email" binding:"required,email"`
	CompleteName                string `json:"complete_name" form:"complete_name" binding:"required"`
	PlaceOfBirth                string `json:"place_of_birth" form:"place_of_birth" binding:"required"`
	Age                         string `json:"age" form:"age" binding:"required"`
	DateOfBirth                 string `json:"date_of_birth" form:"date_of_birth" binding:"required"`
	UserGender                  string `json:"user_gender" form:"user_gender" binding:"required,oneof=male female"`
	KTPNumber                   string `json:"ktp_number" form:"ktp_number" binding:"required"`
	PhoneNumber                 string `json:"phone_number" form:"phone_number" binding:"required"`
	CountryCode                 string `json:"country_code" form:"country_code" binding:"required"`
	EmergencyContactCountryCode string `json:"emergency_contact_country_code" form:"emergency_contact_country_code" binding:"required"`
	EmergencyContactName        string `json:"emergency_contact_name" form:"emergency_contact_name" binding:"required"`
	EmergencyContactRelation    string `json:"emergency_contact_relation" form:"emergency_contact_relation" binding:"required"`
	EmergencyContactPhone       string `json:"emergency_contact_phone" form:"emergency_contact_phone" binding:"required"`
	Address                     string `json:"address" form:"address" binding:"required"`
	DomisiliAddress             string `json:"domisili_address" form:"domisili_address" binding:"required"`
	RTRW                        string `json:"rt_rw" form:"rt_rw" binding:"required"`
	SubDistrict                 string `json:"sub_district" form:"sub_district" binding:"required"`
	District                    string `json:"district" form:"district" binding:"required"`
	Province                    string `json:"province" form:"province" binding:"required"`
	City                        string `json:"city" form:"city" binding:"required"`
	PostalCode                  string `json:"postal_code" form:"postal_code" binding:"required"`
	WorkExperience              string `json:"work_experience" form:"work_experience"`
	UserTool                    string `json:"user_tool" form:"user_tool"`
	IsExGoLife                  string `json:"is_ex_golife" form:"is_ex_golife" binding:"required,oneof=0 1"`
	KindOfMitra                 string `json:"kind_of_mitra" form:"kind_of_mitra" binding:"required,oneof=Dengan Alat,Tanpa Alat"`
	WorkExperienceDuration      string `json:"work_experience_duration" form:"work_experience_duration"`
}

type UpdateMitraRequest struct {
	MitraID                     uint   `json:"mitra_id" binding:"required"`
	Email                       string `json:"email" binding:"required,email"`
	CountryCode                 string `json:"country_code" binding:"required"`
	PhoneNumber                 string `json:"phone_number" binding:"required"`
	UserGender                  string `json:"user_gender" binding:"required"`
	DomisiliAddress             string `json:"domisili_address" binding:"required"`
	Address                     string `json:"address" binding:"required"`
	District                    string `json:"district" binding:"required"`
	SubDistrict                 string `json:"sub_district" binding:"required"`
	City                        string `json:"city" binding:"required"`
	PostalCode                  string `json:"postal_code" binding:"required"`
	IsExGoLife                  string `json:"is_ex_golife" binding:"required,oneof=0 1"`
	WorkExperienceDuration      string `json:"work_experience_duration"`
	EmergencyContactName        string `json:"emergency_contact_name" binding:"required"`
	EmergencyContactRelation    string `json:"emergency_contact_relation" binding:"required"`
	EmergencyContactCountryCode string `json:"emergency_contact_country_code" binding:"required"`
	EmergencyContactPhone       string `json:"emergency_contact_phone" binding:"required"`
}

type UpdateMitraCandidateRequest struct {
	Email                       string `json:"email" binding:"required,email"`
	CountryCode                 string `json:"country_code" binding:"required"`
	PhoneNumber                 string `json:"phone_number" binding:"required"`
	UserGender                  string `json:"user_gender" binding:"required,oneof=Pria,Wanita"`
	DomisiliAddress             string `json:"domisili_address" binding:"required"`
	Address                     string `json:"address" binding:"required"`
	District                    string `json:"district" binding:"required"`
	SubDistrict                 string `json:"sub_district" binding:"required"`
	KindOfMitra                 string `json:"kind_of_mitra" binding:"required"`
	KTPNumber                   string `json:"ktp_number" binding:"required"`
	Province                    string `json:"province" binding:"required"`
	RTRW                        string `json:"rt_rw" binding:"required"`
	City                        string `json:"city" binding:"required"`
	PlaceOfBirth                string `json:"place_of_birth" binding:"required"`
	DateOfBirth                 string `json:"date_of_birth" binding:"required"`
	PostalCode                  string `json:"postal_code" binding:"required"`
	IsExGoLife                  string `json:"is_ex_golife" binding:"required,oneof=Ya,Tidak"`
	WorkExperienceDuration      string `json:"work_experience_duration"`
	EmergencyContactName        string `json:"emergency_contact_name" binding:"required"`
	EmergencyContactCountryCode string `json:"emergency_contact_country_code" binding:"required"`
	EmergencyContactRelation    string `json:"emergency_contact_relation" binding:"required"`
	EmergencyContactPhone       string `json:"emergency_contact_phone" binding:"required"`
}

type ForgotPasswordRequest struct {
	ID       string `json:"id"`
	Email    string `json:"email" binding:"required"`
	OTPType  string `json:"otp_type" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SuspendRequest struct {
	SuspendedReason string `json:"suspended_reason" binding:"required"`
}

type DocumentStatusRequest struct {
	ID     string `json:"id" binding:"required"`
	Status string `json:"status" binding:"required,oneof=0 1"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type ChangeEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}
