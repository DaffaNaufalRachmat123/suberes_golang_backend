package models

type Region struct {
	BaseModel
	CountryID  uint   `gorm:"type:integer" json:"country_id"`
	RegionName string `gorm:"type:varchar(255)" json:"region_name"`

	// Associations
	Country   *Country   `gorm:"foreignKey:CountryID" json:"country,omitempty"`
	Coverages []Coverage `gorm:"foreignKey:RegionID" json:"coverages,omitempty"`
}
