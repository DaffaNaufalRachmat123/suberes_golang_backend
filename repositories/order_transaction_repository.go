package repositories

import (
	"errors"
	"strings"
	"suberes_golang/dtos"
	"suberes_golang/models"
	"time"

	"gorm.io/gorm"
)

type OrderTransactionRepository struct {
	DB *gorm.DB
}

func (r *OrderTransactionRepository) FindRunningOrderByMitraID(mitraID string) (*models.OrderTransaction, error) {
	var orderTransaction models.OrderTransaction
	err := r.DB.Where("mitra_id = ? AND order_status IN (?)", mitraID, []string{"OTW", "ON_PROGRESS"}).First(&orderTransaction).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &orderTransaction, err
}
func (r *OrderTransactionRepository) CountRunningOrders() (int64, error) {
	var count int64

	err := r.DB.Model(&models.OrderTransaction{}).
		Where("order_status IN ?", []string{"OTW", "ON_PROGRESS"}).
		Count(&count).Error

	return count, err
}
func (r *OrderTransactionRepository) CountWaitingOrders() (int64, error) {
	var count int64

	err := r.DB.Model(&models.OrderTransaction{}).
		Where("order_status = ?", "WAITING_FOR_SELECTED_MITRA").
		Count(&count).Error

	return count, err
}
func (r *OrderTransactionRepository) FindById(id string) (*models.OrderTransaction, error) {
	var orderData models.OrderTransaction
	err := r.DB.Where("id = ?", id).First(&orderData)
	return &orderData, err.Error
}

func (r *OrderTransactionRepository) UpdateData(tx *gorm.DB, user *models.OrderTransaction, data map[string]interface{}) error {
	return tx.Model(user).Updates(data).Error
}

func (r *OrderTransactionRepository) GetTodayOrderSummary(mitraID string, start, end time.Time) (*dtos.OrderSummary, error) {
	var result struct {
		OrderCount int64
		Pendapatan int64
	}

	err := r.DB.
		Table("order_transactions").
		Select("COUNT(id) as order_count, COALESCE(SUM(gross_amount_mitra),0) as pendapatan").
		Where("mitra_id = ? AND order_status = ? AND order_time BETWEEN ? AND ?",
			mitraID, "FINISH", start, end).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return &dtos.OrderSummary{
		OrderCount: result.OrderCount,
		Pendapatan: result.Pendapatan,
	}, nil
}

