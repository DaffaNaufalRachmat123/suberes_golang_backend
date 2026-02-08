package models

type SubDistrict struct {
	BaseModel
	CountryID       uint
	RegionID        uint
	DistrictID      uint
	SubDistrictName string
}
