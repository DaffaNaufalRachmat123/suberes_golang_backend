package repositories

import (
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
		Where("createdAt >= ? AND createdAt <= ?", start, end).
		Count(&count).Error
	return count, err
}

func (r *OrderTransactionRepository) CountFinishedOrders(start time.Time, end time.Time) (int64, error) {
	var count int64
	err := r.DB.Model(&models.OrderTransaction{}).
		Where("createdAt >= ? AND createdAt <= ? AND order_status = 'FINISH'", start, end).
		Count(&count).Error
	return count, err
}

func (r *OrderTransactionRepository) SumRevenue(start time.Time, end time.Time) (int64, error) {
	var total int64
	err := r.DB.Model(&models.OrderTransaction{}).
		Select("SUM(gross_amount_company)").
		Where("createdAt >= ? AND createdAt <= ? AND order_status = 'FINISH'", start, end).
		Row().Scan(&total)
	return total, err
}

func (r *OrderTransactionRepository) TotalOrdersByMonth() ([]dtos.TotalOrdersByMonth, error) {
	var result []dtos.TotalOrdersByMonth
	err := r.DB.Model(&models.OrderTransaction{}).
		Select("MONTH(createdAt) as month, CASE WHEN MONTH(createdAt) = 1 THEN 'Januari' WHEN MONTH(createdAt) = 2 THEN 'Februari' WHEN MONTH(createdAt) = 3 THEN 'Maret' WHEN MONTH(createdAt) = 4 THEN 'April' WHEN MONTH(createdAt) = 5 THEN 'Mei' WHEN MONTH(createdAt) = 6 THEN 'Juni' WHEN MONTH(createdAt) = 7 THEN 'Juli' WHEN MONTH(createdAt) = 8 THEN 'Agustus' WHEN MONTH(createdAt) = 9 THEN 'September' WHEN MONTH(createdAt) = 10 THEN 'Oktober' WHEN MONTH(createdAt) = 11 THEN 'November' WHEN MONTH(createdAt) = 12 THEN 'Desember' ELSE '' END as bulan, COUNT(id) as order_count").
		Where("order_status = 'FINISH'").
		Group("MONTH(createdAt), bulan").
		Scan(&result).Error
	return result, err
}

func (r *OrderTransactionRepository) OverviewMonthRevenue() ([]dtos.OverviewMonthRevenue, error) {
	var result []dtos.OverviewMonthRevenue
	err := r.DB.Model(&models.OrderTransaction{}).
		Select("SUM(gross_amount_company) as total_revenue, MONTH(createdAt) as month_number, CASE WHEN MONTH(createdAt) = 1 THEN 'Januari' WHEN MONTH(createdAt) = 2 THEN 'Februari' WHEN MONTH(createdAt) = 3 THEN 'Maret' WHEN MONTH(createdAt) = 4 THEN 'April' WHEN MONTH(createdAt) = 5 THEN 'Mei' WHEN MONTH(createdAt) = 6 THEN 'Juni' WHEN MONTH(createdAt) = 7 THEN 'Juli' WHEN MONTH(createdAt) = 8 THEN 'Agustus' WHEN MONTH(createdAt) = 9 THEN 'September' WHEN MONTH(createdAt) = 10 THEN 'Oktober' WHEN MONTH(createdAt) = 11 THEN 'November' WHEN MONTH(createdAt) = 12 THEN 'Desember' ELSE '' END as bulan").
		Where("order_status = 'FINISH'").
		Group("MONTH(createdAt), month_number").
		Scan(&result).Error
	return result, err
}

func (r *OrderTransactionRepository) OverviewWeekRevenue() ([]dtos.OverviewWeekRevenue, error) {
	var result []dtos.OverviewWeekRevenue
	month := time.Now().Month()
	err := r.DB.Raw("SELECT CONCAT(CASE WHEN MONTH(createdAt) = 1 THEN 'Januari' WHEN MONTH(createdAt) = 2 THEN 'Februari' WHEN MONTH(createdAt) = 3 THEN 'Maret' WHEN MONTH(createdAt) = 4 THEN 'April' WHEN MONTH(createdAt) = 5 THEN 'Mei' WHEN MONTH(createdAt) = 6 THEN 'Juni' WHEN MONTH(createdAt) = 7 THEN 'Juli' WHEN MONTH(createdAt) = 8 THEN 'Agustus' WHEN MONTH(createdAt) = 9 THEN 'September' WHEN MONTH(createdAt) = 10 THEN 'Oktober' WHEN MONTH(createdAt) = 11 THEN 'November' WHEN MONTH(createdAt) = 12 THEN 'Desember' ELSE '' END, ' Week ', FLOOR(((DAY(createdAt) - 1) / 7) + 1)) as month_week, SUM(gross_amount_company) AS total_transaction FROM order_transactions WHERE order_status = 'FINISH' AND MONTH(createdAt) = ? GROUP BY month_week ORDER BY month(createdAt), month_week", month).
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
		Where("order_transactions.createdAt >= ? AND order_transactions.createdAt <= ?", start, end).
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

	query := r.DB.Model(&models.OrderTransaction{})

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

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = query.Preload("Mitra").Preload("Customer").Preload("Service").Preload("SubService").Limit(limit).Offset(offset).Order("created_at DESC").Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
