package dtos

import "suberes_golang/models"

type MitraLoginResponseDTO struct {
	ServerMessage string      `json:"server_message"`
	Status        string      `json:"status"`
	Token         string      `json:"token"`
	Data          models.User `json:"data"`
	SharedPrime   int         `json:"shared_prime"`
	SharedSecret  int64       `json:"shared_secret"`
}
