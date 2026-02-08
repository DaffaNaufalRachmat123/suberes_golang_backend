package models

type ScheduleParticipant struct {
	BaseModel

	ScheduleID              int    `gorm:"column:schedule_id" json:"schedule_id"`
	UserID                  string `gorm:"column:user_id" json:"user_id"`
	ParticipantType         string `gorm:"type:enum('executive_level','c_level','employee_level','mitra_level','customer_level','all_level');column:participant_type" json:"participant_type"`
	ParticipantCompleteName string `gorm:"column:participant_complete_name" json:"participant_complete_name"`
	ParticipantEmail        string `gorm:"column:participant_email" json:"participant_email"`
}

func (ScheduleParticipant) TableName() string {
	return "schedule_participants"
}
