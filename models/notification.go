package models

type Notification struct {
	ID string `gorm:"primaryKey;size:36"`

	CustomerID string
	MitraID    string
	AdminID    int

	OrderID      string
	ServiceID    int
	SubServiceID int

	TransactionID string

	UserType            string
	NotificationType    string
	NotificationTitle   string
	NotificationMessage string

	NotifType string
	IsRead    string
}
