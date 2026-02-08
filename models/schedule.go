package models

type Schedule struct {
	BaseModel

	CreatorID        int    `gorm:"column:creator_id" json:"creator_id"`
	CreatorName      string `gorm:"column:creator_name" json:"creator_name"`
	ScheduleName     string `gorm:"column:schedule_name" json:"schedule_name"`
	ScheduleLevel    string `gorm:"type:enum('executive_level','c_level','employee_level','mitra_level','customer_level','all_level');column:schedule_level" json:"schedule_level"`
	ScheduleTitle    string `gorm:"column:schedule_title" json:"schedule_title"`
	SchedulePlace    string `gorm:"column:schedule_place" json:"schedule_place"`
	ScheduleDateTime string `gorm:"column:schedule_date_time" json:"schedule_date_time"` // Menggunakan string sesuai JS (bukan DATE)
	ScheduleMessage  string `gorm:"type:text;column:schedule_message" json:"schedule_message"`
	ScheduleTemplate string `gorm:"type:text;column:schedule_template" json:"schedule_template"`
	ScheduleIsActive string `gorm:"type:enum('0','1');column:schedule_is_active" json:"schedule_is_active"`
	TimezoneCode     string `gorm:"column:timezone_code" json:"timezone_code"`

	// Relations
	ScheduleParticipants []ScheduleParticipant `gorm:"foreignKey:ScheduleID;references:ID" json:"schedule_participants"`
}

func (Schedule) TableName() string {
	return "schedules"
}
