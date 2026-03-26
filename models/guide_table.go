package models

type GuideTable struct {
	BaseModel
	GuideTitle       string `gorm:"type:varchar(255)" json:"guide_title"`
	GuideDescription string `gorm:"type:text" json:"guide_description"`
	GuideType        string `gorm:"type:varchar(10);check:guide_type IN ('customer','mitra')" json:"guide_type"`
	WatchingCount    int    `gorm:"type:integer;default:0" json:"watching_count"`
}

func (GuideTable) TableName() string {
	return "guide_tables"
}
