package models

type Schedule struct {
	BaseModel
	CreatorID        string `gorm:"type:varchar(36)" json:"creator_id"`
	CreatorName      string `gorm:"type:varchar(255)" json:"creator_name"`
	ScheduleName     string `gorm:"type:varchar(255)" json:"schedule_name"`
	ScheduleLevel    string `gorm:"type:varchar(50);check:schedule_level IN ('executive_level','c_level','employee_level','mitra_level','customer_level','all_level')" json:"schedule_level"`
	ScheduleTitle    string `gorm:"type:varchar(255)" json:"schedule_title"`
	SchedulePlace    string `gorm:"type:varchar(255)" json:"schedule_place"`
	ScheduleDateTime string `gorm:"type:varchar(255)" json:"schedule_date_time"`
	ScheduleMessage  string `gorm:"type:text" json:"schedule_message"`
	ScheduleTemplate string `gorm:"type:text" json:"schedule_template"`
	ScheduleIsActive string `gorm:"type:varchar(1);check:schedule_is_active IN ('0','1')" json:"schedule_is_active"`
	TimezoneCode     string `gorm:"type:varchar(255)" json:"timezone_code"`

	// Associations
	Creator              *User                 `gorm:"foreignKey:CreatorID;references:ID" json:"creator,omitempty"`
	ScheduleParticipants []ScheduleParticipant `gorm:"foreignKey:ScheduleID" json:"schedule_participants,omitempty"`
}

func (Schedule) TableName() string {
	return "schedules"
}
