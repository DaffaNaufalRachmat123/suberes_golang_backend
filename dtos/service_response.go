package dtos

import (
	"suberes_golang/models"
	"time"
)

type ServiceGuaranteeDTO struct {
	ID                   int    `json:"id"`
	ServiceID            int    `json:"service_id"`
	UserID               string `json:"user_id"`
	GuaranteeName        string `json:"guarantee_name"`
	GuaranteeDescription string `json:"guarantee_description"`
	IsGuaranteeEnabled   string `json:"is_guarantee_enabled"`
}

type ServiceDTO struct {
	ID                    int                  `json:"id"`
	ServiceName           string               `json:"service_name"`
	ServiceDescription    string               `json:"service_description"`
	ServiceImageThumbnail string               `json:"service_image_thumbnail"`
	ServiceCount          int                  `json:"service_count"`
	ServiceType           string               `json:"service_type"`
	ServiceCategory       string               `json:"service_category"`
	CreatedAt             time.Time            `json:"createdAt"`
	MaxOrderCount         int                  `json:"max_order_count"`
	PaymentTimeout        int                  `json:"payment_timeout"`
	UpdatedAt             time.Time            `json:"updatedAt"`
	ServiceStatus         string               `json:"service_status"`
	ServiceGuarantee      *ServiceGuaranteeDTO `json:"service_guarantee"`
	SubServices           []models.SubService  `json:"sub_services"`
}

type CategoryServiceSearchResponse struct {
	ID              int          `json:"id"`
	LayananID       int          `json:"layanan_id"`
	CreatorID       string       `json:"creator_id"`
	CategoryService string       `json:"category_service"`
	Services        []ServiceDTO `json:"services"`
}

func toServiceDTO(s models.Service) ServiceDTO {
	var guarantee *ServiceGuaranteeDTO
	if s.ServiceGuarantee != nil {
		g := s.ServiceGuarantee
		guarantee = &ServiceGuaranteeDTO{
			ID:                   g.ID,
			ServiceID:            g.ServiceID,
			UserID:               g.UserID,
			GuaranteeName:        g.GuaranteeName,
			GuaranteeDescription: g.GuaranteeDescription,
			IsGuaranteeEnabled:   g.IsGuaranteeEnabled,
		}
	}
	subServices := s.SubServices
	if subServices == nil {
		subServices = []models.SubService{}
	}
	return ServiceDTO{
		ID:                    s.ID,
		ServiceName:           s.ServiceName,
		ServiceDescription:    s.ServiceDescription,
		ServiceImageThumbnail: s.ServiceImageThumbnail,
		ServiceCount:          s.ServiceCount,
		ServiceType:           s.ServiceType,
		ServiceCategory:       s.ServiceCategory,
		CreatedAt:             s.CreatedAt,
		MaxOrderCount:         s.MaxOrderCount,
		PaymentTimeout:        s.PaymentTimeout,
		UpdatedAt:             s.UpdatedAt,
		ServiceStatus:         s.ServiceStatus,
		ServiceGuarantee:      guarantee,
		SubServices:           subServices,
	}
}

func ToCategoryServiceSearchResponse(cs models.CategoryService) CategoryServiceSearchResponse {
	services := make([]ServiceDTO, len(cs.Services))
	for i, s := range cs.Services {
		services[i] = toServiceDTO(s)
	}
	return CategoryServiceSearchResponse{
		ID:              int(cs.ID),
		LayananID:       cs.LayananID,
		CreatorID:       cs.CreatorID,
		CategoryService: cs.CategoryService,
		Services:        services,
	}
}
