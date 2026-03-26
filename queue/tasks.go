package queue

import "encoding/json"

const (
	TypeOrderQueueCash       = "order:cash"
	TypeOrderQueueVA         = "order:va"
	TypeOrderOfferExpired    = "order:offer_expired"
	TypeOrderSelectedExpired = "order:selected_expired"
)

type OrderQueueCashPayload struct {
	OrderID    string `json:"order_id"`
	CustomerID string `json:"customer_id"`
}

func NewOrderQueueCashTask(orderID, customerID string) ([]byte, error) {
	payload := OrderQueueCashPayload{
		OrderID:    orderID,
		CustomerID: customerID,
	}
	return json.Marshal(payload)
}

type OrderQueueVAPayload struct {
	OrderID string `json:"order_id"`
}

func NewOrderQueueVATask(orderID string) ([]byte, error) {
	payload := OrderQueueVAPayload{
		OrderID: orderID,
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
