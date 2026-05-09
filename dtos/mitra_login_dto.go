package dtos

type MitraLoginDTO struct {
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required"`
	DeviceID        string `json:"device_id"`
	DeviceName      string `json:"device_name"`
	DeviceOS        string `json:"device_os"`
	DeviceOSAndroid string `json:"device_os_android"`
}
