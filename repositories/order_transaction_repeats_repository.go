package repositories

import (
	"suberes_golang/dtos"
	"time"

	"gorm.io/gorm"
)

type OrderTransactionRepeatsRepository struct {
	DB *gorm.DB
}

func (r *OrderTransactionRepeatsRepository) GetTodayRepeatSummary(mitraID int, start, end time.Time) (*dtos.OrderSummary, error) {
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
