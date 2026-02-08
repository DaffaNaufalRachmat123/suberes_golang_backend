package models

import "time"

type User struct {
	// === Identity ===
	ID           string `gorm:"primaryKey;size:36;column:id" json:"id"`
	CompleteName string `gorm:"column:complete_name" json:"complete_name"`
	Email        string `gorm:"column:email" json:"email"`
	CountryCode  string `gorm:"column:country_code" json:"country_code"`
	PhoneNumber  string `gorm:"column:phone_number" json:"phone_number"`
	Password     string `gorm:"column:password" json:"-"`

	// === Level & Rating ===
	UserType       string  `gorm:"type:enum('customer','mitra','admin');column:user_type" json:"user_type"`
	UserLevel      string  `gorm:"type:enum('no level','silver','gold','platinum');column:user_level;default:'no level'" json:"user_level"` // Tambahkan default
	ColorCodeLevel string  `gorm:"size:10;column:color_code_level" json:"color_code_level"`
	UserRating     float64 `gorm:"column:user_rating;default:0" json:"user_rating"`

	// === Profile Data ===
	UserGender       string `gorm:"type:enum('male','female');column:user_gender;default:'male'" json:"user_gender"`
	UserProfileImage string `gorm:"type:text;column:user_profile_image" json:"user_profile_image"`
	Age              string `gorm:"column:age;default:'0'" json:"age"` // Pakai kutip karena tipe datanya string
	DateOfBirth      string `gorm:"size:100;column:date_of_birth" json:"date_of_birth"`
	PlaceOfBirth     string `gorm:"size:100;column:place_of_birth" json:"place_of_birth"`

	// === Documents & KTP ===
	KTPNumber           string `gorm:"size:16;column:ktp_number" json:"ktp_number"`
	KTPImage            string `gorm:"type:text;column:ktp_image" json:"ktp_image"`
	KKImage             string `gorm:"type:text;column:kk_image" json:"kk_image"`
	IsDocumentCompleted string `gorm:"type:enum('0','1');column:is_document_completed;default:'0'" json:"is_document_completed"`

	// === Status & Activity ===
	IsLoggedIn  string `gorm:"type:enum('0','1');column:is_logged_in;default:'0'" json:"is_logged_in"`
	UserStatus  string `gorm:"type:enum('stay','waiting','picked up','on progress');column:user_status;default:'stay'" json:"user_status"`
	IsBusy      string `gorm:"type:enum('yes','no');column:is_busy;default:'no'" json:"is_busy"`
	IsActive    string `gorm:"type:enum('yes','no');column:is_active;default:'no'" json:"is_active"`
	IsAutoBid   string `gorm:"type:enum('yes','no');column:is_auto_bid;default:'no'" json:"is_auto_bid"`
	SocketID    string `gorm:"type:text;column:socket_id" json:"socket_id"`
	BrowserName string `gorm:"size:20;column:browser_name" json:"browser_name"`
	IsInCall    string `gorm:"type:enum('0','1');column:is_in_call;default:'0'" json:"is_in_call"`

	// === Location ===
	Latitude        string `gorm:"column:latitude" json:"latitude"`
	Longitude       string `gorm:"column:longitude" json:"longitude"`
	Address         string `gorm:"type:text;column:address" json:"address"`
	DomisiliAddress string `gorm:"column:domisili_address" json:"domisili_address"`
	RTRW            string `gorm:"size:10;column:rt_rw" json:"rt_rw"`
	SubDistrict     string `gorm:"size:100;column:sub_district" json:"sub_district"`
	District        string `gorm:"size:100;column:district" json:"district"`
	Province        string `gorm:"size:100;column:province" json:"province"`
	City            string `gorm:"size:100;column:city" json:"city"`
	PostalCode      string `gorm:"size:10;column:postal_code" json:"postal_code"`

	// === Mitra Specific Status ===
	IsMitraInvited   string `gorm:"type:enum('0','1');column:is_mitra_invited;default:'0'" json:"is_mitra_invited"`
	IsMitraAccepted  string `gorm:"type:enum('0','1');column:is_mitra_accepted;default:'0'" json:"is_mitra_accepted"`
	IsMitraRejected  string `gorm:"type:enum('0','1');column:is_mitra_rejected;default:'0'" json:"is_mitra_rejected"`
	IsMitraActivated string `gorm:"type:enum('0','1');column:is_mitra_activated;default:'0'" json:"is_mitra_activated"`
	RejectionCount   int    `gorm:"column:rejection_count;default:0" json:"rejection_count"`

	// === Suspension & Violation ===
	IsSuspended          string `gorm:"type:enum('0','1');column:is_suspended;default:'0'" json:"is_suspended"`
	SuspendedReason      string `gorm:"type:text;column:suspended_reason" json:"suspended_reason"`
	ViolationCount       int    `gorm:"column:violation_count;default:0" json:"violation_count"`
	ViolationDangerCount int    `gorm:"column:violation_danger_count;default:0" json:"violation_danger_count"`
	NonactivateReason    string `gorm:"type:text;column:nonactivate_reason" json:"nonactivate_reason"`
	ActivateReason       string `gorm:"type:text;column:activate_reason" json:"activate_reason"`

	// === Running Order Info ===
	// Pointer *int digunakan agar bisa NULL di database.
	// Jika tidak di-set, nilainya NULL (bukan 0), ini behavior standar SQL.
	OrderIDRunning      *int `gorm:"column:order_id_running" json:"order_id_running"`
	SubOrderIDRunning   *int `gorm:"column:sub_order_id_running" json:"sub_order_id_running"`
	CustomerIDRunning   *int `gorm:"column:customer_id_running" json:"customer_id_running"`
	ServiceIDRunning    *int `gorm:"column:service_id_running" json:"service_id_running"`
	SubServiceIDRunning *int `gorm:"column:sub_service_id_running" json:"sub_service_id_running"`

	// === Experience & Tools ===
	WorkExperience         string `gorm:"type:text;column:work_experience" json:"work_experience"`
	UserTool               string `gorm:"type:text;column:user_tool" json:"user_tool"`
	WorkExperienceCleaning string `gorm:"type:enum('0','1');column:work_experience_cleaning;default:'0'" json:"work_experience_cleaning"`
	IsExGolife             string `gorm:"type:enum('0','1');column:is_ex_golife;default:'0'" json:"is_ex_golife"`
	KindOfMitra            string `gorm:"type:enum('Dengan Alat','Tanpa Alat');column:kind_of_mitra;default:'Tanpa Alat'" json:"kind_of_mitra"`
	WorkExperienceDuration string `gorm:"size:100;column:work_experience_duration" json:"work_experience_duration"`

	// === Emergency Contact ===
	EmergencyContactName        string `gorm:"size:150;column:emergency_contact_name" json:"emergency_contact_name"`
	EmergencyContactRelation    string `gorm:"size:100;column:emergency_contact_relation" json:"emergency_contact_relation"`
	EmergencyContactPhone       string `gorm:"size:20;column:emergency_contact_phone" json:"emergency_contact_phone"`
	EmergencyContactCountryCode string `gorm:"size:20;column:emergency_contact_country_code" json:"emergency_contact_country_code"`

	// === Financial & Bank ===
	AccountBalance   int64  `gorm:"column:account_balance;default:0" json:"account_balance"`
	CoverSavingsBook string `gorm:"type:text;column:cover_savings_book" json:"cover_savings_book"`
	TodayOrder       int    `gorm:"column:today_order;default:0" json:"today_order"`
	TodayIncome      int64  `gorm:"column:today_income;default:0" json:"today_income"`
	TotalOrder       int    `gorm:"column:total_order;default:0" json:"total_order"`
	TotalBills       int    `gorm:"column:total_bills;default:0" json:"total_bills"`

	// === Security & PIN ===
	PayPin                    string `gorm:"type:text;column:pay_pin" json:"-"`
	DisbursementPin           string `gorm:"type:text;column:disbursement_pin" json:"-"`
	SharedPrime               int    `gorm:"column:shared_prime;default:0" json:"-"`
	SharedBase                int    `gorm:"column:shared_base;default:0" json:"-"`
	SharedSecret              int    `gorm:"column:shared_secret;default:0" json:"-"`
	PrivateKeyPayPin          string `gorm:"type:text;column:private_key_pay_pin" json:"-"`
	PublicKeyPayPin           string `gorm:"type:text;column:public_key_pay_pin" json:"-"`
	PrivateKeyDisbursementPin string `gorm:"type:text;column:private_key_disbursement_pin" json:"-"`
	PublicKeyDisbursementPin  string `gorm:"type:text;column:public_key_disbursement_pin" json:"-"`

	// === Invitation Info ===
	TimeInvited  *time.Time `gorm:"column:time_invited" json:"time_invited"`
	PlaceInvited string     `gorm:"size:255;column:place_invited" json:"place_invited"`
	NoteInvited  string     `gorm:"type:text;column:note_invited" json:"note_invited"`

	// === System Info ===
	FirebaseToken        string `gorm:"type:text;column:firebase_token" json:"firebase_token"`
	RegisteredFromMobile string `gorm:"type:enum('0','1');column:registered_from_mobile;default:'0'" json:"registered_from_mobile"`

	// Agar CreatedAt dan UpdatedAt dihandle MySQL (bukan Go), gunakan default:CURRENT_TIMESTAMP
	CreatedAt time.Time `gorm:"column:createdAt;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updatedAt;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`

	// === RELATIONS (Tidak berubah) ===
	UserOTP      *UserOTP      `gorm:"foreignKey:UsersID;references:ID" json:"user_otp"`
	Transactions []Transaction `gorm:"foreignKey:CustomerID;references:ID" json:"transactions"`

	// ... sisa relasi lainnya ...
	CustomerOrders []OrderTransaction `gorm:"foreignKey:CustomerID;references:ID" json:"customer_orders"`
	// dst...
}
