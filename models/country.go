package models

type Country struct {
	BaseModel
	CountryName string

	Regions []Region `gorm:"foreignKey:CountryID"`
}
