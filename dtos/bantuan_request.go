package dtos

type BantuanRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	HelpType    string `json:"help_type" binding:"required,oneof=customer mitra"`
}
