package models

type Service struct {
	BaseModel
	ParentID              int    `gorm:"type:integer" json:"parent_id"`
	ServiceName           string `gorm:"type:varchar(255)" json:"service_name"`
	ServiceDescription    string `gorm:"type:text" json:"service_description"`
	ServiceImageThumbnail string `gorm:"type:text" json:"service_image_thumbnail"`
	ServiceCount          int    `gorm:"type:integer" json:"service_count"`
	ServiceType           string `gorm:"type:varchar(20);check:service_type IN ('Durasi','Luas Ruangan','Single')" json:"service_type"`
	ServiceCategory       string `gorm:"type:varchar(20);check:service_category IN ('Cleaning','Disinfectant','Fogging','Borongan','Lainnya')" json:"service_category"`
	IsResidental          string `gorm:"type:varchar(5);check:is_residental IN ('true','false')" json:"is_residental"`
	ServiceStatus         string `gorm:"type:varchar(20);check:service_status IN ('Regular','Premium','Pro Premium')" json:"service_status"`
	IsActive              string `gorm:"type:varchar(1);check:is_active IN ('0','1')" json:"is_active"`
	MaxOrderCount         int    `gorm:"type:integer" json:"max_order_count"`
	PaymentTimeout        int    `gorm:"type:integer" json:"payment_timeout"`

	// Associations
	CategoryService         *CategoryService         `gorm:"foreignKey:ParentID;references:ID" json:"category_service,omitempty"`
	SubServices             []SubService             `gorm:"foreignKey:ServiceID" json:"sub_services,omitempty"`
	OrderTransactions       []OrderTransaction       `gorm:"foreignKey:ServiceID" json:"order_transactions,omitempty"`
	ServicePromos           []ServicePromo           `gorm:"foreignKey:ServiceID" json:"service_promos,omitempty"`
	OrderTransactionRepeats []OrderTransactionRepeat `gorm:"foreignKey:ServiceID" json:"order_transaction_repeats,omitempty"`
	UserRatings             []UserRating             `gorm:"foreignKey:ServiceID" json:"user_ratings,omitempty"`
	OrderChats              []OrderChat              `gorm:"foreignKey:ServiceID" json:"order_chats,omitempty"`
	ServiceGuarantee        *ServiceGuarantee        `gorm:"foreignKey:ServiceID" json:"service_guarantee,omitempty"`
	Notifications           []Notification           `gorm:"foreignKey:ServiceID" json:"notifications,omitempty"`
}
