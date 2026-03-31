package dtos

// TermsConditionCreateRequest adalah representasi JSON input untuk membuat terms condition
type TermsConditionCreateRequest struct {
	Title       string `json:"title" binding:"required"`
	Body        string `json:"body" binding:"required"`
	IsActive    string `json:"is_active" binding:"required,oneof=0 1"`
	TocType     string `json:"toc_type" binding:"required,oneof=terms_of_condition terms_of_service privacy_policy"`
	TocUserType string `json:"toc_user_type" binding:"required,oneof=customer mitra"`
}

// TermsConditionUpdateRequest adalah representasi JSON input untuk update terms condition
type TermsConditionUpdateRequest struct {
	Title       string `json:"title" binding:"required"`
	Body        string `json:"body" binding:"required"`
	IsActive    string `json:"is_active" binding:"required,oneof=0 1"`
	TocType     string `json:"toc_type" binding:"required,oneof=terms_of_condition terms_of_service privacy_policy"`
	TocUserType string `json:"toc_user_type" binding:"required,oneof=customer mitra"`
}

// TermsConditionUpdateStatusRequest adalah representasi JSON input untuk update status TOC
type TermsConditionUpdateStatusRequest struct {
	TocType     string `json:"toc_type" binding:"required,oneof=terms_of_condition terms_of_service privacy_policy"`
	TocUserType string `json:"toc_user_type" binding:"required,oneof=customer mitra"`
	IsActive    string `json:"is_active" binding:"required,oneof=0 1"`
}
