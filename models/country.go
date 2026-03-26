package models

type Country struct {
	BaseModel
	CountryName string `gorm:"type:varchar(255)" json:"country_name"`

	// Associations
	Coverages []Coverage `gorm:"foreignKey:CountryID" json:"coverages,omitempty"`
	Region    *Region    `gorm:"foreignKey:CountryID" json:"region,omitempty"`
}
