package dtos

type ChangeEmailDTO struct {
	Email string `json:"email" binding:"required,email"`
}
