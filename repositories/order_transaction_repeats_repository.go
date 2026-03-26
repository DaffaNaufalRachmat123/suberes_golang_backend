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