func (r *OrderTransactionRepository) FindDynamicOrderTransactionMap(
	selectFields []string,
	conditions map[string]interface{},
	orQuery string,
	orArgs []interface{},
) (map[string]interface{}, error) {

	result := make(map[string]interface{})

	query := r.DB.Table("order_transactions")

	if len(selectFields) > 0 {
		query = query.Select(selectFields)
	}

	if len(conditions) > 0 {
		query = query.Where(conditions)
	}

	if orQuery != "" {
		query = query.Where(orQuery, orArgs...)
	}

	if err := query.Take(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *OrderTransactionRepository) CreateOrderData(tx *gorm.DB, data models.OrderTransaction) (*models.OrderTransaction, error) {
	if err := tx.Create(&data).Error; err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *OrderTransactionRepository) CountOrders(start time.Time, end time.Time) (int64, error) {
	var count int64
	err := r.DB.Model(&models.OrderTransaction{}).
		Where("created_at >= ? AND created_at <= ?", start, end).
		Count(&count).Error
	return count, err
}

func (r *OrderTransactionRepository) CountFinishedOrders(start time.Time, end time.Time) (int64, error) {
	var count int64
	err := r.DB.Model(&models.OrderTransaction{}).
		Where("created_at >= ? AND created_at <= ? AND order_status = 'FINISH'", start, end).
		Count(&count).Error
	return count, err
}

func (r *OrderTransactionRepository) SumRevenue(start time.Time, end time.Time) (int64, error) {
	var total int64
	err := r.DB.Model(&models.OrderTransaction{}).
		Select("SUM(gross_amount_company)").
		Where("created_at >= ? AND created_at <= ? AND order_status = 'FINISH'", start, end).
		Row().Scan(&total)
	return total, err
}

func (r *OrderTransactionRepository) TotalOrdersByMonth() ([]dtos.TotalOrdersByMonth, error) {
	var result []dtos.TotalOrdersByMonth
	err := r.DB.Model(&models.OrderTransaction{}).
		Select("MONTH(created_at) as month, CASE WHEN MONTH(created_at) = 1 THEN 'Januari' WHEN MONTH(created_at) = 2 THEN 'Februari' WHEN MONTH(created_at) = 3 THEN 'Maret' WHEN MONTH(created_at) = 4 THEN 'April' WHEN MONTH(created_at) = 5 THEN 'Mei' WHEN MONTH(created_at) = 6 THEN 'Juni' WHEN MONTH(created_at) = 7 THEN 'Juli' WHEN MONTH(created_at) = 8 THEN 'Agustus' WHEN MONTH(created_at) = 9 THEN 'September' WHEN MONTH(created_at) = 10 THEN 'Oktober' WHEN MONTH(created_at) = 11 THEN 'November' WHEN MONTH(created_at) = 12 THEN 'Desember' ELSE '' END as bulan, COUNT(id) as order_count").
		Where("order_status = 'FINISH'").
		Group("MONTH(created_at), bulan").
		Scan(&result).Error
	return result, err
}

func (r *OrderTransactionRepository) OverviewMonthRevenue() ([]dtos.OverviewMonthRevenue, error) {
	var result []dtos.OverviewMonthRevenue
	err := r.DB.Model(&models.OrderTransaction{}).
		Select("SUM(gross_amount_company) as total_revenue, MONTH(created_at) as month_number, CASE WHEN MONTH(created_at) = 1 THEN 'Januari' WHEN MONTH(created_at) = 2 THEN 'Februari' WHEN MONTH(created_at) = 3 THEN 'Maret' WHEN MONTH(created_at) = 4 THEN 'April' WHEN MONTH(created_at) = 5 THEN 'Mei' WHEN MONTH(created_at) = 6 THEN 'Juni' WHEN MONTH(created_at) = 7 THEN 'Juli' WHEN MONTH(created_at) = 8 THEN 'Agustus' WHEN MONTH(created_at) = 9 THEN 'September' WHEN MONTH(created_at) = 10 THEN 'Oktober' WHEN MONTH(created_at) = 11 THEN 'November' WHEN MONTH(created_at) = 12 THEN 'Desember' ELSE '' END as bulan").
		Where("order_status = 'FINISH'").
		Group("MONTH(created_at), month_number").
		Scan(&result).Error
	return result, err
}

func (r *OrderTransactionRepository) OverviewWeekRevenue() ([]dtos.OverviewWeekRevenue, error) {
	var result []dtos.OverviewWeekRevenue
	month := time.Now().Month()
	err := r.DB.Raw("SELECT CONCAT(CASE WHEN MONTH(created_at) = 1 THEN 'Januari' WHEN MONTH(created_at) = 2 THEN 'Februari' WHEN MONTH(created_at) = 3 THEN 'Maret' WHEN MONTH(created_at) = 4 THEN 'April' WHEN MONTH(created_at) = 5 THEN 'Mei' WHEN MONTH(created_at) = 6 THEN 'Juni' WHEN MONTH(created_at) = 7 THEN 'Juli' WHEN MONTH(created_at) = 8 THEN 'Agustus' WHEN MONTH(created_at) = 9 THEN 'September' WHEN MONTH(created_at) = 10 THEN 'Oktober' WHEN MONTH(created_at) = 11 THEN 'November' WHEN MONTH(created_at) = 12 THEN 'Desember' ELSE '' END, ' Week ', FLOOR(((DAY(created_at) - 1) / 7) + 1)) as month_week, SUM(gross_amount_company) AS total_transaction FROM order_transactions WHERE order_status = 'FINISH' AND MONTH(created_at) = ? GROUP BY month_week ORDER BY month(created_at), month_week", month).
		Scan(&result).Error
	return result, err
}

func (r *OrderTransactionRepository) FrequentlyUsedService() ([]dtos.FrequentlyUsedService, error) {
	var result []dtos.FrequentlyUsedService
	err := r.DB.Model(&models.Service{}).Order("service_count DESC").Scan(&result).Error
	return result, err
}

func (r *OrderTransactionRepository) MitraOrderToday(start time.Time, end time.Time) ([]dtos.MitraOrderToday, error) {
	var result []dtos.MitraOrderToday
	err := r.DB.Model(&models.OrderTransaction{}).
		Select("COUNT(mitra_id) as order_count, users.id, users.complete_name").
		Joins("JOIN users ON users.id = order_transactions.mitra_id").
		Where("order_transactions.created_at >= ? AND order_transactions.created_at <= ?", start, end).
		Group("mitra_id").
		Order("order_count DESC").
		Scan(&result).Error
	return result, err
}

func (r *OrderTransactionRepository) FindOrderByPaymentID(tx *gorm.DB, paymentID string) (*models.OrderTransaction, error) {
	var order models.OrderTransaction
	err := tx.Where("payment_id_pay = ?", paymentID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderTransactionRepository) UpdateOrderStatus(tx *gorm.DB, orderID string, currentStatus string, updates map[string]interface{}) error {
	return tx.Model(&models.OrderTransaction{}).Where("id = ? AND order_status = ?", orderID, currentStatus).UpdateColumns(updates).Error
}

func (r *OrderTransactionRepository) UpdateWithConditions(
	tx *gorm.DB,
	where map[string]interface{},
	updates map[string]interface{},
) error {

	query := tx.Model(&models.OrderTransaction{})

	// Handle AND conditions
	if andMap, ok := where["AND"].(map[string]interface{}); ok {
		for key, value := range andMap {
			switch v := value.(type) {
			case []string, []int, []interface{}:
				query = query.Where(key+" IN ?", v)
			default:
				query = query.Where(key+" = ?", v)
			}
		}
	}

	// Handle OR conditions
	if orMap, ok := where["OR"].(map[string]interface{}); ok {
		for key, value := range orMap {
			switch v := value.(type) {
			case []string, []int, []interface{}:
				query = query.Or(key+" IN ?", v)
			default:
				query = query.Or(key+" = ?", v)
			}
		}
	}

	return query.Updates(updates).Error
}

func (r *OrderTransactionRepository) FindVoidableOrder(tx *gorm.DB, paymentID string) (*models.OrderTransaction, error) {
	var order models.OrderTransaction
	err := tx.Where("payment_id_pay = ? AND order_status = ?", paymentID, "CANCELED_VOID").First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderTransactionRepository) UpdateVoidStatus(tx *gorm.DB, orderID, status string) error {
	return tx.Model(&models.OrderTransaction{}).Where("id = ?", orderID).Update("void_status", status).Error
}

func (r *OrderTransactionRepository) FindAllByStatusWithPagination(status string, page, limit int, search string) ([]models.OrderTransaction, int64, error) {
	var orders []models.OrderTransaction
	var total int64

	// Exclude coordinate columns — they're only needed in detail endpoints.
	// Crypto fields (private_key_rsa, public_key_rsa, mitra_secret, customer_secret)
	// are already json:"-" on the model, but we still omit them from the SELECT
	// to avoid fetching unnecessary bytes from the DB.
	safeColumns := strings.Join([]string{
		"id", "notification_id", "temp_id", "customer_id", "mitra_id",
		"service_id", "sub_service_id", "order_radius", "customer_name",
		"order_type", "mitra_gender", "order_time", "order_progress_time",
		"order_blast_time", "order_time_temp", "order_origin_soon_time",
		"order_timestamp", "address", "canceled_user", "canceled_reason",
		"order_note", "payment_id", "sub_payment_id", "payment_type",
		"order_status", "void_status", "void_description", "id_transaction",
		"is_rated", "is_rated_customer", "rated", "rated_customer",
		"is_mitra_online", "is_customer_online",
		"is_additional", "order_count_additional", "order_time_additional",
		"gross_amount_additional", "gross_amount", "gross_amount_mitra",
		"gross_amount_company", "gross_amount_company_after_deduction",
		"total_order_repeat", "total_waiting_repeat", "total_done_repeat",
		"is_paid_customer", "timezone_code",
		"payment_id_pay", "order_id_pay", "customer_id_pay", "amount_pay",
		"payment_option", "pending_amount", "status", "is_live",
		"expire_time", "expiry_date", "account_number_va", "mobile_ewallet",
		"qr_string", "qr_code", "va_id", "owner_id", "external_id",
		"bank_code", "merchant_code", "name", "account_number",
		"expected_amount", "is_closed", "expiration_date", "xendit_id",
		"is_single_use", "currency", "xendit_status", "shared_prime",
		"expiration_time", "offer_expired_job_id", "offer_selected_job_id",
		"on_progress_job_id", "ewallet_notify_job_id",
		"created_at", "updated_at",
	}, ", ")

	query := r.DB.Model(&models.OrderTransaction{}).Select(safeColumns)

	if search != "" {
		query = query.Where("id_transaction LIKE ?", "%"+search+"%")
	}

	switch status {
	case "RUNNING":
		query = query.Where("order_status IN ?", []string{"OTW", "ON_PROGRESS"})
	case "REPEAT":
		query = query.Where("order_type = ?", "repeat")
	case "CANCELED":
		query = query.Where("order_status IN ?", []string{"CANCELED", "CANCELED_LATE_PAYMENT", "CANCELED_BY_SYSTEM", "CANCELED_VOID", "CANCELED_VOID_BY_SYSTEM"})
	default:
		query = query.Where("order_status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Scoped preloads: User associations — select only safe, non-sensitive fields.
	// (Sensitive fields are already json:"-" on the User model, but this also
	// avoids fetching large text columns like firebase_token unnecessarily.)
	safeUserSelect := strings.Join([]string{
		"id", "complete_name", "email", "country_code", "phone_number",
		"user_type", "user_level", "color_code_level", "user_rating",
		"user_gender", "user_profile_image", "is_logged_in", "user_status",
		"is_busy", "is_document_completed", "is_mitra_invited",
		"is_mitra_accepted", "is_mitra_rejected", "is_mitra_activated",
		"is_suspended", "suspended_reason", "is_active", "is_auto_bid",
		"kind_of_mitra", "today_order", "today_income", "total_order",
		"account_balance", "total_bills", "age", "date_of_birth",
		"created_at", "updated_at",
	}, ", ")

	offset := (page - 1) * limit
	err := query.
		Preload("Mitra", func(db *gorm.DB) *gorm.DB {
			return db.Select(safeUserSelect)
		}).
		Preload("Customer", func(db *gorm.DB) *gorm.DB {
			return db.Select(safeUserSelect)
		}).
		Preload("Service").
		Preload("SubService").
		Limit(limit).Offset(offset).Order("order_transactions.created_at DESC").
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// ---- New methods for order_transaction endpoints ----

type AdminDashboardStats struct {
	Running              int64 `json:"running"`
	WaitSchedule         int64 `json:"wait_schedule"`
	Canceled             int64 `json:"canceled"`
	Done                 int64 `json:"done"`
	WaitingSelectedMitra int64 `json:"waiting_selected_mitra"`
	Repeat               int64 `json:"repeat"`
}

func (r *OrderTransactionRepository) GetAdminDashboard() (*AdminDashboardStats, error) {
	var stats AdminDashboardStats
	db := r.DB.Model(&models.OrderTransaction{})

	db.Where("order_status IN ?", []string{"OTW", "ON_PROGRESS"}).Count(&stats.Running)
	db.Where("order_status = ?", "WAIT_SCHEDULE").Count(&stats.WaitSchedule)
	db.Where("order_status IN ?", []string{"CANCELED", "CANCELED_LATE_PAYMENT", "CANCELED_BY_SYSTEM", "CANCELED_VOID", "CANCELED_VOID_BY_SYSTEM"}).Count(&stats.Canceled)
	db.Where("order_status = ?", "FINISH").Count(&stats.Done)
	db.Where("order_status = ?", "WAITING_FOR_SELECTED_MITRA").Count(&stats.WaitingSelectedMitra)
	db.Where("order_type = ?", "repeat").Count(&stats.Repeat)

	return &stats, nil
}

func (r *OrderTransactionRepository) FindByIDTransaction(idTransaction string) (*models.OrderTransaction, error) {
	var order models.OrderTransaction
	err := r.DB.Where("id_transaction = ?", idTransaction).
		Preload("Payment").Preload("SubPayment").
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderTransactionRepository) FindForAdminDetail(orderID string) (*models.OrderTransaction, error) {
	var order models.OrderTransaction
	err := r.DB.Where("id = ?", orderID).
		Preload("Customer").
		Preload("Mitra").
		Preload("Service").
		Preload("SubService").
		Preload("Payment").
		Preload("SubPayment").
		Preload("OrderTransactionRepeats").
		Preload("SubServiceAddeds").
		Preload("UserRatings").
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

type SelectedMitraResult struct {
	ID            string  `json:"id"`
	CompleteName  string  `json:"complete_name"`
	UserGender    string  `json:"user_gender"`
	Latitude      string  `json:"latitude"`
	Longitude     string  `json:"longitude"`
	FirebaseToken *string `json:"firebase_token"`
	UserStatus    string  `json:"user_status"`
	IsBusy        string  `json:"is_busy"`
	Distance      float64 `json:"distance"`
}

func (r *OrderTransactionRepository) FindSelectedMitraPaginated(orderID string, customerLat, customerLng float64, limit, offset int) ([]SelectedMitraResult, int64, error) {
	var results []SelectedMitraResult
	var total int64

	baseQuery := r.DB.Table("users u").
		Select(`u.id, u.complete_name, u.user_gender, u.latitude, u.longitude, u.firebase_token, u.user_status, u.is_busy,
			(6371 * acos(cos(radians(?)) * cos(radians(CAST(u.latitude AS DECIMAL(10,8))))
			* cos(radians(CAST(u.longitude AS DECIMAL(10,8))) - radians(?))
			+ sin(radians(?)) * sin(radians(CAST(u.latitude AS DECIMAL(10,8)))))) AS distance`,
			customerLat, customerLng, customerLat).
		Joins("INNER JOIN order_selected_mitras osm ON osm.mitra_id = u.id AND osm.order_id = ?", orderID).
		Where("u.user_type = ?", "mitra")

	r.DB.Table("users u").
		Joins("INNER JOIN order_selected_mitras osm ON osm.mitra_id = u.id AND osm.order_id = ?", orderID).
		Where("u.user_type = ?", "mitra").
		Count(&total)

	err := baseQuery.Order("distance ASC").Limit(limit).Offset(offset).Scan(&results).Error
	return results, total, err
}

func (r *OrderTransactionRepository) FindComingSoonForMitra(mitraID string, limit, offset int) ([]models.OrderTransaction, error) {
	var orders []models.OrderTransaction
	err := r.DB.Where("mitra_id = ? AND order_type = ? AND order_status IN ?", mitraID, "coming soon", []string{"OTW", "WAIT_SCHEDULE"}).
		Preload("Customer").Preload("Service").Preload("SubService").Preload("Payment").Preload("SubPayment").
		Order("order_time ASC").
		Limit(limit).Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *OrderTransactionRepository) FindRunningForCustomer(customerID string, limit, offset int) ([]models.OrderTransaction, error) {
	var orders []models.OrderTransaction
	err := r.DB.Where("customer_id = ? AND order_status IN ?", customerID, []string{"OTW", "ON_PROGRESS"}).
		Preload("Mitra").Preload("Service").Preload("SubService").Preload("Payment").Preload("SubPayment").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *OrderTransactionRepository) FindCanceledForCustomer(customerID string, limit, offset int) ([]models.OrderTransaction, error) {
	var orders []models.OrderTransaction
	err := r.DB.Where("customer_id = ? AND order_status IN ?", customerID,
		[]string{"CANCELED", "CANCELED_LATE_PAYMENT", "CANCELED_BY_SYSTEM", "CANCELED_VOID", "CANCELED_VOID_BY_SYSTEM", "CANCELED_CANT_FIND_MITRA"}).
		Preload("Mitra").Preload("Service").Preload("SubService").Preload("Payment").Preload("SubPayment").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *OrderTransactionRepository) FindDoneForCustomer(customerID string, limit, offset int) ([]models.OrderTransaction, error) {
	var orders []models.OrderTransaction
	err := r.DB.Where("customer_id = ? AND order_status = ?", customerID, "FINISH").
		Preload("Mitra").Preload("Service").Preload("SubService").Preload("Payment").Preload("SubPayment").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *OrderTransactionRepository) FindDoneRangeDateForCustomer(customerID, startDate, endDate string, limit, offset int) ([]models.OrderTransaction, error) {
	var orders []models.OrderTransaction
	err := r.DB.Where("customer_id = ? AND order_status = ? AND DATE(order_time) BETWEEN ? AND ?",
		customerID, "FINISH", startDate, endDate).
		Preload("Mitra").Preload("Service").Preload("SubService").Preload("Payment").Preload("SubPayment").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *OrderTransactionRepository) FindRepeatForCustomer(customerID string, limit, offset int) ([]models.OrderTransaction, error) {
	var orders []models.OrderTransaction
	err := r.DB.Where("customer_id = ? AND order_type = ? AND order_status NOT IN ?",
		customerID, "repeat", []string{"CANCELED", "CANCELED_CANT_FIND_MITRA"}).
		Preload("Mitra").Preload("Service").Preload("SubService").Preload("Payment").Preload("SubPayment").
		Preload("OrderTransactionRepeats").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *OrderTransactionRepository) FindRepeatSearchForCustomer(customerID, completeName string, limit, offset int) ([]models.OrderTransaction, error) {
	var orders []models.OrderTransaction
	err := r.DB.Where("customer_id = ? AND order_type = ? AND order_status NOT IN ?",
		customerID, "repeat", []string{"CANCELED", "CANCELED_CANT_FIND_MITRA"}).
		Joins("JOIN users ON users.id = order_transactions.mitra_id").
		Where("users.complete_name LIKE ?", "%"+completeName+"%").
		Preload("Mitra").Preload("Service").Preload("SubService").Preload("Payment").Preload("SubPayment").
		Preload("OrderTransactionRepeats").
		Order("order_transactions.created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *OrderTransactionRepository) FindRunningOrderDetail(orderID, customerID, mitraID string) (*models.OrderTransaction, error) {
	var order models.OrderTransaction
	query := r.DB.Where("id = ?", orderID)
	if customerID != "0" && customerID != "" {
		query = query.Where("customer_id = ?", customerID)
	}
	if mitraID != "0" && mitraID != "" {
		query = query.Where("mitra_id = ?", mitraID)
	}
	err := query.
		Preload("Customer").Preload("Mitra").
		Preload("Service").Preload("SubService").
		Preload("Payment").Preload("SubPayment").
		Preload("SubServiceAddeds.SubServiceAdditional").
		Preload("SubServiceAddeds").
		First(&order).Error
	return &order, err
}

func (r *OrderTransactionRepository) FindVirtualAccountOrders(customerID string, limit, offset int) ([]models.OrderTransaction, error) {
	var orders []models.OrderTransaction
	err := r.DB.Where("customer_id = ? AND payment_type = ? AND (id_transaction IS NULL OR id_transaction = '') AND order_status NOT IN ?",
		customerID, "virtual account", []string{"CANCELED", "FINISH", "CANCELED_VOID", "CANCELED_VOID_BY_SYSTEM"}).
		Preload("Payment").Preload("SubPayment").Preload("Service").Preload("SubService").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *OrderTransactionRepository) FindOrderDetailFull(orderID, customerID, mitraID string, loadAllRepeat bool) (*models.OrderTransaction, error) {
	var order models.OrderTransaction
	query := r.DB.Where("id = ?", orderID)
	if customerID != "0" && customerID != "" {
		query = query.Where("customer_id = ?", customerID)
	}
	if mitraID != "0" && mitraID != "" {
		query = query.Where("mitra_id = ?", mitraID)
	}
	query = query.
		Preload("Customer").Preload("Mitra").
		Preload("Service").Preload("SubService").
		Preload("Payment").Preload("SubPayment").
		Preload("SubServiceAddeds.SubServiceAdditional").
		Preload("SubServiceAddeds").
		Preload("UserRatings")
	if loadAllRepeat {
		query = query.Preload("OrderTransactionRepeats")
	}
	err := query.First(&order).Error
	return &order, err
}

func (r *OrderTransactionRepository) FindIsAutoBid(orderID, mitraID string) (*models.OrderOffer, error) {
	var offer models.OrderOffer
	err := r.DB.Where("order_id = ? AND mitra_id = ?", orderID, mitraID).First(&offer).Error
	if err != nil {
		return nil, err
	}
	return &offer, nil
}

func (r *OrderTransactionRepository) FindWithDirection(orderID string) (*models.OrderTransaction, error) {
	var order models.OrderTransaction
	err := r.DB.Select("id, direction_response, customer_latitude, customer_longitude, mitra_latitude, mitra_longitude").
		Where("id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderTransactionRepository) UpdateDirectionResponse(tx *gorm.DB, orderID, directionResponse string) error {
	return tx.Model(&models.OrderTransaction{}).
		Where("id = ?", orderID).
		Update("direction_response", directionResponse).Error
}

func (r *OrderTransactionRepository) FindFullForUpdateOnProgress(orderID string) (*models.OrderTransaction, error) {
	var order models.OrderTransaction
	err := r.DB.Where("id = ?", orderID).
		Preload("SubService").Preload("Service").Preload("Mitra").Preload("Customer").
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderTransactionRepository) FindFullForFinish(orderID string) (*models.OrderTransaction, error) {
	var order models.OrderTransaction
	err := r.DB.Where("id = ?", orderID).
		Preload("SubService").Preload("Service").Preload("Mitra").Preload("Customer").Preload("Payment").Preload("SubPayment").
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

type MitraDashboardOrderCounts struct {
	OrderSoonCount     int64
	OrderDoneCount     int64
	OrderRepeatCount   int64
	OrderCanceledCount int64
}

func (r *OrderTransactionRepository) GetMitraDashboardOrderCounts(mitraID string) (*MitraDashboardOrderCounts, error) {
	var result MitraDashboardOrderCounts

	if err := r.DB.Model(&models.OrderTransaction{}).
		Where("mitra_id = ? AND order_type = ? AND order_status = ?", mitraID, "coming soon", "WAIT_SCHEDULE").
		Count(&result.OrderSoonCount).Error; err != nil {
		return nil, err
	}

	if err := r.DB.Model(&models.OrderTransaction{}).
		Where("mitra_id = ? AND order_status = ?", mitraID, "FINISH").
		Count(&result.OrderDoneCount).Error; err != nil {
		return nil, err
	}

	if err := r.DB.Model(&models.OrderTransaction{}).
		Where("mitra_id = ? AND order_type = ?", mitraID, "repeat").
		Count(&result.OrderRepeatCount).Error; err != nil {
		return nil, err
	}

	if err := r.DB.Model(&models.OrderTransaction{}).
		Where("mitra_id = ? AND order_status IN ?", mitraID,
			[]string{"CANCELED", "CANCELED_LATE_PAYMENT", "CANCELED_BY_SYSTEM", "CANCELED_VOID", "CANCELED_VOID_BY_SYSTEM"}).
		Count(&result.OrderCanceledCount).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *OrderTransactionRepository) FindRunningOrderDetailByMitraID(mitraID string) (*models.OrderTransaction, error) {
	var order models.OrderTransaction
	err := r.DB.
		Where("mitra_id = ? AND order_status IN ?", mitraID, []string{"OTW", "ON_PROGRESS"}).
		Preload("Service").
		Preload("SubService").
		First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &order, err
}
