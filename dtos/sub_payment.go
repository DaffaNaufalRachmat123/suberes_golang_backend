package dtos

type SubPaymentTutorialRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type SubPaymentUpdateRequest struct {
	Icon               string                     `json:"icon"`
	Title              string                     `json:"title"`
	TitlePayment       string                     `json:"title_payment"`
	Enabled            string                     `json:"enabled" binding:"omitempty,oneof=0 1"`
	Desc               string                     `json:"desc"`
	SubPaymentTutorial *SubPaymentTutorialRequest `json:"sub_payment_tutorial"`
}
