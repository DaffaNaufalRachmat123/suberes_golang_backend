package repositories

import (
	"suberes_golang/models"
)

// ─── Result types ─────────────────────────────────────────────────────────────

// PendapatanDateResult holds distinct order-date entries for the date picker list.
type PendapatanDateResult struct {
	OrderTimeCustomer string `gorm:"column:order_time_customer" json:"order_time_customer"`
	OrderTimestamp    string `gorm:"column:order_timestamp"     json:"order_timestamp"`
}

// PendapatanCalculate is the aggregate summary for a given date range.
type PendapatanCalculate struct {
	OrderCount  int64 `gorm:"column:order_count"  json:"order_count"`
	TotalAmount int64 `gorm:"column:total_amount" json:"total_amount"`
}

// PendapatanGroupResult is the per-payment-method breakdown.
type PendapatanGroupResult struct {
	PaymentID   int            `gorm:"column:payment_id"   json:"payment_id"`
	MitraID     string         `gorm:"column:mitra_id"     json:"mitra_id"`
	TotalAmount int64          `gorm:"column:total_amount" json:"total_amount"`
	Payment     models.Payment `gorm:"-"                 json:"payment"`
}

// PendapatanOrderItem mirrors the order list entry returned by the Node.js route,
// adding the human-readable order_type_label.
type PendapatanOrderItem struct {
	models.OrderTransaction
	OrderTypeLabel string `json:"order_type_label"`
}

// ─── Queries ──────────────────────────────────────────────────────────────────

// FindPendapatanDates returns all distinct calendar dates (YYYY-MM-DD) on which
// the mitra has at least one FINISH order — used to populate the date-picker list.
func (r *MitraRepository) FindPendapatanDates(mitraID string) ([]PendapatanDateResult, error) {
	var results []PendapatanDateResult
	err := r.DB.Table("order_transactions").
		Select("TO_CHAR(order_time, 'YYYY-MM-DD') AS order_time_customer, MIN(order_timestamp) AS order_timestamp").
		Where("mitra_id = ? AND order_status = ?", mitraID, "FINISH").
		Group("TO_CHAR(order_time, 'YYYY-MM-DD')").
		Order("order_time_customer DESC").
		Scan(&results).Error
	if results == nil {
		results = []PendapatanDateResult{}
	}
	return results, err
}

// CalculatePendapatan aggregates order_count and total gross_amount_mitra for
// FINISH orders within the given date range (inclusive, "YYYY-MM-DD" strings).
func (r *MitraRepository) CalculatePendapatan(mitraID, startDate, endDate string) (*PendapatanCalculate, error) {
	var result PendapatanCalculate
	err := r.DB.Table("order_transactions").
		Select("COUNT(id) AS order_count, COALESCE(SUM(gross_amount_mitra), 0) AS total_amount").
		Where("mitra_id = ? AND order_status = ? AND order_time BETWEEN ? AND ?",
			mitraID, "FINISH", startDate, endDate).
		Scan(&result).Error
	return &result, err
}

// GroupPendapatanByPayment aggregates total income per payment method within the
// date range and eager-loads the related Payment record.
func (r *MitraRepository) GroupPendapatanByPayment(mitraID, startDate, endDate string) ([]PendapatanGroupResult, error) {
	var rows []PendapatanGroupResult
	err := r.DB.Table("order_transactions").
		Select("payment_id, mitra_id, COALESCE(SUM(gross_amount_mitra), 0) AS total_amount").
		Where("mitra_id = ? AND order_status = ? AND order_time BETWEEN ? AND ?",
			mitraID, "FINISH", startDate, endDate).
		Group("payment_id, mitra_id").
		Order("total_amount DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []PendapatanGroupResult{}
	}

	// Eager-load Payment for each group row
	for i := range rows {
		var payment models.Payment
		if err := r.DB.Where("id = ?", rows[i].PaymentID).First(&payment).Error; err == nil {
			rows[i].Payment = payment
		}
	}
	return rows, nil
}

// FindOrderListByPayment fetches the full order list for a specific payment method
// within the date range, including the service and repeat-transaction associations,
// plus a human-readable order type label.
func (r *MitraRepository) FindOrderListByPayment(mitraID, startDate, endDate string, paymentID int) ([]PendapatanOrderItem, error) {
	var orders []models.OrderTransaction
	err := r.DB.
		Preload("Service").
		Preload("OrderTransactionRepeats").
		Where("mitra_id = ? AND order_status = ? AND payment_id = ? AND order_time BETWEEN ? AND ?",
			mitraID, "FINISH", paymentID, startDate, endDate).
		Order("order_time DESC").
		Find(&orders).Error
	if err != nil {
		return nil, err
	}

	items := make([]PendapatanOrderItem, 0, len(orders))
	for _, o := range orders {
		label := "-"
		switch o.OrderType {
		case "now", "coming soon":
			label = "Order Sekali"
		case "repeat":
			label = "Order Berulang"
		}
		items = append(items, PendapatanOrderItem{
			OrderTransaction: o,
			OrderTypeLabel:   label,
		})
	}
	return items, nil
}
