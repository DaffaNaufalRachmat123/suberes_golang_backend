package models

import "time"

type SyaratKetentuan struct {
	BaseModel

	CreatorID          int       `gorm:"column:creator_id" json:"creator_id"`
	Title              string    `gorm:"column:title" json:"title"`
	Body               string    `gorm:"type:text;column:body" json:"body"`
	Image              string    `gorm:"type:text;column:image" json:"image"`
	IsPendaftaranMitra string    `gorm:"type:enum('0','1');column:is_pendaftaran_mitra" json:"is_pendaftaran_mitra"`
	IsActive           string    `gorm:"type:enum('0','1');column:is_active" json:"is_active"`
	ExpiredDate        time.Time `gorm:"column:expired_date" json:"expired_date"`
}

func (SyaratKetentuan) TableName() string {
	return "syarat_ketentuan"
}
