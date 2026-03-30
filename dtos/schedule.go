package dtos

type CreateScheduleRequest struct {
	CreatorID        string `json:"creator_id" binding:"required"`
	CreatorName      string `json:"creator_name" binding:"required"`
	ScheduleName     string `json:"schedule_name" binding:"required"`
	ScheduleLevel    string `json:"schedule_level" binding:"required,oneof=executive_level c_level employee_level mitra_level customer_level all_level"`
	ScheduleTitle    string `json:"schedule_title" binding:"required"`
	SchedulePlace    string `json:"schedule_place" binding:"required"`
	ScheduleDateTime string `json:"schedule_date_time" binding:"required"`
	ScheduleMessage  string `json:"schedule_message" binding:"required"`
	ScheduleIsActive string `json:"schedule_is_active" binding:"required,oneof=0 1"`
	TimezoneCode     string `json:"timezone_code" binding:"required"`
}

type UpdateScheduleRequest struct {
	CreatorID        string `json:"creator_id" binding:"required"`
	CreatorName      string `json:"creator_name" binding:"required"`
	ScheduleName     string `json:"schedule_name" binding:"required"`
	ScheduleLevel    string `json:"schedule_level" binding:"required,oneof=executive_level c_level employee_level mitra_level customer_level all_level"`
	ScheduleTitle    string `json:"schedule_title" binding:"required"`
	SchedulePlace    string `json:"schedule_place" binding:"required"`
	ScheduleDateTime string `json:"schedule_date_time" binding:"required"`
	ScheduleMessage  string `json:"schedule_message" binding:"required"`
	ScheduleIsActive string `json:"schedule_is_active" binding:"required,oneof=0 1"`
	TimezoneCode     string `json:"timezone_code" binding:"required"`
}
