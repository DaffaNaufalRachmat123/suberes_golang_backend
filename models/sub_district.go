package models

type SubDistrict struct {
	BaseModel
	CountryID       uint   `gorm:"type:integer" json:"country_id"`
	RegionID        uint   `gorm:"type:integer" json:"region_id"`
	DistrictID      uint   `gorm:"type:integer" json:"district_id"`
	SubDistrictName string `gorm:"type:varchar(255)" json:"sub_district_name"`

	// Associations
	Country   *Country   `gorm:"foreignKey:CountryID" json:"country,omitempty"`
	Region    *Region    `gorm:"foreignKey:RegionID" json:"region,omitempty"`
	District  *District  `gorm:"foreignKey:DistrictID" json:"district,omitempty"`
	Coverages []Coverage `gorm:"foreignKey:SubDistrictID" json:"coverages,omitempty"`
}
