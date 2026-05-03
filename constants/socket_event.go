package constants

const (
	REGISTER_CHAT_MITRA_ID    = "register_chat_mitra_id"
	REGISTER_CHAT_CUSTOMER_ID = "register_customer_chat_id"
	REGISTER_CALL_MITRA_ID    = "register_call_mitra_id"
	REGISTER_CALL_CUSTOMER_ID = "register_call_customer_id"

	SEND_REQUEST_CALL_TO_MITRA    = "send_request_call_to_mitra"
	SEND_REQUEST_CALL_TO_CUSTOMER = "send_request_call_to_customer"

	SEND_REQUEST_CANCEL_CALL_TO_MITRA    = "send_request_cancel_call_to_mitra"
	SEND_REQUEST_CANCEL_CALL_TO_CUSTOMER = "send_request_cancel_call_to_customer"

	SEND_REQUEST_END_CALL_TO_MITRA    = "send_request_end_call_to_mitra"
	SEND_REQUEST_END_CALL_TO_CUSTOMER = "send_request_end_call_to_customer"

	SEND_REQUEST_CALL_STATUS_TO_MITRA    = "send_request_call_status_to_mitra"
	SEND_REQUEST_CALL_STATUS_TO_CUSTOMER = "send_request_call_status_to_customer"

	SEND_REQUEST_TO_GET_RESPONSE_FROM_MITRA    = "send_request_to_get_response_from_mitra"
	SEND_REQUEST_TO_GET_RESPONSE_FROM_CUSTOMER = "send_request_to_get_response_from_customer"

	SEND_ACTIVE_RESPONSE_TO_MITRA    = "send_active_response_to_mitra"
	SEND_ACTIVE_RESPONSE_TO_CUSTOMER = "send_active_response_to_customer"

	SOCKET_CALL_CUSTOMER_LISTENER = "socket_call_customer_listener"
	SOCKET_CALL_MITRA_LISTENER    = "socket_call_mitra_listener"

	PICKUP_RESPONSE_FROM_CUSTOMER = "pickup_response_from_customer"
	PICKUP_RESPONSE_FROM_MITRA    = "pickup_response_from_mitra"

	SEND_PICKUP_RESPONSE_TO_MITRA    = "send_pickup_response_to_mitra"
	SEND_PICKUP_RESPONSE_TO_CUSTOMER = "send_pickup_response_to_customer"

	SEND_REQUEST_END_TO_MITRA    = "send_request_end_to_mitra"
	SEND_REQUEST_END_TO_CUSTOMER = "send_request_end_to_customer"

	REGISTERED_CHAT_DATA        = "registered_chat_data"
	UPDATE_MITRA_ONLINE         = "update_mitra_online"
	UPDATE_CUSTOMER_ONLINE      = "update_customer_online"
	REGISTER_SOCKET_STATUS      = "register_socket_status"
	REGISTER_SOCKET_CALL_STATUS = "register_socket_call_status"

	CHAT_RECEIVER       = "chat_receiver"
	CHAT_SEND_TO_SERVER = "chat_send_to_server"
	CHAT_MITRA          = "chat_mitra_listener"
	CHAT_CUSTOMER       = "chat_customer_listener"

	ONLINE_STATUS = "online_status"

	TYPING_LISTENER         = "typing_listener"
	TYPING_STOPPED_LISTENER = "typing_stopped_listener"
	TYPING_EMIT             = "typing_emit"
	TYPING_STOPPED          = "typing_stopped"

	MESSAGE_SEND_STATUS        = "message_send_status"
	UPDATE_MESSAGE_STATUS      = "update_message_status"
	UPDATE_MESSAGE_STATUS_SEND = "update_message_status_send"

	INITIATE_SOCKET_ADMIN = "INITIATE_SOCKET_ADMIN"
	REGISTER_SOCKET_ADMIN = "REGISTER_SOCKET_ADMIN"
	RESPONSE_SOCKET_ADMIN = "RESPONSE_SOCKET_ADMIN"
	MESSAGE_SOCKET_ADMIN  = "MESSAGE_SOCKET_ADMIN"

	NOTIFICATION_ORDER_RUNNING              = "NOTIFICATION_ORDER_RUNNING"
	NOTIFICATION_ORDER_FINISH               = "NOTIFICATION_ORDER_FINISH"
	NOTIFICATION_ORDER_CANCELED             = "NOTIFICATION_ORDER_CANCELED"
	NOTIFICATION_ORDER_WAITING_FOR_SELECTED = "NOTIFICATION_ORDER_WAITING_FOR_SELECTED"
	NOTIFICATION_REJECTED_MITRA             = "REJECTED_MITRA"

	// FCM notification_type values for order broadcasts
	ORDER_BROADCAST = "ORDER_BROADCAST"
	ORDER_AUTO_BID  = "ORDER_AUTO_BID"

	REGISTER_COORDINATE_UPDATE = "REGISTER_COORDINATE_UPDATE"
	REGISTER_COORDINATE        = "REGISTER_COORDINATE"
	REGISTER_COORDINATE_STATUS = "REGISTER_COORDINATE_STATUS"
	COORDINATE_UPDATE          = "COORDINATE_UPDATE"
)
