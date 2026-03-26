package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type OrderRepository struct {
	DB *gorm.DB
}

func (r *OrderRepository) CreateOrder(tx *gorm.DB, order *models.OrderTransaction) error {
	return tx.Create(order).Error
}

func (r *OrderRepository) CreateOrderRepeats(tx *gorm.DB, orderRepeats []models.OrderTransactionRepeat) error {
	return tx.Create(&orderRepeats).Error
}

func (r *OrderRepository) CreateSubServiceAdded(tx *gorm.DB, subServiceAdded []models.SubServiceAdded) error {
	return tx.Create(&subServiceAdded).Error
}

func (r *OrderRepository) UpdateOrder(tx *gorm.DB, order *models.OrderTransaction) error {
	return tx.Save(order).Error
}

func (r *OrderRepository) FindOrderByID(orderID string) (*models.OrderTransaction, error) {
	var order models.OrderTransaction
	err := r.DB.Where("id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) FindOrder(condition map[string]interface{}) (*models.OrderTransaction, error) {
	var order models.OrderTransaction
	err := r.DB.Where(condition).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) UpdateOrderTransaction(tx *gorm.DB, condition map[string]interface{}, data map[string]interface{}) error {
	return tx.Model(&models.OrderTransaction{}).Where(condition).Updates(data).Error
}

func (r *OrderRepository) UpdateOrderTransactionRepeats(tx *gorm.DB, condition map[string]interface{}, data map[string]interface{}) error {
	return tx.Model(&models.OrderTransactionRepeat{}).Where(condition).Updates(data).Error
}

func (r *OrderRepository) DeleteOrderOffers(tx *gorm.DB, orderID string) error {
	return tx.Where("order_id = ?", orderID).Delete(&models.OrderOffer{}).Error
}

func (r *OrderRepository) CreateOrderChat(tx *gorm.DB, chat *models.OrderChat) error {
	return tx.Create(chat).Error
}
