package models

type Transaction struct {
	ID           string `gorm:"primaryKey;size:36;column:id" json:"id"`
	MitraID      string `gorm:"column:mitra_id" json:"mitra_id"`
	CustomerID   string `gorm:"column:customer_id" json:"customer_id"`
	OrderID      string `gorm:"column:order_id" json:"order_id"`
	SubOrderID   int    `gorm:"column:sub_order_id" json:"sub_order_id"`
	RefundID     string `gorm:"column:refund_id" json:"refund_id"`
	UserType     string `gorm:"column:user_type" json:"user_type"`
	RefundAmount int    `gorm:"column:refund_amount" json:"refund_amount"`
	RefundType   string `gorm:"column:refund_type" json:"refund_type"`

	ToolID         int `gorm:"column:tool_id" json:"tool_id"`
	ToolsCreditsID int `gorm:"column:tools_credits_id" json:"tools_credits_id"`
	SubToolsID     int `gorm:"column:sub_tools_id" json:"sub_tools_id"`

	TopupID        string `gorm:"column:topup_id" json:"topup_id"`
	DisbursementID string `gorm:"column:disbursement_id" json:"disbursement_id"`

	ExternalID     string `gorm:"column:external_id" json:"external_id"`
	IdempotencyKey string `gorm:"column:idempotency_key" json:"idempotency_key"`

	AccountOwnerName string `gorm:"column:account_owner_name" json:"account_owner_name"`
	BankID           uint   `gorm:"column:bank_id" json:"bank_id"`
	BankName         string `gorm:"column:bank_name" json:"bank_name"`
	BankCode         string `gorm:"column:bank_code" json:"bank_code"`
	AccountNumber    string `gorm:"column:account_number" json:"account_number"`

	TransactionName   string `gorm:"column:transaction_name" json:"transaction_name"`
	TransactionAmount int    `gorm:"column:transaction_amount" json:"transaction_amount"`
	TransactionFee    int    `gorm:"column:transaction_fee" json:"transaction_fee"`
	LastAmount        int    `gorm:"column:last_amount" json:"last_amount"`

	MobileEwalletURL string `gorm:"column:mobile_ewallet_url" json:"mobile_ewallet_url"`
	TimezoneCode     string `gorm:"column:timezone_code" json:"timezone_code"`

	TransactionType    string `gorm:"column:transaction_type" json:"transaction_type"`
	TransactionTypeFor string `gorm:"column:transaction_type_for" json:"transaction_type_for`
	TransactionFor     string `gorm:"column:transaction_for" json:"transaction_for"`
	TransactionStatus  string `gorm:"column:transaction_status" json:"transaction_status"`
	FailureCode        string `gorm:"column:failure_code" json:"failure_code"`

	TransactionDescription string `gorm:"column:transaction_description" json:"transaction_description"`
	CreatedAt              string `gorm:"column:created_at" json:"created_at"`
	UpdatedAt              string `gorm:"column:updated_at" json:"updated_at"`
}
