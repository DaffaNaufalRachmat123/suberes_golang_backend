package models

type District struct {
	BaseModel
	CountryID    uint   `gorm:"type:integer" json:"country_id"`
	RegionID     uint   `gorm:"type:integer" json:"region_id"`
	DistrictName string `gorm:"type:varchar(255)" json:"district_name"`

	// Associations
	Country   *Country   `gorm:"foreignKey:CountryID" json:"country,omitempty"`
	Region    *Region    `gorm:"foreignKey:RegionID" json:"region,omitempty"`
	Coverages []Coverage `gorm:"foreignKey:DistrictID" json:"coverages,omitempty"`
}
