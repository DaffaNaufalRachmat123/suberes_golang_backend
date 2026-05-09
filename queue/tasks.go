package queue

import (
	"encoding/json"
	"time"
)

const (
	TypeOrderQueueCash              = "order:cash"
	TypeOrderQueueVA                = "order:va"
	TypeOrderOfferExpired           = "order:offer_expired"
	TypeOrderOfferMitraExpired      = "order:offer_mitra_expired"
	TypeOrderSelectedExpired        = "order:selected_expired"
	TypeOrderOnProgressToFinish     = "order:on_progress_to_finish"
	TypeOrderEwalletNotifyExpired   = "order:ewallet_notify_expired"
	TypeOrderVAEwalletNotifyExpired = "order:va_ewallet_notify_expired"
	TypeOrderComingSoonRun          = "order:coming_soon_run"
	TypeOrderComingSoonWarning      = "order:coming_soon_warning"
)

type OrderEwalletNotifyExpiredPayload struct {
	OrderID    string `json:"order_id"`
	CustomerID string `json:"customer_id"`
}

func NewOrderEwalletNotifyExpiredTask(orderID, customerID string) ([]byte, error) {
	return json.Marshal(OrderEwalletNotifyExpiredPayload{
		OrderID:    orderID,
		CustomerID: customerID,
	})
}

type OrderComingSoonRunPayload struct {
	OrderID string `json:"order_id"`
}

func NewOrderComingSoonRunTask(orderID string) ([]byte, error) {
	return json.Marshal(OrderComingSoonRunPayload{OrderID: orderID})
}

type OrderComingSoonWarningPayload struct {
	OrderID string `json:"order_id"`
}

func NewOrderComingSoonWarningTask(orderID string) ([]byte, error) {
	return json.Marshal(OrderComingSoonWarningPayload{OrderID: orderID})
}

type OrderOnProgressToFinishPayload struct {
	ID           string `json:"id"`
	CustomerID   string `json:"customer_id"`
	MitraID      string `json:"mitra_id"`
	ServiceID    int    `json:"service_id"`
	SubServiceID int    `json:"sub_service_id"`
}

func NewOrderOnProgressToFinishTask(id, customerID, mitraID string, serviceID, subServiceID int) ([]byte, error) {
	payload := OrderOnProgressToFinishPayload{
		ID:           id,
		CustomerID:   customerID,
		MitraID:      mitraID,
		ServiceID:    serviceID,
		SubServiceID: subServiceID,
	}
	return json.Marshal(payload)
}

type OrderQueueCashPayload struct {
	OrderID         string `json:"order_id"`
	CustomerID      string `json:"customer_id"`
	IsWithTime      bool   `json:"is_with_time"`     // serviceData.ServiceType == "Durasi"
	ServiceDuration int    `json:"service_duration"` // subServiceData.MinutesSubServices
	SearchAttempt   int    `json:"search_attempt"`   // how many times the search has been tried (0 = first)
	FirstEnqueued   int64  `json:"first_enqueued"`   // unix timestamp of the first enqueue time
}

func NewOrderQueueCashTask(orderID, customerID string, isWithTime bool, serviceDuration int) ([]byte, error) {
	payload := OrderQueueCashPayload{
		OrderID:         orderID,
		CustomerID:      customerID,
		IsWithTime:      isWithTime,
		ServiceDuration: serviceDuration,
		SearchAttempt:   0,
		FirstEnqueued:   time.Now().Unix(),
	}
	return json.Marshal(payload)
}

type OrderQueueVAPayload struct {
	OrderID         string `json:"order_id"`
	IsWithTime      bool   `json:"is_with_time"`     // serviceData.ServiceType == "Durasi"
	ServiceDuration int    `json:"service_duration"` // subServiceData.MinutesSubServices
	FirstEnqueued   int64  `json:"first_enqueued"`   // unix timestamp of the first enqueue time
}

func NewOrderQueueVATask(orderID string, isWithTime bool, serviceDuration int) ([]byte, error) {
	payload := OrderQueueVAPayload{
		OrderID:         orderID,
		IsWithTime:      isWithTime,
		ServiceDuration: serviceDuration,
		FirstEnqueued:   time.Now().Unix(),
	}
	return json.Marshal(payload)
}

type OrderOfferExpiredPayload struct {
	OrderID                  string `json:"order_id"`
	TempID                   string `json:"temp_id"`
	CustomerID               string `json:"customer_id"`
	NotificationID           int    `json:"notification_id"`
	MinuteDifferenceSelected string `json:"minute_difference_selected"`
}

func NewOrderOfferExpiredTask(orderID, tempID, customerID string, notificationID int, minuteDifferenceSelected string) ([]byte, error) {
	payload := OrderOfferExpiredPayload{
		OrderID:                  orderID,
		TempID:                   tempID,
		CustomerID:               customerID,
		NotificationID:           notificationID,
		MinuteDifferenceSelected: minuteDifferenceSelected,
	}
	return json.Marshal(payload)
}

type OrderSelectedExpiredPayload struct {
	OrderID string `json:"order_id"`
}

func NewOrderSelectedExpiredTask(orderID string) ([]byte, error) {
	payload := OrderSelectedExpiredPayload{
		OrderID: orderID,
	}
	return json.Marshal(payload)
}

type OrderOfferMitraExpiredPayload struct {
	OrderID string `json:"order_id"`
	MitraID string `json:"mitra_id"`
	TempID  string `json:"temp_id"`
}

func NewOrderOfferMitraExpiredTask(orderID, mitraID, tempID string) ([]byte, error) {
	return json.Marshal(OrderOfferMitraExpiredPayload{
		OrderID: orderID,
		MitraID: mitraID,
		TempID:  tempID,
	})
}
