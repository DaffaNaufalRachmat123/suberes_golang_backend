package models

type ScheduleParticipant struct {
	BaseModel
	ScheduleID              int    `gorm:"type:integer" json:"schedule_id"`
	UserID                  string `gorm:"type:varchar(36)" json:"user_id"`
	ParticipantType         string `gorm:"type:varchar(50);check:participant_type IN ('executive_level','c_level','employee_level','mitra_level','customer_level','all_level')" json:"participant_type"`
	ParticipantCompleteName string `gorm:"type:varchar(255)" json:"participant_complete_name"`
	ParticipantEmail        string `gorm:"type:varchar(255)" json:"participant_email"`

	// Associations
	Schedule *Schedule `gorm:"foreignKey:ScheduleID;references:ID" json:"schedule,omitempty"`
	User     *User     `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

func (ScheduleParticipant) TableName() string {
	return "schedule_participants"
}
