package dtos

// PanduanCreateRequest adalah representasi JSON input untuk membuat panduan
type PanduanCreateRequest struct {
	GuideTitle       string `json:"guide_title" binding:"required"`
	GuideDescription string `json:"guide_description" binding:"required"`
	GuideType        string `json:"guide_type" binding:"required,oneof=customer mitra"`
}

// PanduanUpdateRequest adalah representasi JSON input untuk update panduan
type PanduanUpdateRequest struct {
	GuideTitle       string `json:"guide_title" binding:"required"`
	GuideDescription string `json:"guide_description" binding:"required"`
	GuideType        string `json:"guide_type" binding:"required,oneof=customer mitra"`
}
