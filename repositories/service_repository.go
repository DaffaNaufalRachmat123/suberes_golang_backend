package repositories

import (
	"fmt"
	"suberes_golang/models"

	"gorm.io/gorm"
)

type ServiceRepository struct {
	DB *gorm.DB
}

func (r *ServiceRepository) FindAllPagination(parent_id, page, limit int) ([]models.Service, int64, error) {
	var services []models.Service
	var total int64
	offset := (page - 1) * limit
	query := r.DB.Model(models.Service{}).Where("parent_id = ?", parent_id)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Limit(limit).Offset(offset).Find(&services).Error; err != nil {
		return nil, 0, err
	}
	return services, total, nil
}
func (r *ServiceRepository) FindByID(id int) (*models.Service, error) {
	var service models.Service
	if err := r.DB.First(&service, id).Error; err != nil {
		return nil, err
	}
	return &service, nil
}
func (r *ServiceRepository) Search(layananID int, serviceName string) ([]models.Service, error) {
	var services []models.Service

	err := r.DB.
		Model(&models.Service{}).
		Joins("JOIN category_services ON category_services.id = services.category_service_id").
		Where("category_services.layanan_id = ?", layananID).
		Where("services.service_name LIKE ?", fmt.Sprintf("%%%s%%", serviceName)).
		Preload("CategoryService").
		Find(&services).Error

	if err != nil {
		return nil, err
	}

	return services, nil
}
func (r *ServiceRepository) FindLayananServices(id int) ([]models.LayananService, error) {
	var layanan []models.LayananService

	err := r.DB.
		Preload("CategoryServices.Services").
		Where("id = ?", id).
		Find(&layanan).Error

	if err != nil {
		return nil, err
	}

	return layanan, nil
}
func (r *ServiceRepository) GetRunningService(serviceID, subServiceID int) (*models.Service, error) {
	var service models.Service

	err := r.DB.
		Preload("SubServices", "id = ?", subServiceID).
		Where("id = ?", serviceID).
		First(&service).Error

	if err != nil {
		return nil, err
	}

	return &service, nil
}
func (r *ServiceRepository) FindServiceType(serviceID int) (*models.CategoryService, error) {
	var category models.CategoryService

	err := r.DB.
		Preload("Services", "id = ?", serviceID).
		Preload("Services.SubServices").
		Preload("Services.SubServices.SubServiceAdditionals").
		Joins("JOIN services ON services.parent_id = category_services.id").
		Where("services.id = ?", serviceID).
		First(&category).Error

	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *ServiceRepository) FindPopular(limit int) ([]models.Service, error) {
	var services []models.Service
	err := r.DB.Limit(limit).Find(&services).Error
	return services, err
}
func (r *ServiceRepository) Create(tx *gorm.DB, service *models.Service) error {
	return tx.Create(service).Error
}
func (r *ServiceRepository) Update(tx *gorm.DB, id int, data map[string]interface{}) error {
	return tx.Model(&models.Service{}).Where("id = ?", id).Updates(data).Error
}
func (r *ServiceRepository) Delete(tx *gorm.DB, id int) error {
	return tx.Delete(&models.Service{}, id).Error
}
