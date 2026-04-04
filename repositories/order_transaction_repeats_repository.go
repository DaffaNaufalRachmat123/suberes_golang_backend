package repositories

import (
	"suberes_golang/dtos"
	"suberes_golang/models"
	"time"

	"gorm.io/gorm"
)

type OrderTransactionRepeatsRepository struct {
	DB *gorm.DB
}

func (r *OrderTransactionRepeatsRepository) GetTodayRepeatSummary(mitraID string, start, end time.Time) (*dtos.OrderSummary, error) {
	var result struct {
		OrderCount int64
		Pendapatan int64
	}

	err := r.DB.
		Table("order_transaction_repeats").
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

func (r *OrderTransactionRepeatsRepository) UpdateRepeatOrderStatus(tx *gorm.DB, orderID, status string) error {
	return tx.Model(&models.OrderTransactionRepeat{}).Where("order_id = ?", orderID).Update("order_status", status).Error
}
func (r *OrderTransactionRepeatsRepository) FindByWhereDynamic(
	tx *gorm.DB,
	where map[string]interface{},
) ([]models.OrderTransactionRepeat, error) {

	var orders []models.OrderTransactionRepeat

	db := r.DB
	if tx != nil {
		db = tx
	}

	err := db.
		Model(&models.OrderTransactionRepeat{}).
		Where(where).
		Find(&orders).Error

	if err != nil {
		return nil, err
	}

	return orders, nil
}

// ---- New methods ----

func (r *OrderTransactionRepeatsRepository) FindByOrderAndSubID(orderID string, subID int) (*models.OrderTransactionRepeat, error) {
	var repeat models.OrderTransactionRepeat
	err := r.DB.Where("order_id = ? AND id = ?", orderID, subID).
		Preload("OrderTransaction").Preload("Customer").Preload("Mitra").
		Preload("Service").Preload("SubService").
		First(&repeat).Error
	if err != nil {
		return nil, err
	}
	return &repeat, nil
}

func (r *OrderTransactionRepeatsRepository) FindBySubID(subID int) (*models.OrderTransactionRepeat, error) {
	var repeat models.OrderTransactionRepeat
	err := r.DB.Where("id = ?", subID).
		Preload("OrderTransaction").Preload("Customer").Preload("Mitra").
		Preload("Service").Preload("SubService").
		First(&repeat).Error
	if err != nil {
		return nil, err
	}
	return &repeat, nil
}

func (r *OrderTransactionRepeatsRepository) UpdateRepeatData(tx *gorm.DB, subID int, data map[string]interface{}) error {
	return tx.Model(&models.OrderTransactionRepeat{}).Where("id = ?", subID).Updates(data).Error
}

func (r *OrderTransactionRepeatsRepository) UpdateRepeatByOrderAndSubID(tx *gorm.DB, orderID string, subID int, data map[string]interface{}) error {
	return tx.Model(&models.OrderTransactionRepeat{}).Where("order_id = ? AND id = ?", orderID, subID).Updates(data).Error
}

func (r *OrderTransactionRepeatsRepository) FindAllByOrderID(orderID string) ([]models.OrderTransactionRepeat, error) {
	var repeats []models.OrderTransactionRepeat
	err := r.DB.Where("order_id = ?", orderID).Find(&repeats).Error
	return repeats, err
}

// FindRepeatListByOrderPaged returns a paginated list of repeat sub-orders
// filtered by order_id, mitra_id, and customer_id.
func (r *OrderTransactionRepeatsRepository) FindRepeatListByOrderPaged(orderID, mitraID, customerID string, page, limit int) ([]models.OrderTransactionRepeat, int64, error) {
	var repeats []models.OrderTransactionRepeat
	var total int64
	offset := (page - 1) * limit

	q := r.DB.Model(&models.OrderTransactionRepeat{}).
		Where("order_id = ? AND mitra_id = ? AND customer_id = ?", orderID, mitraID, customerID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("SubService").
		Order("id ASC").Limit(limit).Offset(offset).Find(&repeats).Error
	return repeats, total, err
}
