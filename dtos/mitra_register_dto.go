package dtos

type MitraRegisterDTO struct {
	Email                      string `json:"email" binding:"required,email"`
	CompleteName               string `json:"complete_name" binding:"required"`
	PlaceOfBirth               string `json:"place_of_birth" binding:"required"`
	Age                        string `json:"age" binding:"required"`
	DateOfBirth                string `json:"date_of_birth" binding:"required"`
	UserGender                 string `json:"user_gender" binding:"required,oneof=male female"`
	KTPNumber                  string `json:"ktp_number" binding:"required"`
	PhoneNumber                string `json:"phone_number" binding:"required"`
	CountryCode                string `json:"country_code" binding:"required"`
	EmergencyContactCountryCode string `json:"emergency_contact_country_code" binding:"required"`
	EmergencyContactName       string `json:"emergency_contact_name" binding:"required"`
	EmergencyContactRelation   string `json:"emergency_contact_relation" binding:"required"`
	EmergencyContactPhone      string `json:"emergency_contact_phone" binding:"required"`
	Address                    string `json:"address" binding:"required"`
	DomisiliAddress            string `json:"domisili_address" binding:"required"`
	RTRW                       string `json:"rt_rw" binding:"required"`
	SubDistrict                string `json:"sub_district" binding:"required"`
	District                   string `json:"district" binding:"required"`
	Province                   string `json:"province" binding:"required"`
	City                       string `json:"city" binding:"required"`
	PostalCode                 string `json:"postal_code" binding:"required"`
	WorkExperience             string `json:"work_experience"`
	UserTool                   string `json:"user_tool"`
	IsExGolife                 string `json:"is_ex_golife" binding:"required,oneof=0 1"`
	KindOfMitra                string `json:"kind_of_mitra" binding:"required,oneof=Dengan Alat_Tanpa Alat"`
	WorkExperienceDuration     string `json:"work_experience_duration"`
}
