package repositories

import (
	"suberes_golang/models"
)

// ── Result Types ──────────────────────────────────────────────────────────────

type OrderHistoryDateResult struct {
	OrderTimeCustomer string `json:"order_time_customer"`
	OrderTimestamp    string `json:"order_timestamp"`
}

// ── Common status slices ──────────────────────────────────────────────────────

var histCanceledStatuses = []string{
	"CANCELED", "CANCELED_VOID", "CANCELED_VOID_BY_SYSTEM", "CANCELED_BY_SYSTEM",
	"CANCELED_CANT_FIND_MITRA", "CANCELED_LATE_PAYMENT", "CANCELED_FAILED_PAYMENT",
}

var histRunningStatuses = []string{
	"FINDING_MITRA", "OTW", "ON_PROGRESS", "WAITING_FOR_SELECTED_MITRA",
}

// ── Order Canceleds ───────────────────────────────────────────────────────────

func (r *OrderTransactionRepository) FindCanceledDatesByMitra(mitraID string) ([]OrderHistoryDateResult, error) {
	var results []OrderHistoryDateResult
	err := r.DB.Table("order_transactions").
		Select("TO_CHAR(order_time, 'YYYY-MM-DD') AS order_time_customer, MIN(order_timestamp) AS order_timestamp").
		Where("mitra_id = ? AND order_status IN ?", mitraID, histCanceledStatuses).
		Group("TO_CHAR(order_time, 'YYYY-MM-DD')").
		Order("order_time_customer DESC").
		Scan(&results).Error
	if results == nil {
		results = []OrderHistoryDateResult{}
	}
	return results, err
}

func (r *OrderTransactionRepository) CountCanceledByMitra(mitraID string) (int64, error) {
	var total int64
	err := r.DB.Model(&models.OrderTransaction{}).
		Where("mitra_id = ? AND order_status IN ? AND id_transaction != ''", mitraID, histCanceledStatuses).
		Count(&total).Error
	return total, err
}

