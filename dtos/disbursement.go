package dtos

type TopupRequest struct {
	BankID   int    `json:"bank_id" binding:"required"`
	BankName string `json:"bank_name" binding:"required"`
	BankCode string `json:"bank_code" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Amount   int64  `json:"amount" binding:"required"`
	TopupFee int64  `json:"topup_fee"`
}

type DisburseRequest struct {
	Amount            int64  `json:"amount" binding:"required"`
	BankID            int    `json:"bank_id" binding:"required"`
	BankCode          string `json:"bank_code" binding:"required"`
	BankName          string `json:"bank_name" binding:"required"`
	AccountHolderName string `json:"account_holder_name" binding:"required"`
	AccountNumber     string `json:"account_number" binding:"required"`
	Description       string `json:"description" binding:"required"`
	Password          string `json:"password" binding:"required"`
}

type DisburseCustomerRequest struct {
	Amount            int64  `json:"amount" binding:"required"`
	BankID            int    `json:"bank_id" binding:"required"`
	AccountHolderName string `json:"account_holder_name" binding:"required"`
	AccountNumber     string `json:"account_number" binding:"required"`
	Description       string `json:"description" binding:"required"`
	Pin               string `json:"pin"`
}

type TopupCallbackPayload struct {
	BankCode      string `json:"bank_code"`
	ExternalID    string `json:"external_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
	AccountNumber string `json:"account_number"`
	Currency      string `json:"currency"`
}

type DisbursementCallbackPayload struct {
	ID          string `json:"id"`
	ExternalID  string `json:"external_id"`
	Status      string `json:"status"`
	FailureCode string `json:"failure_code"`
	Amount      int64  `json:"amount"`
}

type ValidateBankRequest struct {
	AccountNumber string `json:"account_number" binding:"required"`
	BankCode      string `json:"bank_code" binding:"required"`
}

// XenditEwalletChargeAPIResponse is the direct Xendit ewallet charge response structure.
type XenditEwalletChargeAPIResponse struct {
	ID      string `json:"id"`
	Actions struct {
		MobileWebCheckoutURL string `json:"mobile_web_checkout_url"`
	} `json:"actions"`
}

// XenditDisbursementAPIResponse is the Xendit disbursement creation response.
type XenditDisbursementAPIResponse struct {
	ID         string `json:"id"`
	ExternalID string `json:"external_id"`
}

// XenditVAAPIResponse is the Xendit virtual account creation response.
type XenditVAAPIResponse struct {
	ID             string `json:"id"`
	ExternalID     string `json:"external_id"`
	AccountNumber  string `json:"account_number"`
	BankCode       string `json:"bank_code"`
	MerchantCode   string `json:"merchant_code"`
	Name           string `json:"name"`
	ExpectedAmount int64  `json:"expected_amount"`
	Status         string `json:"status"`
}
