package dtos

import "suberes_golang/models"

type MitraProfileResponseDTO struct {
	Profile    *models.User `json:"profile"`
	OrderCount struct {
		OrderCount       int `json:"order_count"`
		PendapatanOrder int `json:"pendapatan_order"`
	} `json:"order_count"`
	BillData int `json:"bill_data"`
}
