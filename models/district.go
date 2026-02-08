package models

type District struct {
	BaseModel
	CountryID    uint
	RegionID     uint
	DistrictName string

	SubDistricts []SubDistrict `gorm:"foreignKey:DistrictID"`
}
