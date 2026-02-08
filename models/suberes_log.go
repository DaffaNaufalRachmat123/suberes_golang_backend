package models

type SuberesLogs struct {
	ID      uint   `gorm:"primaryKey;column:id" json:"id"`
	LogName string `gorm:"column:log_name" json:"log_name"`
	LogType string `gorm:"column:log_type" json:"log_type"`
	LogURL  string `gorm:"column:log_url" json:"log_url"`
	LogBody string `gorm:"type:text;column:log_body" json:"log_body"`
	LogTime string `gorm:"column:log_time" json:"log_time"`
}

func (SuberesLogs) TableName() string {
	return "suberes_logs"
}
