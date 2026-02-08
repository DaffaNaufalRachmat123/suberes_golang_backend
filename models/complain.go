package models

type Complain struct {
	BaseModel

	ComplainCode string `gorm:"size:10;column:complain_code" json:"complain_code"`
	CustomerID   string `gorm:"size:36;column:customer_id" json:"customer_id"`
	TitleProblem string `gorm:"type:text;column:title_problem" json:"title_problem"`
	Problem      string `gorm:"column:problem" json:"problem"`
	Status       string `gorm:"type:enum('SENT','ON_REVIEW','SOLVED','CLOSED');column:status" json:"status"`

	// Relations
	ComplainImages []ComplainImage `gorm:"foreignKey:ComplainID;references:ID" json:"complain_images"`
}

func (Complain) TableName() string {
	return "complains"
}
