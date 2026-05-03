package models

import (
	"time"

	"encoding/json"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderTransaction struct {
	// ...existing fields...

	// Virtual fields for countdown/timeout (tidak disimpan di DB, hanya untuk response JSON)
	CountDownPaymentTimeout          *int64    `gorm:"-" json:"count_down_payment_timeout,omitempty"`
	TimeoutComingSoon                *int64    `gorm:"-" json:"timeout_coming_soon,omitempty"`
	TimeoutLimit                     *int64    `gorm:"-" json:"timeout_limit,omitempty"`
	ID                               string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	NotificationID                   int       `gorm:"type:integer" json:"notification_id"`
	TempID                           string    `gorm:"type:varchar(255)" json:"temp_id"`
	CustomerID                       string    `gorm:"type:varchar(36)" json:"customer_id"`
	MitraID                          *string   `gorm:"type:varchar(36)" json:"mitra_id"`
	ServiceID                        int       `gorm:"type:integer" json:"service_id"`
	SubServiceID                     int       `gorm:"type:integer" json:"sub_service_id"`
	OrderRadius                      int       `gorm:"type:integer" json:"order_radius"`
	CustomerName                     string    `gorm:"type:varchar(255)" json:"customer_name"`
	OrderType                        string    `gorm:"type:varchar(20);check:order_type IN ('now','coming soon','repeat')" json:"order_type"`
	MitraGender                      string    `gorm:"type:varchar(10);check:mitra_gender IN ('male','female')" json:"mitra_gender"`
	OrderTime                        time.Time `gorm:"type:timestamp" json:"order_time"`
	OrderProgressTime                time.Time `gorm:"type:timestamp" json:"order_progress_time"`
	OrderBlastTime                   time.Time `gorm:"type:timestamp" json:"order_blast_time"`
	OrderTimeTemp                    time.Time `gorm:"type:timestamp" json:"order_time_temp"`
	OrderOriginSoonTime              string    `gorm:"type:varchar(255)" json:"order_origin_soon_time"`
	OrderTimestamp                   string    `gorm:"type:varchar(255)" json:"order_timestamp"`
	Address                          string    `gorm:"type:text" json:"address"`
	CanceledUser                     *string   `gorm:"type:varchar(20);check:canceled_user IN ('customer','mitra','admin','superadmin')" json:"canceled_user"`
	CanceledReason                   string    `gorm:"type:text" json:"canceled_reason"`
	OrderNote                        string    `gorm:"type:varchar(255)" json:"order_note"`
	PaymentID                        int       `gorm:"type:integer" json:"payment_id"`
	SubPaymentID                     int       `gorm:"type:integer" json:"sub_payment_id"`
	PaymentType                      string    `gorm:"type:varchar(20);check:payment_type IN ('tunai','balance','ewallet','virtual account')" json:"payment_type"`
	OrderStatus                      string    `gorm:"type:varchar(50);check:order_status IN ('WAITING_PAYMENT','WAIT_SCHEDULE','OTW','ON_PROGRESS','FINISH','CANCELED','CANCELED_CANT_FIND_MITRA','CANCELED_BY_SYSTEM','CANCELED_VOID','CANCELED_VOID_BY_SYSTEM','CANCELED_REFUND','CANCELED_LATE_PAYMENT','CANCELED_FAILED_PAYMENT','FINDING_MITRA','PROCESSING_PAYMENT','WAITING_FOR_SELECTED_MITRA')" json:"order_status"`
	VoidStatus                       *string   `gorm:"type:varchar(20);check:void_status IN ('VOID_PENDING','VOID_SUCCEEDED','VOID_FAILED')" json:"void_status"`
	VoidDescription                  string    `gorm:"type:text" json:"void_description"`
	IDTransaction                    string    `gorm:"type:varchar(255)" json:"id_transaction"`
	IsRated                          string    `gorm:"type:varchar(1);check:is_rated IN ('0','1')" json:"is_rated"`
	IsRatedCustomer                  string    `gorm:"type:varchar(1);check:is_rated_customer IN ('0','1')" json:"is_rated_customer"`
	Rated                            int       `gorm:"type:integer" json:"rated"`
	RatedCustomer                    int       `gorm:"type:integer" json:"rated_customer"`
	IsMitraOnline                    string    `gorm:"type:varchar(1);check:is_mitra_online IN ('0','1')" json:"is_mitra_online"`
	IsCustomerOnline                 string    `gorm:"type:varchar(1);check:is_customer_online IN ('0','1')" json:"is_customer_online"`
	ChatMitraID                      string    `gorm:"type:text" json:"chat_mitra_id"`
	ChatCustomerID                   string    `gorm:"type:text" json:"chat_customer_id"`
	CallMitraID                      string    `gorm:"type:text" json:"call_mitra_id"`
	CallCustomerID                   string    `gorm:"type:text" json:"call_customer_id"`
	CoordinateReceiverID             string    `gorm:"type:text" json:"coordinate_receiver_id"`
	IsAdditional                     string    `gorm:"type:varchar(5);check:is_additional IN ('true','false')" json:"is_additional"`
	OrderCountAdditional             int       `gorm:"type:integer" json:"order_count_additional"`
	OrderTimeAdditional              time.Time `gorm:"type:timestamp" json:"order_time_additional"`
	GrossAmountAdditional            int64     `gorm:"type:bigint" json:"gross_amount_additional"`
	GrossAmount                      int64     `gorm:"type:bigint" json:"gross_amount"`
	GrossAmountMitra                 int64     `gorm:"type:bigint" json:"gross_amount_mitra"`
	GrossAmountCompany               int64     `gorm:"type:bigint" json:"gross_amount_company"`
	GrossAmountCompanyAfterDeduction int64     `gorm:"type:bigint" json:"gross_amount_company_after_deduction"`
	TotalOrderRepeat                 int       `gorm:"type:integer" json:"total_order_repeat"`
	TotalWaitingRepeat               int       `gorm:"type:integer" json:"total_waiting_repeat"`
	TotalDoneRepeat                  int       `gorm:"type:integer" json:"total_done_repeat"`
	IsPaidCustomer                   string    `gorm:"type:varchar(1);check:is_paid_customer IN ('0','1')" json:"is_paid_customer"`
	PrivateKeyRSA                    string    `gorm:"type:text" json:"-"`
	PublicKeyRSA                     string    `gorm:"type:text" json:"-"`
	PrivateKeyMitra                  int64     `gorm:"type:bigint" json:"-"`
	PublicKeyMitra                   int64     `gorm:"type:bigint" json:"public_key_mitra"`
	PublicKeyCustomer                int64     `gorm:"type:bigint" json:"public_key_customer"`
	MitraSecret                      int64     `gorm:"type:bigint" json:"-"`
	CustomerSecret                   int64     `gorm:"type:bigint" json:"-"`
	TimezoneCode                     string    `gorm:"type:varchar(255)" json:"timezone_code"`
	CustomerLatitude                 float64   `gorm:"type:float" json:"customer_latitude"`
	CustomerLongitude                float64   `gorm:"type:float" json:"customer_longitude"`
	MitraLatitude                    float64   `gorm:"type:float" json:"mitra_latitude"`
	MitraLongitude                   float64   `gorm:"type:float" json:"mitra_longitude"`
	PaymentIDPay                     string    `gorm:"type:varchar(255)" json:"payment_id_pay"`
	OrderIDPay                       string    `gorm:"type:varchar(255)" json:"order_id_pay"`
	CustomerIDPay                    string    `gorm:"type:varchar(255)" json:"customer_id_pay"`
	AmountPay                        string    `gorm:"type:varchar(255)" json:"amount_pay"`
	PaymentOption                    string    `gorm:"type:varchar(255)" json:"payment_option"`
	PendingAmount                    string    `gorm:"type:varchar(255)" json:"pending_amount"`
	Status                           string    `gorm:"type:varchar(255)" json:"status"`
	IsLive                           string    `gorm:"type:varchar(5);check:is_live IN ('true','false')" json:"is_live"`
	AccessToken                      string    `gorm:"type:text" json:"access_token"`
	ExpireTime                       string    `gorm:"type:varchar(255)" json:"expire_time"`
	ExpiryDate                       string    `gorm:"type:varchar(255)" json:"expiry_date"`
	AccountNumberVA                  string    `gorm:"type:varchar(255)" json:"account_number_va"`
	MobileEwallet                    string    `gorm:"type:text" json:"mobile_ewallet"`
	CheckoutURLEwallet               string    `gorm:"type:text" json:"checkout_url_ewallet"` // pastikan di DB: text
	WebURLEwallet                    string    `gorm:"type:text" json:"web_url_ewallet"`      // pastikan di DB: text
	QRString                         string    `gorm:"type:text" json:"qr_string"`            // pastikan di DB: text
	QRCode                           string    `gorm:"type:text" json:"qr_code"`              // pastikan di DB: text
	VAID                             string    `gorm:"type:varchar(255)" json:"va_id"`
	OwnerID                          string    `gorm:"type:varchar(255)" json:"owner_id"`
	ExternalID                       string    `gorm:"type:varchar(255)" json:"external_id"`
	BankCode                         string    `gorm:"type:varchar(255)" json:"bank_code"`
	MerchantCode                     string    `gorm:"type:varchar(255)" json:"merchant_code"`
	Name                             string    `gorm:"type:varchar(255)" json:"name"`
	AccountNumber                    string    `gorm:"type:varchar(255)" json:"account_number"`
	ExpectedAmount                   int       `gorm:"type:integer" json:"expected_amount"`
	IsClosed                         string    `gorm:"type:varchar(1);check:is_closed IN ('0','1')" json:"is_closed"`
	ExpirationDate                   string    `gorm:"type:varchar(255)" json:"expiration_date"`
	XenditID                         string    `gorm:"type:varchar(255)" json:"xendit_id"`
	IsSingleUse                      string    `gorm:"type:varchar(1);check:is_single_use IN ('0','1')" json:"is_single_use"`
	Currency                         string    `gorm:"type:varchar(255)" json:"currency"`
	XenditStatus                     string    `gorm:"type:varchar(255)" json:"xendit_status"`
	SharedPrime                      int       `gorm:"type:integer" json:"shared_prime"`
	ExpirationTime                   string    `gorm:"type:varchar(255)" json:"expiration_time"`
	OfferExpiredJobID                string    `gorm:"type:varchar(255)" json:"offer_expired_job_id"`
	OfferSelectedJobID               string    `gorm:"type:varchar(255)" json:"offer_selected_job_id"`
	OnProgressJobID                  string    `gorm:"type:varchar(36)" json:"on_progress_job_id"`
	EwalletNotifyJobID               string    `gorm:"type:varchar(255)" json:"ewallet_notify_job_id"`
	DirectionResponse                string    `gorm:"type:text" json:"direction_response"`
	CreatedAt                        time.Time `gorm:"type:timestamp;default:now()" json:"created_at"`
	UpdatedAt                        time.Time `gorm:"type:timestamp;default:now()" json:"updated_at"`

	// Virtual computed field — populated by select expressions, never inserted/updated/migrated
	CountDownCanTakeOrder *int64 `gorm:"column:count_down_can_take_order;-:migration;-:create;-:update" json:"count_down_can_take_order,omitempty"`

	// Associations
	Customer                *User                    `gorm:"foreignKey:CustomerID;references:ID" json:"customer"`
	Mitra                   *User                    `gorm:"foreignKey:MitraID;references:ID" json:"mitra"`
	Service                 *Service                 `gorm:"foreignKey:ServiceID;references:ID" json:"service"`
	SubService              *SubService              `gorm:"foreignKey:SubServiceID;references:ID" json:"sub_service"`
	Payment                 *Payment                 `gorm:"foreignKey:PaymentID;references:ID" json:"payment"`
	SubPayment              *SubPayment              `gorm:"foreignKey:SubPaymentID;references:ID" json:"sub_payment"`
	OrderRepeats            []OrderRepeat            `gorm:"foreignKey:OrderID" json:"order_repeats"`
	OrderRejects            []OrderRejected          `gorm:"foreignKey:OrderID" json:"order_rejects"`
	UserRatings             []UserRating             `gorm:"foreignKey:OrderID" json:"user_ratings"`
	OrderTransactionRepeats []OrderTransactionRepeat `gorm:"foreignKey:OrderID" json:"order_transaction_repeats"`
	SubServiceAddeds        []SubServiceAdded        `gorm:"foreignKey:OrderID" json:"sub_service_addeds"`
	Transactions            []Transaction            `gorm:"foreignKey:OrderID" json:"transactions"`
	OrderChat               *OrderChat               `gorm:"foreignKey:OrderID" json:"order_chat"`
	OrderOffer              *OrderOffer              `gorm:"foreignKey:OrderID" json:"order_offer"`
	Notifications           []Notification           `gorm:"foreignKey:OrderID" json:"notifications"`
	OrderSelectedMitras     []OrderSelectedMitra     `gorm:"foreignKey:OrderID" json:"order_selected_mitras"`
}

func (OrderTransaction) TableName() string {
	return "order_transactions"
}

func (o *OrderTransaction) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

// Custom MarshalJSON agar SubServiceAddeds selalu [] jika nil
func (o *OrderTransaction) MarshalJSON() ([]byte, error) {
	type Alias OrderTransaction
	// Pastikan SubServiceAddeds tidak nil
	if o.SubServiceAddeds == nil {
		o.SubServiceAddeds = make([]SubServiceAdded, 0)
	}

	return json.Marshal((*Alias)(o))
}
