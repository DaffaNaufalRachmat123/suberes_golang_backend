package models

type Coverage struct {
	BaseModel
	CountryID     uint `gorm:"type:integer" json:"country_id"`
	RegionID      uint `gorm:"type:integer" json:"region_id"`
	DistrictID    uint `gorm:"type:integer" json:"district_id"`
	SubDistrictID uint `gorm:"type:integer" json:"sub_district_id"`

	// Associations
	Country     *Country     `gorm:"foreignKey:CountryID" json:"country,omitempty"`
	Region      *Region      `gorm:"foreignKey:RegionID" json:"region,omitempty"`
	District    *District    `gorm:"foreignKey:DistrictID" json:"district,omitempty"`
	SubDistrict *SubDistrict `gorm:"foreignKey:SubDistrictID" json:"sub_district,omitempty"`
}
