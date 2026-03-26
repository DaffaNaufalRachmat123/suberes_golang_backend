package models

type Notification struct {
	ID                  string `gorm:"type:varchar(36);primaryKey" json:"id"`
	CustomerID          string `gorm:"type:varchar(36)" json:"customer_id"`
	AdminID             int    `gorm:"type:integer" json:"admin_id"`
	MitraID             string `gorm:"type:varchar(36)" json:"mitra_id"`
	OrderID             string `gorm:"type:varchar(36)" json:"order_id"`
	SubOrderID          int    `gorm:"type:integer" json:"sub_order_id"`
	ServiceID           int    `gorm:"type:integer" json:"service_id"`
	SubServiceID        int    `gorm:"type:integer" json:"sub_service_id"`
	TransactionID       string `gorm:"type:varchar(36)" json:"transaction_id"`
	UserType            string `gorm:"type:varchar(10);check:user_type IN ('customer','mitra','admin')" json:"user_type"`
	NotificationType    string `gorm:"type:varchar(255)" json:"notification_type"`
	NotificationTitle   string `gorm:"type:varchar(255)" json:"notification_title"`
	NotificationMessage string `gorm:"type:varchar(255)" json:"notification_message"`
	NotifType           string `gorm:"type:varchar(20);check:notif_type IN ('promo','order','topup','disbursement','status','news','other')" json:"notif_type"`
	IsRead              string `gorm:"type:varchar(1);check:is_read IN ('0','1')" json:"is_read"`

	// Associations
	CustomerData            *User                   `gorm:"foreignKey:CustomerID;references:ID" json:"customer_data,omitempty"`
	MitraData               *User                   `gorm:"foreignKey:MitraID;references:ID" json:"mitra_data,omitempty"`
	OrderTransaction        *OrderTransaction       `gorm:"foreignKey:OrderID;references:ID" json:"order_transaction,omitempty"`
	OrderTransactionRepeat  *OrderTransactionRepeat `gorm:"foreignKey:SubOrderID;references:ID" json:"order_transaction_repeat,omitempty"`
	Service                 *Service                `gorm:"foreignKey:ServiceID;references:ID" json:"service,omitempty"`
	SubService              *SubService             `gorm:"foreignKey:SubServiceID;references:ID" json:"sub_service,omitempty"`
	Transaction             *Transaction            `gorm:"foreignKey:TransactionID;references:ID" json:"transaction,omitempty"`
}
