package dtos

type BankListCreateItem struct {
	Name             string  `json:"name" binding:"required"`
	Code             string  `json:"code" binding:"required"`
	DisbursementCode string  `json:"disbursement_code"`
	BankImage        string  `json:"bank_image"`
	CanTopup         string  `json:"can_topup"`
	CanDisbursement  string  `json:"can_disbursement"`
	MinTopup         int     `json:"min_topup"`
	MinDisbursement  int     `json:"min_disbursement"`
	TopupFee         float64 `json:"topup_fee"`
	DisbursementFee  float64 `json:"disbursement_fee"`
	IsPercentage     string  `json:"is_percentage"`
}

type BankListUpdateRequest struct {
	Name             string  `json:"name"`
	Code             string  `json:"code"`
	DisbursementCode string  `json:"disbursement_code"`
	BankImage        string  `json:"bank_image"`
	CanTopup         string  `json:"can_topup"`
	CanDisbursement  string  `json:"can_disbursement"`
	MinTopup         int     `json:"min_topup"`
	MinDisbursement  int     `json:"min_disbursement"`
	TopupFee         float64 `json:"topup_fee"`
	DisbursementFee  float64 `json:"disbursement_fee"`
	IsPercentage     string  `json:"is_percentage"`
	MethodType       string  `json:"method_type"`
}
