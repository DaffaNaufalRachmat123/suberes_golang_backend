package models

type Complain struct {
	BaseModel
	ComplainCode string `gorm:"type:varchar(10)" json:"complain_code"`
	CustomerID   string `gorm:"type:varchar(36)" json:"customer_id"`
	TitleProblem string `gorm:"type:text" json:"title_problem"`
	Problem      string `gorm:"type:varchar(255)" json:"problem"`
	Status       string `gorm:"type:varchar(20);check:status IN ('SENT','ON_REVIEW','SOLVED','CLOSED')" json:"status"`

	// Associations
	Customer       *User           `gorm:"foreignKey:CustomerID;references:ID" json:"customer,omitempty"`
	ComplainImages []ComplainImage `gorm:"foreignKey:ComplainID;references:ID" json:"complain_images,omitempty"`
}

func (Complain) TableName() string {
	return "complains"
}
