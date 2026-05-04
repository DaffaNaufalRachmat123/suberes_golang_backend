package queue

import "encoding/json"

const (
	TypeOrderQueueCash              = "order:cash"
	TypeOrderQueueVA                = "order:va"
	TypeOrderOfferExpired           = "order:offer_expired"
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
