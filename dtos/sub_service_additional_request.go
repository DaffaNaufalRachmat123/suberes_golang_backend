package dtos

type CreateSubServiceAdditionalRequest struct {
	SubServiceID   int     `json:"sub_service_id" binding:"required"`
	Title          string  `json:"title" binding:"required"`
	BaseAmount     float64 `json:"base_amount"`
	Amount         float64 `json:"amount"`
	AdditionalType string  `json:"additional_type" binding:"required"`
}

type UpdateSubServiceAdditionalRequest struct {
	ID             int
	SubServiceID   int     `json:"sub_service_id" binding:"required"`
	Title          string  `json:"title" binding:"required"`
	BaseAmount     float64 `json:"base_amount"`
	Amount         float64 `json:"amount"`
	AdditionalType string  `json:"additional_type" binding:"required"`
}

type DeleteSubServiceAdditionalRequest struct {
	Password string `json:"password" binding:"required"`
}
