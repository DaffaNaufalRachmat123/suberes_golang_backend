package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type OrderOfferRepository struct {
	DB *gorm.DB
}

func (r *OrderOfferRepository) FindOneByWhere(tx *gorm.DB, where map[string]interface{}) (*models.OrderOffer, error) {
	var order models.OrderOffer

	db := r.DB
	if tx != nil {
		db = tx
	}

	err := db.
		Model(&models.OrderOffer{}).
		Where(where).
		First(&order).Error

	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *OrderOfferRepository) DeleteByWhere(tx *gorm.DB, where map[string]interface{}) error {

	db := r.DB
	if tx != nil {
		db = tx
	}

	err := db.
		Where(where).
		Delete(&models.OrderOffer{}).Error

	if err != nil {
		return err
	}

	return nil
}

// FindPaginatedByMitraID returns a paginated list of order offers for a given mitra,
// including the associated order_transaction (with service, sub_service, customer).
func (r *OrderOfferRepository) FindPaginatedByMitraID(mitraID string, page, limit int) ([]models.OrderOffer, int64, error) {
	var offers []models.OrderOffer
	var total int64

	offset := (page - 1) * limit

	query := r.DB.Model(&models.OrderOffer{}).Where("order_offers.mitra_id = ?", mitraID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("OrderTransaction", func(db *gorm.DB) *gorm.DB {
			return db.
				Preload("Service").
				Preload("SubService").
				Preload("Customer")
		}).
		Order("order_offers.id DESC").
		Limit(limit).
		Offset(offset).
		Find(&offers).Error

	return offers, total, err
}

// FindDetailByOrderAndMitra returns a single order offer with full order detail and computed countdown.
// countdownSQL is a raw SQL expression for the countdown column.
func (r *OrderOfferRepository) FindDetailByOrderAndMitra(orderID, mitraID, countdownSQL string) (*models.OrderOffer, error) {
	var offer models.OrderOffer

	err := r.DB.Model(&models.OrderOffer{}).
		Where("order_offers.order_id = ? AND order_offers.mitra_id = ?", orderID, mitraID).
		Preload("OrderTransaction", func(db *gorm.DB) *gorm.DB {
			return db.
				Select("order_transactions.*, "+countdownSQL+" AS count_down_can_take_order").
				Preload("Service").
				Preload("SubService").
				Preload("Payment").
				Preload("SubPayment").
				Preload("Customer").
				Preload("OrderTransactionRepeats").
				Preload("SubServiceAddeds", func(db *gorm.DB) *gorm.DB {
					return db.Preload("SubServiceAdditional")
				})
		}).
		First(&offer).Error

	if err != nil {
		return nil, err
	}

	return &offer, nil
}
func (r *OrderOfferRepository) CountByMitraID(mitraID string) (int64, error) {
	var count int64
	err := r.DB.Model(&models.OrderOffer{}).Where("mitra_id = ?", mitraID).Count(&count).Error
	return count, err
}
