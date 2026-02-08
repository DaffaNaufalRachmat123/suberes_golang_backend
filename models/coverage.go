package models

type Coverage struct {
	BaseModel
	CountryID     uint
	RegionID      uint
	DistrictID    uint
	SubDistrictID uint
}
