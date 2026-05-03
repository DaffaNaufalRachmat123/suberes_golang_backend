package models

import "time"

type Transaction struct {
	ID                     string    `gorm:"type:varchar(36);primaryKey;column:id" json:"id"`
	MitraID                *string   `gorm:"type:varchar(36);default:null" json:"mitra_id"`
	CustomerID             *string   `gorm:"type:varchar(36);default:null" json:"customer_id"`
	OrderID                *string   `gorm:"type:varchar(36);default:null" json:"order_id"`
	SubOrderID             *int      `gorm:"type:integer;default:null" json:"sub_order_id"`
	RefundID               string    `gorm:"type:varchar(255)" json:"refund_id"`
	UserType               string    `gorm:"type:varchar(10);check:user_type IN ('customer','mitra')" json:"user_type"`
	RefundAmount           int64     `gorm:"type:bigint" json:"refund_amount"`
	RefundType             string    `gorm:"type:varchar(255)" json:"refund_type"`
	ToolID                 *int      `gorm:"type:integer;default:null" json:"tool_id"`
	ToolsCreditsID         *int      `gorm:"type:integer;default:null" json:"tools_credits_id"`
	SubToolsID             *int      `gorm:"type:integer;default:null" json:"sub_tools_id"`
	TopupID                string    `gorm:"type:varchar(10)" json:"topup_id"`
	DisbursementID         string    `gorm:"type:varchar(18)" json:"disbursement_id"`
	OrderIDTransaction     string    `gorm:"type:varchar(10)" json:"order_id_transaction"`
	ExternalID             string    `gorm:"type:varchar(255)" json:"external_id"`
	IdempotencyKey         string    `gorm:"type:varchar(255)" json:"idempotency_key"`
	AccountOwnerName       string    `gorm:"type:varchar(255)" json:"account_owner_name"`
	BankID                 *int      `gorm:"type:integer;default:null" json:"bank_id"`
	BankName               string    `gorm:"type:varchar(255)" json:"bank_name"`
	BankCode               string    `gorm:"type:varchar(255)" json:"bank_code"`
	AccountNumber          string    `gorm:"type:varchar(255)" json:"account_number"`
	TransactionName        string    `gorm:"type:varchar(255)" json:"transaction_name"`
	TransactionAmount      int64     `gorm:"type:bigint" json:"transaction_amount"`
	TransactionFee         int64     `gorm:"type:bigint" json:"transaction_fee"`
	LastAmount             int64     `gorm:"type:bigint" json:"last_amount"`
	MobileEwalletURL       string    `gorm:"type:text" json:"mobile_ewallet_url"`
	TimezoneCode           string    `gorm:"type:varchar(255)" json:"timezone_code"`
	TransactionType        string    `gorm:"type:varchar(20);check:transaction_type IN ('transaction_out','transaction_in')" json:"transaction_type"`
	TransactionTypeFor     string    `gorm:"type:varchar(50)" json:"transaction_type_for"`
	TransactionFor         string    `gorm:"type:varchar(20);check:transaction_for IN ('order','cicilan','disbursement','topup','other')" json:"transaction_for"`
	TransactionStatus      string    `gorm:"type:varchar(20);check:transaction_status IN ('success','waiting','pending','failed')" json:"transaction_status"`
	FailureCode            string    `gorm:"type:varchar(100)" json:"failure_code"`
	TransactionDescription string    `gorm:"type:text" json:"transaction_description"`
	CreatedAt              time.Time `gorm:"type:timestamp;default:now()" json:"created_at"`
	UpdatedAt              time.Time `gorm:"type:timestamp;default:now()" json:"updated_at"`

	// Associations
	MitraTransactionData    *User                   `gorm:"foreignKey:MitraID;references:ID" json:"mitra_transaction_data,omitempty"`
	CustomerTransactionData *User                   `gorm:"foreignKey:CustomerID;references:ID" json:"customer_transaction_data,omitempty"`
	OrderTransaction        *OrderTransaction       `gorm:"foreignKey:OrderID;references:ID" json:"order_transaction,omitempty"`
	OrderTransactionRepeat  *OrderTransactionRepeat `gorm:"foreignKey:SubOrderID;references:ID" json:"order_transaction_repeat,omitempty"`
	Tool                    *Tool                   `gorm:"foreignKey:ToolID;references:ID" json:"tool,omitempty"`
	ToolsCredit             *ToolCredit             `gorm:"foreignKey:ToolsCreditsID;references:ID" json:"tools_credit,omitempty"`
	BankList                *BankList               `gorm:"foreignKey:BankID;references:ID" json:"bank_list,omitempty"`
	SubToolCredit           *SubToolCredit          `gorm:"foreignKey:SubToolsID;references:ID" json:"sub_tool_credit,omitempty"`
	Notifications           []Notification          `gorm:"foreignKey:TransactionID" json:"notifications,omitempty"`
}
