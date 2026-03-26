package models

import (
	"time"
)

type User struct {
	ID                          string     `gorm:"type:varchar(36);primaryKey;column:id" json:"id"`
	CompleteName                string     `gorm:"type:varchar(255);not null" json:"complete_name"`
	Email                       string     `gorm:"type:varchar(255);not null;unique" json:"email"`
	CountryCode                 string     `gorm:"type:varchar(10);not null" json:"country_code"`
	PhoneNumber                 string     `gorm:"type:varchar(20);not null;unique" json:"phone_number"`
	Password                    string     `gorm:"type:text" json:"-"`
	UserType                    string     `gorm:"type:varchar(10);not null;default:'customer';check:user_type IN ('superadmin','customer','mitra','admin')" json:"user_type"`
	UserLevel                   string     `gorm:"type:varchar(10);not null;default:'no level';check:user_level IN ('no level','silver','gold','platinum')" json:"user_level"`
	ColorCodeLevel              string     `gorm:"type:varchar(10);not null" json:"color_code_level"`
	UserRating                  float64    `gorm:"type:float;default:0" json:"user_rating"`
	UserGender                  string     `gorm:"type:varchar(10);default:'male';check:user_gender IN ('male','female')" json:"user_gender"`
	UserProfileImage            string     `gorm:"type:text" json:"user_profile_image"`
	IsLoggedIn                  string     `gorm:"type:varchar(1);default:'0';check:is_logged_in IN ('0','1')" json:"is_logged_in"`
	UserStatus                  string     `gorm:"type:varchar(20);default:'stay';check:user_status IN ('stay','waiting','picked up','on progress')" json:"user_status"`
	IsBusy                      string     `gorm:"type:varchar(3);default:'no';check:is_busy IN ('yes','no')" json:"is_busy"`
	IsDocumentCompleted         string     `gorm:"type:varchar(1);default:'0';check:is_document_completed IN ('0','1')" json:"is_document_completed"`
	IsMitraInvited              string     `gorm:"type:varchar(1);default:'0';check:is_mitra_invited IN ('0','1')" json:"is_mitra_invited"`
	IsMitraAccepted             string     `gorm:"type:varchar(1);default:'0';check:is_mitra_accepted IN ('0','1')" json:"is_mitra_accepted"`
	IsMitraRejected             string     `gorm:"type:varchar(1);default:'0';check:is_mitra_rejected IN ('0','1')" json:"is_mitra_rejected"`
	IsMitraActivated            string     `gorm:"type:varchar(1);default:'0';check:is_mitra_activated IN ('0','1')" json:"is_mitra_activated"`
	IsSuspended                 string     `gorm:"type:varchar(1);default:'0';check:is_suspended IN ('0','1')" json:"is_suspended"`
	SuspendedReason             string     `gorm:"type:text" json:"suspended_reason"`
	ViolationCount              int        `gorm:"type:integer" json:"violation_count"`
	ViolationDangerCount        int        `gorm:"type:integer" json:"violation_danger_count"`
	IsActive                    string     `gorm:"type:varchar(3);default:'no';check:is_active IN ('yes','no')" json:"is_active"`
	IsAutoBid                   string     `gorm:"type:varchar(3);default:'no';check:is_auto_bid IN ('yes','no')" json:"is_auto_bid"`
	OrderIDRunning              *string    `gorm:"type:integer" json:"order_id_running"`
	SubOrderIDRunning           *int       `gorm:"type:integer" json:"sub_order_id_running"`
	CustomerIDRunning           *int       `gorm:"type:integer" json:"customer_id_running"`
	ServiceIDRunning            *int       `gorm:"type:integer" json:"service_id_running"`
	SubServiceIDRunning         *int       `gorm:"type:integer" json:"sub_service_id_running"`
	Latitude                    string     `gorm:"type:varchar(255);default:''" json:"latitude"`
	Longitude                   string     `gorm:"type:varchar(255);default:''" json:"longitude"`
	FirebaseToken               *string    `gorm:"type:text;default:''" json:"firebase_token"`
	Age                         string     `gorm:"type:varchar(255);default:'0'" json:"age"`
	KTPNumber                   string     `gorm:"type:varchar(16);default:''" json:"ktp_number"`
	DateOfBirth                 string     `gorm:"type:varchar(100);default:''" json:"date_of_birth"`
	PlaceOfBirth                string     `gorm:"type:varchar(100);default:''" json:"place_of_birth"`
	KTPImage                    string     `gorm:"type:text;default:''" json:"ktp_image"`
	KKImage                     string     `gorm:"type:text;default:''" json:"kk_image"`
	Address                     string     `gorm:"type:text;default:''" json:"address"`
	DomisiliAddress             string     `gorm:"type:varchar(255);default:''" json:"domisili_address"`
	RTRW                        string     `gorm:"type:varchar(10);default:''" json:"rt_rw"`
	SubDistrict                 string     `gorm:"type:varchar(100);default:''" json:"sub_district"`
	District                    string     `gorm:"type:varchar(100);default:''" json:"district"`
	Province                    string     `gorm:"type:varchar(100);default:''" json:"province"`
	City                        string     `gorm:"type:varchar(100);default:''" json:"city"`
	PostalCode                  string     `gorm:"type:varchar(10);default:''" json:"postal_code"`
	WorkExperience              string     `gorm:"type:text;default:''" json:"work_experience"`
	UserTool                    string     `gorm:"type:text" json:"user_tool"`
	WorkExperienceCleaning      string     `gorm:"type:varchar(1);default:'0';check:work_experience_cleaning IN ('0','1')" json:"work_experience_cleaning"`
	IsExGolife                  string     `gorm:"type:varchar(1);default:'0';check:is_ex_golife IN ('0','1')" json:"is_ex_golife"`
	KindOfMitra                 string     `gorm:"type:varchar(20);default:'Tanpa Alat';check:kind_of_mitra IN ('Dengan Alat','Tanpa Alat')" json:"kind_of_mitra"`
	WorkExperienceDuration      string     `gorm:"type:varchar(100);default:''" json:"work_experience_duration"`
	EmergencyContactName        string     `gorm:"type:varchar(150);default:''" json:"emergency_contact_name"`
	EmergencyContactRelation    string     `gorm:"type:varchar(100);default:''" json:"emergency_contact_relation"`
	EmergencyContactPhone       string     `gorm:"type:varchar(20);default:''" json:"emergency_contact_phone"`
	EmergencyContactCountryCode string     `gorm:"type:varchar(20);default:''" json:"emergency_contact_country_code"`
	CoverSavingsBook            string     `gorm:"type:text;default:''" json:"cover_savings_book"`
	TimeInvited                 *time.Time `gorm:"type:timestamp" json:"time_invited"`
	PlaceInvited                string     `gorm:"type:varchar(255)" json:"place_invited"`
	NoteInvited                 string     `gorm:"type:text" json:"note_invited"`
	TodayOrder                  int        `gorm:"type:integer;default:0" json:"today_order"`
	TodayIncome                 int64      `gorm:"type:bigint;default:0" json:"today_income"`
	TotalOrder                  int        `gorm:"type:integer;default:0" json:"total_order"`
	AccountBalance              int64      `gorm:"type:bigint;default:0" json:"account_balance"`
	PayPin                      string     `gorm:"type:text" json:"-"`
	DisbursementPin             string     `gorm:"type:text" json:"-"`
	TotalBills                  int        `gorm:"type:integer;default:0" json:"total_bills"`
	SharedPrime                 int64      `gorm:"type:bigint;default:0" json:"-"`
	SharedBase                  int64      `gorm:"type:bigint;default:0" json:"-"`
	SharedSecret                int64      `gorm:"type:bigint;default:0" json:"-"`
	PrivateKeyPayPin            string     `gorm:"type:text" json:"-"`
	PublicKeyPayPin             string     `gorm:"type:text" json:"-"`
	PrivateKeyDisbursementPin   string     `gorm:"type:text" json:"-"`
	PublicKeyDisbursementPin    string     `gorm:"type:text" json:"-"`
	RejectionCount              int        `gorm:"type:integer;default:0" json:"rejection_count"`
	SocketID                    string     `gorm:"type:text;default:''" json:"socket_id"`
	BrowserName                 string     `gorm:"type:varchar(20);default:''" json:"browser_name"`
	IsInCall                    string     `gorm:"type:varchar(1);default:'0';check:is_in_call IN ('0','1')" json:"is_in_call"`
	NonactivateReason           string     `gorm:"type:text;default:''" json:"nonactivate_reason"`
	ActivateReason              string     `gorm:"type:text;default:''" json:"activate_reason"`
	RegisteredFromMobile        string     `gorm:"type:varchar(1);default:'0';check:registered_from_mobile IN ('0','1')" json:"registered_from_mobile"`
	CreatedAt                   time.Time  `gorm:"type:timestamp;default:now()" json:"created_at"`
	UpdatedAt                   time.Time  `gorm:"type:timestamp;default:now()" json:"updated_at"`

	// Associations
	UserOTP                   *UserOTP                 `gorm:"foreignKey:users_id;references:ID" json:"user_otp,omitempty"`
	CustomerOrderTransactions []OrderTransaction       `gorm:"foreignKey:customer_id;references:ID" json:"customer_order_transactions,omitempty"`
	MitraOrderTransactions    []OrderTransaction       `gorm:"foreignKey:mitra_id;references:ID" json:"mitra_order_transactions,omitempty"`
	CustomerRatings           []UserRating             `gorm:"foreignKey:customer_id;references:ID" json:"customer_ratings,omitempty"`
	MitraRatings              []UserRating             `gorm:"foreignKey:mitra_id;references:ID" json:"mitra_ratings,omitempty"`
	CustomerRepeatOrders      []OrderTransactionRepeat `gorm:"foreignKey:customer_id;references:ID" json:"customer_repeat_orders,omitempty"`
	MitraRepeatOrders         []OrderTransactionRepeat `gorm:"foreignKey:mitra_id;references:ID" json:"mitra_repeat_orders,omitempty"`
	SentMessages              []Message                `gorm:"foreignKey:sender_id;references:ID" json:"sent_messages,omitempty"`
	PaymentAccounts           []PaymentAccount         `gorm:"foreignKey:mitra_id;references:ID" json:"payment_accounts,omitempty"`
	BeneficiaryTransactions   []BeneficiaryTransaction `gorm:"foreignKey:user_id;references:ID" json:"beneficiary_transactions,omitempty"`
	UsersTools                []UserTool               `gorm:"foreignKey:user_id;references:ID" json:"users_tools,omitempty"`
	ToolsCredits              []ToolCredit             `gorm:"foreignKey:mitra_id;references:ID" json:"tools_credits,omitempty"`
	SubToolsCredits           []SubToolCredit          `gorm:"foreignKey:mitra_id;references:ID" json:"sub_tools_credits,omitempty"`
	CreatedBanners            []BannerList             `gorm:"foreignKey:creator_id;references:ID" json:"created_banners,omitempty"`
	CreatedLayananServices    []LayananService         `gorm:"foreignKey:creator_id;references:ID" json:"created_layanan_services,omitempty"`
	CreatedNews               []NewsList               `gorm:"foreignKey:creator_id;references:ID" json:"created_news,omitempty"`
	CreatedSyaratKetentuan    []SyaratKetentuan        `gorm:"foreignKey:creator_id;references:ID" json:"created_syarat_ketentuan,omitempty"`
	MitraTransactions         []Transaction            `gorm:"foreignKey:mitra_id;references:ID" json:"mitra_transactions,omitempty"`
	CustomerTransactions      []Transaction            `gorm:"foreignKey:customer_id;references:ID" json:"customer_transactions,omitempty"`
	MitraOrderChats           []OrderChat              `gorm:"foreignKey:mitra_id;references:ID" json:"mitra_order_chats,omitempty"`
	CustomerOrderChats        []OrderChat              `gorm:"foreignKey:customer_id;references:ID" json:"customer_order_chats,omitempty"`
	CustomerOrderOffers       []OrderOffer             `gorm:"foreignKey:customer_id;references:ID" json:"customer_order_offers,omitempty"`
	MitraOrderOffers          []OrderOffer             `gorm:"foreignKey:mitra_id;references:ID" json:"mitra_order_offers,omitempty"`
	CustomerNotifications     []Notification           `gorm:"foreignKey:customer_id;references:ID" json:"customer_notifications,omitempty"`
	MitraNotifications        []Notification           `gorm:"foreignKey:mitra_id;references:ID" json:"mitra_notifications,omitempty"`
	AdminPrivacyPolicies      []PrivacyPolicy          `gorm:"foreignKey:admin_id;references:ID" json:"admin_privacy_policies,omitempty"`
	CustomerComplains         []Complain               `gorm:"foreignKey:customer_id;references:ID" json:"customer_complains,omitempty"`
	MitraSelectedOrders       []OrderSelectedMitra     `gorm:"foreignKey:mitra_id;references:ID" json:"mitra_selected_orders,omitempty"`
}
