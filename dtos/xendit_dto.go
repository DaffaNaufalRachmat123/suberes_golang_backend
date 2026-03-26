package dtos

// XenditEwalletData defines the structure for the 'data' field in a Xendit callback.
type XenditEwalletData struct {
	ID          string `json:"id"`
	ReferenceID string `json:"reference_id"`
	Status      string `json:"status"`
	VoidStatus  string `json:"void_status"`
}

// XenditCallbackPayload is the top-level structure for a Xendit callback.
type XenditCallbackPayload struct {
	Event string            `json:"event"`
	Data  XenditEwalletData `json:"data"`
}
