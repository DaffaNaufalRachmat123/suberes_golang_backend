package models

import "time"

type SyaratKetentuan struct {
	BaseModel
	CreatorID          string    `gorm:"type:varchar(36)" json:"creator_id"`
	Title              string    `gorm:"type:varchar(255)" json:"title"`
	Body               string    `gorm:"type:text" json:"body"`
	Image              string    `gorm:"type:text" json:"image"`
	IsPendaftaranMitra string    `gorm:"type:varchar(1);check:is_pendaftaran_mitra IN ('0','1')" json:"is_pendaftaran_mitra"`
	IsActive           string    `gorm:"type:varchar(1);check:is_active IN ('0','1')" json:"is_active"`
	ExpiredDate        time.Time `gorm:"type:timestamp" json:"expired_date"`

	// Associations
	Creator *User `gorm:"foreignKey:CreatorID;references:ID" json:"creator,omitempty"`
}

func (SyaratKetentuan) TableName() string {
	return "syarat_ketentuans"
}
