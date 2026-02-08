package models

import "time"

type OrderTransaction struct {
	ID string `gorm:"primaryKey;type:uuid;column:id" json:"id"`

	NotificationID                   int       `gorm:"column:notification_id" json:"notification_id"`
	TempID                           string    `gorm:"column:temp_id" json:"temp_id"`
	CustomerID                       string    `gorm:"size:36;column:customer_id" json:"customer_id"`
	MitraID                          string    `gorm:"size:36;column:mitra_id" json:"mitra_id"`
	ServiceID                        int       `gorm:"column:service_id" json:"service_id"`
	SubServiceID                     int       `gorm:"column:sub_service_id" json:"sub_service_id"`
	OrderRadius                      int       `gorm:"column:order_radius" json:"order_radius"`
	CustomerName                     string    `gorm:"column:customer_name" json:"customer_name"`
	OrderType                        string    `gorm:"type:enum('now','coming soon','repeat');column:order_type" json:"order_type"`
	MitraGender                      string    `gorm:"type:enum('male','female');column:mitra_gender" json:"mitra_gender"`
	OrderTime                        time.Time `gorm:"column:order_time" json:"order_time"`
	OrderProgressTime                time.Time `gorm:"column:order_progress_time" json:"order_progress_time"`
	OrderBlastTime                   time.Time `gorm:"column:order_blast_time" json:"order_blast_time"`
	OrderTimeTemp                    time.Time `gorm:"column:order_time_temp" json:"order_time_temp"`
	OrderOriginSoonTime              string    `gorm:"column:order_origin_soon_time" json:"order_origin_soon_time"`
	OrderTimestamp                   string    `gorm:"column:order_timestamp" json:"order_timestamp"`
	Address                          string    `gorm:"column:address" json:"address"`
	CanceledUser                     string    `gorm:"type:enum('customer','mitra','admin','superadmin');column:canceled_user" json:"canceled_user"`
	CanceledReason                   string    `gorm:"type:text;column:canceled_reason" json:"canceled_reason"`
	OrderNote                        string    `gorm:"column:order_note" json:"order_note"`
	PaymentID                        int       `gorm:"column:payment_id" json:"payment_id"`
	SubPaymentID                     int       `gorm:"column:sub_payment_id" json:"sub_payment_id"`
	PaymentType                      string    `gorm:"type:enum('tunai','balance','ewallet','virtual account');column:payment_type" json:"payment_type"`
	OrderStatus                      string    `gorm:"column:order_status" json:"order_status"` // Enum list is very long, using string for brevity
	VoidStatus                       string    `gorm:"type:enum('VOID_PENDING','VOID_SUCCEEDED','VOID_FAILED');column:void_status" json:"void_status"`
	VoidDescription                  string    `gorm:"type:text;column:void_description" json:"void_description"`
	IDTransaction                    string    `gorm:"column:id_transaction" json:"id_transaction"`
	IsRated                          string    `gorm:"type:enum('0','1');column:is_rated" json:"is_rated"`
	IsRatedCustomer                  string    `gorm:"type:enum('0','1');column:is_rated_customer" json:"is_rated_customer"`
	Rated                            int       `gorm:"column:rated" json:"rated"`
	RatedCustomer                    int       `gorm:"column:rated_customer" json:"rated_customer"`
	IsMitraOnline                    string    `gorm:"type:enum('0','1');column:is_mitra_online" json:"is_mitra_online"`
	IsCustomerOnline                 string    `gorm:"type:enum('0','1');column:is_customer_online" json:"is_customer_online"`
	ChatMitraID                      string    `gorm:"type:text;column:chat_mitra_id" json:"chat_mitra_id"`
	ChatCustomerID                   string    `gorm:"type:text;column:chat_customer_id" json:"chat_customer_id"`
	CallMitraID                      string    `gorm:"type:text;column:call_mitra_id" json:"call_mitra_id"`
	CallCustomerID                   string    `gorm:"type:text;column:call_customer_id" json:"call_customer_id"`
	CoordinateReceiverID             string    `gorm:"type:text;column:coordinate_receiver_id" json:"coordinate_receiver_id"`
	IsAdditional                     string    `gorm:"type:enum('true','false');column:is_additional" json:"is_additional"`
	OrderCountAdditional             int       `gorm:"column:order_count_additional" json:"order_count_additional"`
	OrderTimeAdditional              time.Time `gorm:"column:order_time_additional" json:"order_time_additional"`
	GrossAmountAdditional            int       `gorm:"column:gross_amount_additional" json:"gross_amount_additional"`
	GrossAmount                      int       `gorm:"column:gross_amount" json:"gross_amount"`
	GrossAmountMitra                 int       `gorm:"column:gross_amount_mitra" json:"gross_amount_mitra"`
	GrossAmountCompany               int       `gorm:"column:gross_amount_company" json:"gross_amount_company"`
	GrossAmountCompanyAfterDeduction int       `gorm:"column:gross_amount_company_after_deduction" json:"gross_amount_company_after_deduction"`
	TotalOrderRepeat                 int       `gorm:"column:total_order_repeat" json:"total_order_repeat"`
	TotalWaitingRepeat               int       `gorm:"column:total_waiting_repeat" json:"total_waiting_repeat"`
	TotalDoneRepeat                  int       `gorm:"column:total_done_repeat" json:"total_done_repeat"`
	IsPaidCustomer                   string    `gorm:"type:enum('0','1');column:is_paid_customer" json:"is_paid_customer"`
	PrivateKeyRSA                    string    `gorm:"column:private_key_rsa" json:"private_key_rsa"`
	PublicKeyRSA                     string    `gorm:"column:public_key_rsa" json:"public_key_rsa"`
	PublicKeyMitra                   int64     `gorm:"column:public_key_mitra" json:"public_key_mitra"`
	PublicKeyCustomer                int64     `gorm:"column:public_key_customer" json:"public_key_customer"`
	MitraSecret                      int64     `gorm:"column:mitra_secret" json:"mitra_secret"`
	CustomerSecret                   int64     `gorm:"column:customer_secret" json:"customer_secret"`
	TimezoneCode                     string    `gorm:"column:timezone_code" json:"timezone_code"`
	CustomerLatitude                 float64   `gorm:"column:customer_latitude" json:"customer_latitude"`
	CustomerLongitude                float64   `gorm:"column:customer_longitude" json:"customer_longitude"`
	MitraLatitude                    float64   `gorm:"column:mitra_latitude" json:"mitra_latitude"`
	MitraLongitude                   float64   `gorm:"column:mitra_longitude" json:"mitra_longitude"`

	// Payment Gateway Fields
	PaymentIDPay       string `gorm:"column:payment_id_pay" json:"payment_id_pay"`
	OrderIDPay         string `gorm:"column:order_id_pay" json:"order_id_pay"`
	CustomerIDPay      string `gorm:"column:customer_id_pay" json:"customer_id_pay"`
	AmountPay          string `gorm:"column:amount_pay" json:"amount_pay"`
	PaymentOption      string `gorm:"column:payment_option" json:"payment_option"`
	PendingAmount      string `gorm:"column:pending_amount" json:"pending_amount"`
	Status             string `gorm:"column:status" json:"status"`
	IsLive             string `gorm:"type:enum('true','false');column:is_live" json:"is_live"`
	AccessToken        string `gorm:"type:text;column:access_token" json:"access_token"`
	ExpireTime         string `gorm:"column:expire_time" json:"expire_time"`
	ExpiryDate         string `gorm:"column:expiry_date" json:"expiry_date"`
	AccountNumberVA    string `gorm:"column:account_number_va" json:"account_number_va"`
	MobileEwallet      string `gorm:"column:mobile_ewallet" json:"mobile_ewallet"`
	CheckoutURLEwallet string `gorm:"type:text;column:checkout_url_ewallet" json:"checkout_url_ewallet"`
	WebURLEwallet      string `gorm:"type:text;column:web_url_ewallet" json:"web_url_ewallet"`
	QRString           string `gorm:"type:text;column:qr_string" json:"qr_string"`
	QRCode             string `gorm:"type:text;column:qr_code" json:"qr_code"`
	VAID               string `gorm:"column:va_id" json:"va_id"`
	OwnerID            string `gorm:"column:owner_id" json:"owner_id"`
	ExternalID         string `gorm:"column:external_id" json:"external_id"`
	BankCode           string `gorm:"column:bank_code" json:"bank_code"`
	MerchantCode       string `gorm:"column:merchant_code" json:"merchant_code"`
	Name               string `gorm:"column:name" json:"name"`
	AccountNumber      string `gorm:"column:account_number" json:"account_number"`
	ExpectedAmount     int    `gorm:"column:expected_amount" json:"expected_amount"`
	IsClosed           string `gorm:"type:enum('0','1');column:is_closed" json:"is_closed"`
	ExpirationDate     string `gorm:"column:expiration_date" json:"expiration_date"`
	XenditID           string `gorm:"column:xendit_id" json:"xendit_id"`
	IsSingleUse        string `gorm:"type:enum('0','1');column:is_single_use" json:"is_single_use"`
	Currency           string `gorm:"column:currency" json:"currency"`
	XenditStatus       string `gorm:"column:xendit_status" json:"xendit_status"`
	SharedPrime        int    `gorm:"column:shared_prime" json:"shared_prime"`
	ExpirationTime     string `gorm:"column:expiration_time" json:"expiration_time"`

	// Job IDs
	OfferExpiredJobID  string `gorm:"column:offer_expired_job_id" json:"offer_expired_job_id"`
	OfferSelectedJobID string `gorm:"column:offer_selected_job_id" json:"offer_selected_job_id"`
	OnProgressJobID    string `gorm:"size:36;column:on_progress_job_id" json:"on_progress_job_id"`
	EwalletNotifyJobID string `gorm:"column:ewallet_notify_job_id" json:"ewallet_notify_job_id"`
	DirectionResponse  string `gorm:"type:text;column:direction_response" json:"direction_response"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime:true" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime:true" json:"updated_at"`

	// Relations
	SubServiceAdded []SubServiceAdded `gorm:"foreignKey:OrderID;references:ID" json:"sub_service_added"`
	Transactions    []Transaction     `gorm:"foreignKey:OrderID;references:ID" json:"transactions"`
	Notifications   []Notification    `gorm:"foreignKey:OrderID;references:ID" json:"notifications"`
}

func (OrderTransaction) TableName() string {
	return "order_transactions"
}
