package models

type Region struct {
	BaseModel
	CountryID  uint
	RegionName string

	Districts []District `gorm:"foreignKey:RegionID"`
}
