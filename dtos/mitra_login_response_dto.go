package dtos

import "suberes_golang/models"

type UserLoginResponseDTO struct {
	ServerMessage string      `json:"server_message"`
	Status        string      `json:"status"`
	Token         string      `json:"token"`
	RefreshToken  string      `json:"refresh_token"`
	Data          models.User `json:"data"`
	SharedPrime   int64       `json:"shared_prime"`
	SharedSecret  int64       `json:"shared_secret"`
}
