package services

import (
	"errors"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/repositories"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ScheduleService struct {
	ScheduleRepo *repositories.ScheduleRepository
	UserRepo     *repositories.UserRepository
	DB           *gorm.DB
}

func NewScheduleService(scheduleRepo *repositories.ScheduleRepository, userRepo *repositories.UserRepository, db *gorm.DB) *ScheduleService {
	return &ScheduleService{
		ScheduleRepo: scheduleRepo,
		UserRepo:     userRepo,
		DB:           db,
	}
}

const scheduleTemplate = `<!DOCTYPE HTML PUBLIC "-//W3C//DTD XHTML 1.0 Transitional //EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd"> <html xmlns="http://www.w3.org/1999/xhtml" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office"> <head> <meta http-equiv="Content-Type" content="text/html; charset=UTF-8"> <meta name="viewport" content="width=device-width, initial-scale=1.0"> <meta name="x-apple-disable-message-reformatting"> <meta http-equiv="X-UA-Compatible" content="IE=edge"> <title></title> <style type="text/css"> @media only screen and (min-width: 520px) { .u-row { width: 500px !important; } .u-row .u-col { vertical-align: top; } .u-row .u-col-100 { width: 500px !important; } } @media (max-width: 520px) { .u-row-container { max-width: 100% !important; padding-left: 0px !important; padding-right: 0px !important; } .u-row .u-col { min-width: 320px !important; max-width: 100% !important; display: block !important; } .u-row { width: calc(100% - 40px) !important; } .u-col { width: 100% !important; } .u-col>div { margin: 0 auto; } } body { margin: 0; padding: 0; } table, tr, td { vertical-align: top; border-collapse: collapse; } p { margin: 0; } .ie-container table, .mso-container table { table-layout: fixed; } * { line-height: inherit; } a[x-apple-data-detectors='true'] { color: inherit !important; text-decoration: none !important; } table, td { color: #000000; } </style> <link href="https://fonts.googleapis.com/css?family=Montserrat:400,700" rel="stylesheet" type="text/css"> </head> <body class="clean-body u_body" style="margin: 0;padding: 0;-webkit-text-size-adjust: 100%;background-color: #ffffff;color: #000000"> <table style="border-collapse: collapse;table-layout: fixed;border-spacing: 0;mso-table-lspace: 0pt;mso-table-rspace: 0pt;vertical-align: top;min-width: 320px;Margin: 0 auto;background-color: #ffffff;width:100%" cellpadding="0" cellspacing="0"> <tbody> <tr style="vertical-align: top"> <td style="word-break: break-word;border-collapse: collapse !important;vertical-align: top"></tbody></table></body></html>`

func (s *ScheduleService) Index(page, limit int, search string) ([]models.Schedule, int64, error) {
	return s.ScheduleRepo.FindAll(page, limit, search)
}

func (s *ScheduleService) IndexMitra(page, limit int, search string) ([]models.Schedule, int64, error) {
	return s.ScheduleRepo.FindAllMitra(page, limit, search)
}

func (s *ScheduleService) Detail(id string) (*models.Schedule, error) {
	return s.ScheduleRepo.FindByID(id)
}

func (s *ScheduleService) Create(req *dtos.CreateScheduleRequest) (*models.Schedule, error) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	valid, err := helpers.IsScheduleDateValid(req.ScheduleDateTime, req.TimezoneCode)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if !valid {
		tx.Rollback()
		return nil, errors.New("schedule's date cannot be less than today or the same as today")
	}

	admin, err := s.UserRepo.FindAdminById(req.CreatorID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("admin not found")
	}

	existingSchedule, _ := s.ScheduleRepo.FindByNamePlaceAndDate(req.ScheduleName, req.SchedulePlace, req.ScheduleDateTime)
	if existingSchedule != nil {
		tx.Rollback()
		return nil, errors.New("schedule already exist")
	}

	schedule := &models.Schedule{
		CreatorID:        admin.ID,
		CreatorName:      req.CreatorName,
		ScheduleName:     req.ScheduleName,
		ScheduleLevel:    req.ScheduleLevel,
		ScheduleTitle:    req.ScheduleTitle,
		SchedulePlace:    req.SchedulePlace,
		ScheduleDateTime: req.ScheduleDateTime,
		ScheduleMessage:  req.ScheduleMessage,
		ScheduleIsActive: req.ScheduleIsActive,
		TimezoneCode:     req.TimezoneCode,
		ScheduleTemplate: scheduleTemplate,
	}

	if err := s.ScheduleRepo.Create(tx, schedule); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return schedule, nil
}

func (s *ScheduleService) Update(id string, req *dtos.UpdateScheduleRequest) (*models.Schedule, error) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	valid, err := helpers.IsScheduleDateValid(req.ScheduleDateTime, req.TimezoneCode)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if !valid {
		tx.Rollback()
		return nil, errors.New("schedule's date cannot be less than today or the same as today")
	}

	admin, err := s.UserRepo.FindAdminById(req.CreatorID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("admin not found")
	}

	schedule, err := s.ScheduleRepo.FindByID(id)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("schedule not found")
	}

	schedule.CreatorID = admin.ID
	schedule.CreatorName = req.CreatorName
	schedule.ScheduleName = req.ScheduleName
	schedule.ScheduleLevel = req.ScheduleLevel
	schedule.ScheduleTitle = req.ScheduleTitle
	schedule.SchedulePlace = req.SchedulePlace
	schedule.ScheduleDateTime = req.ScheduleDateTime
	schedule.ScheduleMessage = req.ScheduleMessage
	schedule.ScheduleIsActive = req.ScheduleIsActive
	schedule.TimezoneCode = req.TimezoneCode
	schedule.ScheduleTemplate = scheduleTemplate

	if err := s.ScheduleRepo.Update(tx, schedule); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	return schedule, nil
}

func (s *ScheduleService) Delete(id string, adminPassword string, adminId string) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	admin, err := s.UserRepo.FindAdminById(adminId)
	if err != nil {
		tx.Rollback()
		return errors.New("admin not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(adminPassword)); err != nil {
		tx.Rollback()
		return errors.New("unauthorized")
	}

	_, err = s.ScheduleRepo.FindByID(id)
	if err != nil {
		tx.Rollback()
		return errors.New("schedule not found")
	}

	if err := s.ScheduleRepo.Delete(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