func (r *OrderTransactionRepository) FindCanceledByMitraAndDate(mitraID, startDate, endDate, search string, page, limit int) ([]models.OrderTransaction, int64, error) {
	var orders []models.OrderTransaction
	var total int64
	offset := (page - 1) * limit

	q := r.DB.Model(&models.OrderTransaction{}).
		Where("mitra_id = ? AND order_type != ? AND order_status IN ? AND order_time BETWEEN ? AND ?",
			mitraID, "repeat", histCanceledStatuses, startDate, endDate)
	if search != "" {
		q = q.Where("customer_name ILIKE ?", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Customer").Preload("Mitra").Preload("Service").Preload("SubService").
		Preload("Payment").Preload("SubPayment").Preload("OrderTransactionRepeats").
		Order("order_time DESC").Limit(limit).Offset(offset).Find(&orders).Error
	if orders == nil {
		orders = []models.OrderTransaction{}
	}
	return orders, total, err
}

func (r *OrderTransactionRepository) FindCanceledForCustomerPagedFull(customerID, startDate, endDate string, page, limit int) ([]models.OrderTransaction, int64, error) {
	var orders []models.OrderTransaction
	var total int64
	offset := (page - 1) * limit

	q := r.DB.Model(&models.OrderTransaction{}).
		Where("customer_id = ? AND order_type IN ? AND order_status IN ?",
			customerID, []string{"now", "coming soon"}, histCanceledStatuses)
	if startDate != "" && endDate != "" {
		q = q.Where("order_time BETWEEN ? AND ?", startDate, endDate)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Customer").Preload("Mitra").Preload("Service").Preload("SubService").
		Preload("Payment").Preload("SubPayment").Preload("OrderTransactionRepeats").
		Order("order_time DESC").Limit(limit).Offset(offset).Find(&orders).Error
	if orders == nil {
		orders = []models.OrderTransaction{}
	}
	return orders, total, err
}

// ── Order Dones ───────────────────────────────────────────────────────────────

func (r *OrderTransactionRepository) FindDoneDatesByMitra(mitraID string) ([]OrderHistoryDateResult, error) {
	var results []OrderHistoryDateResult
	err := r.DB.Table("order_transactions").
		Select("TO_CHAR(order_time, 'YYYY-MM-DD') AS order_time_customer, MIN(order_timestamp) AS order_timestamp").
		Where("mitra_id = ? AND order_status = ?", mitraID, "FINISH").
		Group("TO_CHAR(order_time, 'YYYY-MM-DD')").
		Order("order_time_customer DESC").
		Scan(&results).Error
	if results == nil {
		results = []OrderHistoryDateResult{}
	}
	return results, err
}

func (r *OrderTransactionRepository) CountDoneByMitra(mitraID string) (int64, error) {
	var total int64
	err := r.DB.Model(&models.OrderTransaction{}).
		Where("mitra_id = ? AND order_status = ? AND id_transaction != ''", mitraID, "FINISH").
		Count(&total).Error
	return total, err
}

func (r *OrderTransactionRepository) FindDoneByMitraAndDate(mitraID, startDate, endDate, search string, page, limit int) ([]models.OrderTransaction, int64, error) {
	var orders []models.OrderTransaction
	var total int64
	offset := (page - 1) * limit

	q := r.DB.Model(&models.OrderTransaction{}).
		Where("mitra_id = ? AND order_status = ? AND order_time BETWEEN ? AND ?",
			mitraID, "FINISH", startDate, endDate)
	if search != "" {
		q = q.Where("customer_name ILIKE ?", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Customer").Preload("Mitra").Preload("Service").Preload("SubService").
		Preload("Payment").Preload("SubPayment").Preload("OrderTransactionRepeats").
		Order("order_time DESC").Limit(limit).Offset(offset).Find(&orders).Error
	if orders == nil {
		orders = []models.OrderTransaction{}
	}
	return orders, total, err
}

func (r *OrderTransactionRepository) FindDoneForCustomerPagedFull(customerID, startDate, endDate string, page, limit int) ([]models.OrderTransaction, int64, error) {
	var orders []models.OrderTransaction
	var total int64
	offset := (page - 1) * limit

	q := r.DB.Model(&models.OrderTransaction{}).
		Where("customer_id = ? AND order_status = ?", customerID, "FINISH")
	if startDate != "" && endDate != "" {
		q = q.Where("order_time BETWEEN ? AND ?", startDate, endDate)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Customer").Preload("Mitra").Preload("Service").Preload("SubService").
		Preload("Payment").Preload("SubPayment").Preload("OrderTransactionRepeats").
		Order("order_time DESC").Limit(limit).Offset(offset).Find(&orders).Error
	if orders == nil {
		orders = []models.OrderTransaction{}
	}
	return orders, total, err
}

// ── Order Coming Soon ─────────────────────────────────────────────────────────

func (r *OrderTransactionRepository) FindComingSoonDatesByMitra(mitraID string) ([]OrderHistoryDateResult, error) {
	var results []OrderHistoryDateResult
	err := r.DB.Table("order_transactions").
		Select("TO_CHAR(order_time, 'YYYY-MM-DD') AS order_time_customer, MIN(order_timestamp) AS order_timestamp").
		Where("mitra_id = ? AND order_type = ?", mitraID, "coming soon").
		Group("TO_CHAR(order_time, 'YYYY-MM-DD')").
		Order("order_time_customer DESC").
		Scan(&results).Error
	if results == nil {
		results = []OrderHistoryDateResult{}
	}
	return results, err
}

func (r *OrderTransactionRepository) CountComingSoonByMitra(mitraID string) (int64, error) {
	var total int64
	err := r.DB.Model(&models.OrderTransaction{}).
		Where("mitra_id = ? AND order_type = ? AND id_transaction != ''", mitraID, "coming soon").
		Count(&total).Error
	return total, err
}

func (r *OrderTransactionRepository) FindComingSoonByMitraAndDate(mitraID, startDate, endDate, search string, page, limit int) ([]models.OrderTransaction, int64, error) {
	var orders []models.OrderTransaction
	var total int64
	offset := (page - 1) * limit

	q := r.DB.Model(&models.OrderTransaction{}).
		Where("mitra_id = ? AND order_type = ? AND order_status = ? AND order_time BETWEEN ? AND ? AND id_transaction != ''",
			mitraID, "coming soon", "WAIT_SCHEDULE", startDate, endDate)
	if search != "" {
		q = q.Where("customer_name ILIKE ?", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Customer").Preload("Mitra").Preload("Service").Preload("SubService").
		Preload("Payment").Preload("SubPayment").
		Order("id DESC").Limit(limit).Offset(offset).Find(&orders).Error
	if orders == nil {
		orders = []models.OrderTransaction{}
	}
	return orders, total, err
}

func (r *OrderTransactionRepository) FindComingSoonForCustomerPagedFull(customerID, startDate, endDate string, page, limit int) ([]models.OrderTransaction, int64, error) {
	var orders []models.OrderTransaction
	var total int64
	offset := (page - 1) * limit

	q := r.DB.Model(&models.OrderTransaction{}).
		Where("customer_id = ? AND order_type = ? AND order_status = ? AND id_transaction != ''",
			customerID, "coming soon", "WAIT_SCHEDULE")
	if startDate != "" && endDate != "" {
		q = q.Where("order_time BETWEEN ? AND ?", startDate, endDate)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Customer").Preload("Mitra").Preload("Service").Preload("SubService").
		Preload("Payment").Preload("SubPayment").
		Order("id DESC").Limit(limit).Offset(offset).Find(&orders).Error
	if orders == nil {
		orders = []models.OrderTransaction{}
	}
	return orders, total, err
}

// ── Order Repeat ──────────────────────────────────────────────────────────────

func (r *OrderTransactionRepository) FindRepeatDatesByMitra(mitraID string) ([]OrderHistoryDateResult, error) {
	var results []OrderHistoryDateResult
	err := r.DB.Table("order_transactions").
		Select("TO_CHAR(order_time, 'YYYY-MM-DD') AS order_time_customer, MIN(order_timestamp) AS order_timestamp").
		Where("mitra_id = ? AND order_type = ?", mitraID, "repeat").
		Group("TO_CHAR(order_time, 'YYYY-MM-DD')").
		Order("order_time_customer DESC").
		Scan(&results).Error
	if results == nil {
		results = []OrderHistoryDateResult{}
	}
	return results, err
}

func (r *OrderTransactionRepository) CountRepeatByMitra(mitraID string) (int64, error) {
	var total int64
	err := r.DB.Model(&models.OrderTransaction{}).
		Where("mitra_id = ? AND order_type = ? AND id_transaction != ''", mitraID, "repeat").
		Count(&total).Error
	return total, err
}

func (r *OrderTransactionRepository) FindRepeatByMitraAndDate(mitraID, startDate, endDate, search string, page, limit int) ([]models.OrderTransaction, int64, error) {
	var orders []models.OrderTransaction
	var total int64
	offset := (page - 1) * limit

	q := r.DB.Model(&models.OrderTransaction{}).
		Where("mitra_id = ? AND order_type = ? AND order_time BETWEEN ? AND ? AND id_transaction != ''",
			mitraID, "repeat", startDate, endDate)
	if search != "" {
		q = q.Where("customer_name ILIKE ?", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Customer").Preload("Mitra").Preload("Service").Preload("SubService").
		Preload("Payment").Preload("SubPayment").Preload("OrderTransactionRepeats").
		Order("id DESC").Limit(limit).Offset(offset).Find(&orders).Error
	if orders == nil {
		orders = []models.OrderTransaction{}
	}
	return orders, total, err
}

// CountRepeatOrdersForMitra counts all repeat sub-orders belonging to a mitra.
func (r *OrderTransactionRepository) CountRepeatOrdersForMitra(mitraID string) (int64, error) {
	var total int64
	err := r.DB.Raw(`
		SELECT COUNT(a.order_id) FROM order_transaction_repeats a
		LEFT JOIN order_transactions b ON a.order_id = b.id
		WHERE a.mitra_id = ? AND b.order_type = 'repeat'
	`, mitraID).Scan(&total).Error
	return total, err
}

// CountFinishedRepeatForMitra counts finished repeat sub-orders for a mitra.
func (r *OrderTransactionRepository) CountFinishedRepeatForMitra(mitraID string) (int64, error) {
	var total int64
	err := r.DB.Raw(`
		SELECT COUNT(a.order_id) FROM order_transaction_repeats a
		LEFT JOIN order_transactions b ON a.order_id = b.id
		WHERE a.mitra_id = ? AND a.order_status = 'FINISH'
	`, mitraID).Scan(&total).Error
	return total, err
}

func (r *OrderTransactionRepository) FindRepeatForCustomerPagedFull(customerID string, page, limit int) ([]models.OrderTransaction, int64, error) {
	var orders []models.OrderTransaction
	var total int64
	offset := (page - 1) * limit

	q := r.DB.Model(&models.OrderTransaction{}).
		Where("customer_id = ? AND order_type = ?", customerID, "repeat")
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Customer").Preload("Mitra").Preload("Service").Preload("SubService").
		Preload("Payment").Preload("SubPayment").Preload("OrderTransactionRepeats").
		Order("id DESC").Limit(limit).Offset(offset).Find(&orders).Error
	if orders == nil {
		orders = []models.OrderTransaction{}
	}
	return orders, total, err
}

// CountAllRepeatOrders counts all rows in order_transaction_repeats (global, no filter).
func (r *OrderTransactionRepository) CountAllRepeatOrders() (int64, error) {
	var total int64
	err := r.DB.Raw(`
		SELECT COUNT(a.id) FROM order_transaction_repeats a
		LEFT JOIN order_transactions b ON a.order_id = b.id
	`).Scan(&total).Error
	return total, err
}

// CountAllFinishedRepeatOrders counts all FINISH rows in order_transaction_repeats (global, no filter).
func (r *OrderTransactionRepository) CountAllFinishedRepeatOrders() (int64, error) {
	var total int64
	err := r.DB.Raw(`
		SELECT COUNT(a.id) FROM order_transaction_repeats a
		LEFT JOIN order_transactions b ON a.order_id = b.id
		WHERE a.order_status = 'FINISH'
	`).Scan(&total).Error
	return total, err
}

// ── Order Pending ─────────────────────────────────────────────────────────────

func (r *OrderTransactionRepository) CountPendingByCustomer(customerID string) (int64, error) {
	var total int64
	err := r.DB.Model(&models.OrderTransaction{}).
		Where("customer_id = ? AND order_status = ?", customerID, "WAITING_PAYMENT").
		Count(&total).Error
	return total, err
}

func (r *OrderTransactionRepository) FindPendingByMitraPaged(mitraID string, page, limit int) ([]models.OrderTransaction, int64, error) {
	var orders []models.OrderTransaction
	var total int64
	offset := (page - 1) * limit

	q := r.DB.Model(&models.OrderTransaction{}).
		Where("mitra_id = ? AND order_status = ?", mitraID, "WAITING_PAYMENT")
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Customer").Preload("Mitra").Preload("Service").Preload("SubService").
		Preload("Payment").Preload("SubPayment").
		Order("id DESC").Limit(limit).Offset(offset).Find(&orders).Error
	if orders == nil {
		orders = []models.OrderTransaction{}
	}
	return orders, total, err
}

func (r *OrderTransactionRepository) FindPendingByCustomerPaged(customerID string, page, limit int) ([]models.OrderTransaction, int64, error) {
	var orders []models.OrderTransaction
	var total int64
	offset := (page - 1) * limit

	q := r.DB.Model(&models.OrderTransaction{}).
		Where("customer_id = ? AND order_status = ?", customerID, "WAITING_PAYMENT")
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Customer").Preload("Mitra").Preload("Service").Preload("SubService").
		Preload("Payment").Preload("SubPayment").
		Order("id DESC").Limit(limit).Offset(offset).Find(&orders).Error
	if orders == nil {
		orders = []models.OrderTransaction{}
	}
	return orders, total, err
}

// ── Order Running ─────────────────────────────────────────────────────────────

func (r *OrderTransactionRepository) FindRunningForCustomerPagedFull(customerID string, page, limit int) ([]models.OrderTransaction, int64, error) {
	var orders []models.OrderTransaction
	var total int64
	offset := (page - 1) * limit

	q := r.DB.Model(&models.OrderTransaction{}).
		Where("customer_id = ? AND order_status IN ? AND id_transaction != ''",
			customerID, histRunningStatuses)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Customer").Preload("Mitra").Preload("Service").Preload("SubService").
		Preload("Payment").Preload("SubPayment").Preload("OrderTransactionRepeats").
		Order("id DESC").Limit(limit).Offset(offset).Find(&orders).Error
	if orders == nil {
		orders = []models.OrderTransaction{}
	}
	return orders, total, err
}
