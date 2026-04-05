package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type ScheduleRepository struct {
	DB *gorm.DB
}

func (r *ScheduleRepository) FindAll(page, limit int, search string) ([]models.Schedule, int64, error) {
	var schedules []models.Schedule
	var total int64

	db := r.DB.Model(&models.Schedule{})

	if search != "" {
		db = db.Where("schedule_name LIKE ? OR schedule_title LIKE ? OR schedule_place LIKE ? OR schedule_message LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&schedules).Error
	if err != nil {
		return nil, 0, err
	}

	return schedules, total, nil
}

func (r *ScheduleRepository) FindAllMitra(page, limit int, search string) ([]models.Schedule, int64, error) {
	var schedules []models.Schedule
	var total int64

	db := r.DB.Model(&models.Schedule{})
	db = db.Where("schedule_level = ?", "mitra_level")

	if search != "" {
		db = db.Where("schedule_name LIKE ? OR schedule_title LIKE ? OR schedule_place LIKE ? OR schedule_message LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = db.Offset(offset).Limit(limit).Order("id DESC").Find(&schedules).Error
	if err != nil {
		return nil, 0, err
	}

	return schedules, total, nil
}

func (r *ScheduleRepository) FindByID(id string) (*models.Schedule, error) {
	var schedule models.Schedule
	err := r.DB.Preload("Creator").First(&schedule, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (r *ScheduleRepository) FindByNamePlaceAndDate(name, place, dateTime string) (*models.Schedule, error) {
	var schedule models.Schedule
	err := r.DB.Where("schedule_name = ? AND schedule_place = ? AND schedule_date_time = ?", name, place, dateTime).First(&schedule).Error
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (r *ScheduleRepository) Create(tx *gorm.DB, schedule *models.Schedule) error {
	return tx.Create(schedule).Error
}

func (r *ScheduleRepository) Update(tx *gorm.DB, schedule *models.Schedule) error {
	return tx.Save(schedule).Error
}

func (r *ScheduleRepository) Delete(tx *gorm.DB, id string) error {
	return tx.Where("id = ?", id).Delete(&models.Schedule{}).Error
}

func (r *ScheduleRepository) FindMitraLevelByID(id int64) (*models.Schedule, error) {
	var schedule models.Schedule
	err := r.DB.Where("id = ? AND schedule_level = ?", id, "mitra_level").First(&schedule).Error
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}
