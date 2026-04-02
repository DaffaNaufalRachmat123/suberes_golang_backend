package dtos

type SubServiceAdditionalRequest struct {
	Title          string  `json:"title" binding:"required"`
	Amount         float64 `json:"amount" binding:"required"`
	AdditionalType string  `json:"additional_type" binding:"required"`
}

type SubServiceCreateRequest struct {
	ServiceID             int                           `json:"service_id" binding:"required"`
	SubPriceServiceTitle  string                        `json:"sub_price_service_title" binding:"required"`
	SubPriceService       int64                         `json:"sub_price_service" binding:"required"`
	SubServiceDescription string                        `json:"sub_service_description" binding:"required"`
	CompanyPercentage     float64                       `json:"company_percentage" binding:"required"`
	MinutesSubServices    int                           `json:"minutes_sub_services"`
	Criteria              string                        `json:"criteria"`
	IsRecommended         string                        `json:"is_recommended" binding:"required,oneof=0 1"`
	SubServiceAdditionals []SubServiceAdditionalRequest `json:"sub_service_additionals" binding:"required,dive"`
}

type SubServiceUpdateRequest struct {
	ID                    int     `json:"id" binding:"required"`
	ServiceID             int     `json:"service_id" binding:"required"`
	SubPriceServiceTitle  string  `json:"sub_price_service_title" binding:"required"`
	SubPriceService       int64   `json:"sub_price_service" binding:"required"`
	SubServiceDescription string  `json:"sub_service_description" binding:"required"`
	CompanyPercentage     float64 `json:"company_percentage" binding:"required"`
	MinutesSubServices    int     `json:"minutes_sub_services"`
	Criteria              string  `json:"criteria"`
	IsRecommended         string  `json:"is_recommended" binding:"required,oneof=0 1"`
}

type SubServiceDeleteRequest struct {
	Password string `json:"password" binding:"required"`
}
