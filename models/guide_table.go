package models

type GuideTable struct {
	BaseModel

	GuideTitle       string `gorm:"column:guide_title" json:"guide_title"`
	GuideDescription string `gorm:"type:text;column:guide_description" json:"guide_description"`
	GuideType        string `gorm:"type:enum('customer','mitra');column:guide_type" json:"guide_type"`
	WatchingCount    int    `gorm:"column:watching_count" json:"watching_count"`
}

func (GuideTable) TableName() string {
	return "guide_tables"
}
