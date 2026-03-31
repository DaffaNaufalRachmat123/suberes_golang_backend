package services

import (
"errors"
"suberes_golang/dtos"
"suberes_golang/models"
"suberes_golang/repositories"

"gorm.io/gorm"
)

type TermsConditionService struct {
	TermsConditionRepo *repositories.TermsConditionRepository
	DB                 *gorm.DB
}

// GetAllTermsConditions retrieves all terms conditions with pagination
func (s *TermsConditionService) GetAllTermsConditions(page, limit int) ([]models.TermsCondition, int64, error) {
	return s.TermsConditionRepo.FindAll(page, limit)
}

// GetTermsConditionByID retrieves a single terms condition by ID
func (s *TermsConditionService) GetTermsConditionByID(id uint) (*models.TermsCondition, error) {
	return s.TermsConditionRepo.FindByID(id)
}

// GetTermsConditionByTypeAndUserType retrieves active terms condition for user
func (s *TermsConditionService) GetTermsConditionByTypeAndUserType(tocType, tocUserType string) (*models.TermsCondition, error) {
	return s.TermsConditionRepo.FindByTypeAndUserType(tocType, tocUserType, "1")
}

// CreateTermsCondition creates a new terms condition with validation
func (s *TermsConditionService) CreateTermsCondition(req *dtos.TermsConditionCreateRequest, creatorID string, force bool) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if there's an active TOC for this type and user type (unless force is true)
if req.IsActive == "1" && !force {
existingActive, err := s.TermsConditionRepo.FindByTypeAndUserType(req.TocType, req.TocUserType, "1")
if err != nil {
tx.Rollback()
return err
}
if existingActive != nil {
tx.Rollback()
return errors.New("There is still an active TOC for this user")
}
}

// If setting as active, deactivate others
canSelect := "0"
if req.IsActive == "1" {
canSelect = "1"
if err := s.TermsConditionRepo.UpdateMultipleByTypeAndUserType(
tx,
req.TocType,
req.TocUserType,
map[string]interface{}{
"can_select": "0",
"is_active":  "0",
},
0, // no exclude ID on create
); err != nil {
tx.Rollback()
return err
}
}

termsCondition := &models.TermsCondition{
CreatorID:   creatorID,
Title:       req.Title,
Body:        req.Body,
IsActive:    req.IsActive,
CanSelect:   canSelect,
TocType:     req.TocType,
TocUserType: req.TocUserType,
}

if err := s.TermsConditionRepo.Create(tx, termsCondition); err != nil {
tx.Rollback()
return err
}

return tx.Commit().Error
}

// UpdateTermsCondition updates an existing terms condition
func (s *TermsConditionService) UpdateTermsCondition(id uint, req *dtos.TermsConditionUpdateRequest, creatorID string, force bool) error {
tx := s.DB.Begin()
defer func() {
if r := recover(); r != nil {
tx.Rollback()
}
}()

// Find existing TOC
existingTOC, err := s.TermsConditionRepo.FindByID(id)
if err != nil {
tx.Rollback()
return err
}
if existingTOC == nil {
tx.Rollback()
return errors.New("There are no TOC data")
}

// Check if there's another active TOC for the same type and user type
	activeData, err := s.TermsConditionRepo.FindActiveExcept(req.TocType, req.TocUserType, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Handle activation logic
	if req.IsActive == "1" {
		if !force && activeData != nil {
			tx.Rollback()
			return errors.New("Masih ada TOC aktif dengan kategori untuk user ini")
		}

		// If forcing or no active data, deactivate others
		if force || activeData != nil {
			if err := s.TermsConditionRepo.UpdateMultipleByTypeAndUserType(
tx,
req.TocType,
req.TocUserType,
map[string]interface{}{
"can_select": "0",
"is_active":  "0",
},
id,
); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// Handle deactivation logic - enable others if this was active
	if existingTOC.IsActive == "1" && req.IsActive == "0" {
		if err := s.TermsConditionRepo.UpdateMultipleByTypeAndUserType(
tx,
req.TocType,
req.TocUserType,
map[string]interface{}{
"can_select": "1",
},
id,
); err != nil {
			tx.Rollback()
			return err
		}
	}

	// Update the TOC
	existingTOC.CreatorID = creatorID
	existingTOC.Title = req.Title
	existingTOC.Body = req.Body
	existingTOC.IsActive = req.IsActive
	existingTOC.TocType = req.TocType
	existingTOC.TocUserType = req.TocUserType

	if req.IsActive == "1" {
		existingTOC.CanSelect = "1"
	}

	if err := s.TermsConditionRepo.Update(tx, existingTOC); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// UpdateTermsConditionStatus updates the status of a terms condition
func (s *TermsConditionService) UpdateTermsConditionStatus(id uint, req *dtos.TermsConditionUpdateStatusRequest) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if the TOC exists with the specified type and user type
	existingTOC, err := s.TermsConditionRepo.FindByID(id)
	if err != nil {
		tx.Rollback()
		return err
	}
	if existingTOC == nil {
		tx.Rollback()
		return errors.New("TOC not found")
	}

	// Validate that the TOC belongs to the specified type and user type
	if existingTOC.TocType != req.TocType || existingTOC.TocUserType != req.TocUserType {
		tx.Rollback()
		return errors.New("TOC type or user type mismatch")
	}

	// Handle activation
	if req.IsActive == "1" {
		// Check if there's another active TOC
activeData, err := s.TermsConditionRepo.FindActiveExcept(req.TocType, req.TocUserType, id)
if err != nil {
tx.Rollback()
return err
}
if activeData != nil {
tx.Rollback()
return errors.New("Masih ada TOC yang aktif untuk sasaran ini")
}

// Deactivate others
if err := s.TermsConditionRepo.UpdateMultipleByTypeAndUserType(
tx,
req.TocType,
req.TocUserType,
map[string]interface{}{
"can_select": "0",
"is_active":  "0",
},
id,
); err != nil {
tx.Rollback()
return err
}

// Activate this one
if err := s.TermsConditionRepo.UpdateStatus(tx, id, "1", "1"); err != nil {
tx.Rollback()
return err
}
} else {
// Handle deactivation - activate another TOC if available
if existingTOC.IsActive == "1" {
if err := s.TermsConditionRepo.UpdateMultipleByTypeAndUserType(
tx,
req.TocType,
req.TocUserType,
map[string]interface{}{
"can_select": "1",
},
id,
); err != nil {
tx.Rollback()
return err
}
}

// Deactivate this one
if err := s.TermsConditionRepo.UpdateStatus(tx, id, "0", existingTOC.CanSelect); err != nil {
tx.Rollback()
return err
}
}

return tx.Commit().Error
}

// DeleteTermsCondition deletes a terms condition
func (s *TermsConditionService) DeleteTermsCondition(id uint) error {
tx := s.DB.Begin()
defer func() {
if r := recover(); r != nil {
tx.Rollback()
}
}()

// Find existing TOC
existingTOC, err := s.TermsConditionRepo.FindByID(id)
if err != nil {
tx.Rollback()
return err
}
if existingTOC == nil {
tx.Rollback()
return errors.New("Data TOC tak ditemukan")
}

// Cannot delete if active
if existingTOC.IsActive == "1" {
tx.Rollback()
return errors.New("Harap nonaktifkan TOC dahulu")
}

if err := s.TermsConditionRepo.Delete(tx, existingTOC); err != nil {
tx.Rollback()
return err
}

return tx.Commit().Error
}
