// Package i18n menyediakan sistem internasionalisasi untuk response HTTP.
// Gunakan T(lang, key) untuk mengambil teks yang sudah diterjemahkan.
// Bahasa ditentukan dari header "device_language" (ID atau EN, default ID).
package i18n

// Konstanta kunci pesan yang digunakan di seluruh aplikasi.
// Nama konstanta menggunakan prefix "Msg" untuk HTTP response,
// dan "Notif" untuk judul/isi push notification FCM.
const (
	// ── Generic ──────────────────────────────────────────────────────────────
	MsgSuccess            = "success"
	MsgOK                 = "ok"
	MsgBadRequest         = "bad_request"
	MsgInternalError      = "internal_error"
	MsgNotFound           = "not_found"
	MsgInvalidID          = "invalid_id"
	MsgInvalidPayload     = "invalid_payload"
	MsgInvalidRequest     = "invalid_request"
	MsgMissingParams      = "missing_params"
	MsgUnauthorized       = "unauthorized"
	MsgTooManyRequests    = "too_many_requests"
	MsgRequestBodyLarge   = "request_body_large"
	MsgEventNotHandled    = "event_not_handled"
	MsgInvalidCallbackTok = "invalid_callback_token"

	// ── Auth / Token ─────────────────────────────────────────────────────────
	MsgTokenExpired              = "token_expired"
	MsgTokenInvalidFormat        = "token_invalid_format"
	MsgTokenInvalidOrExpired     = "token_invalid_or_expired"
	MsgTokenInvalidClaims        = "token_invalid_claims"
	MsgTokenInvalidPayload       = "token_invalid_payload"
	MsgServerConfigError         = "server_config_error"
	MsgAccountNotActive          = "account_not_active"
	MsgDeviceMismatch            = "device_mismatch"
	MsgMissingRefreshToken       = "missing_refresh_token"
	MsgRefreshTokenInvalid       = "refresh_token_invalid"
	MsgRefreshTokenExpired       = "refresh_token_expired"
	MsgRefreshTokenNotRecognized = "refresh_token_not_recognized"
	MsgRefreshTokenUpdated       = "refresh_token_updated"

	// ── User / Auth ───────────────────────────────────────────────────────────
	MsgLoginSuccess           = "login_success"
	MsgLogoutSuccess          = "logout_success"
	MsgRegisterSuccess        = "register_success"
	MsgPasswordWrong          = "password_wrong"
	MsgPasswordUpdated        = "password_updated"
	MsgPasswordChanged        = "password_changed"
	MsgPasswordNotMatch       = "password_not_match"
	MsgProfileUpdated         = "profile_updated"
	MsgEmailChanged           = "email_changed"
	MsgEmailPasswordUpdated   = "email_password_updated"
	MsgEmailNotRegistered     = "email_not_registered"
	MsgPhoneChanged           = "phone_changed"
	MsgPhoneAlreadyUsed       = "phone_already_used"
	MsgOtpSent                = "otp_sent"
	MsgOtpValid               = "otp_valid"
	MsgOtpWrong               = "otp_wrong"
	MsgForgotPasswordSuccess  = "forgot_password_success"
	MsgChangeDataSuccess      = "change_data_success"
	MsgRemoveAccountSuccess   = "remove_account_success"
	MsgFirebaseTokenUpdated   = "firebase_token_updated"
	MsgAccountAlreadyLoggedIn = "account_already_logged_in"

	// ── PIN ───────────────────────────────────────────────────────────────────
	MsgPinCheckSuccess = "pin_check_success"
	MsgOldPinDifferent = "old_pin_different"

	// ── Admin ─────────────────────────────────────────────────────────────────
	MsgAdminCreated       = "admin_created"
	MsgAdminNotFound      = "admin_not_found"
	MsgAdminStatusUpdated = "admin_status_updated"

	// ── Customer ─────────────────────────────────────────────────────────────
	MsgCustomerNotFound = "customer_not_found"

	// ── Mitra ────────────────────────────────────────────────────────────────
	MsgMitraNotFound         = "mitra_not_found"
	MsgMitraDataNotFound     = "mitra_data_not_found"
	MsgMitraUpdated          = "mitra_updated"
	MsgMitraStatusUpdated    = "mitra_status_updated"
	MsgMitraActiveUpdated    = "mitra_active_updated"
	MsgMitraAutoBidUpdated   = "mitra_auto_bid_updated"
	MsgMitraCoordUpdated     = "mitra_coord_updated"
	MsgMitraCandidateUpdated = "mitra_candidate_updated"
	MsgMitraInvited          = "mitra_invited"
	MsgRejectionCountUpdated = "rejection_count_updated"
	MsgDocumentStatusUpdated = "document_status_updated"
	MsgMitraActivated        = "mitra_activated"
	MsgMitraRejected         = "mitra_rejected"

	// ── Order ─────────────────────────────────────────────────────────────────
	MsgOrderCreated        = "order_created"
	MsgOrderTookSuccess    = "order_took_success"
	MsgTransactionNotFound = "transaction_not_found"

	// ── Payment / Disbursement ────────────────────────────────────────────────
	MsgTopUpCreated           = "top_up_created"
	MsgDisbursementCreated    = "disbursement_created"
	MsgDisbursementNotAllowed = "disbursement_not_allowed"
	MsgBankNotFound           = "bank_not_found"
	MsgBankDataNotFound       = "bank_data_not_found"
	MsgBankOrEwalletNotFound  = "bank_or_ewallet_not_found"
	MsgAmountInsufficient     = "amount_insufficient"
	MsgPaymentCreated         = "payment_created"
	MsgPaymentDeleted         = "payment_deleted"
	MsgPaymentUpdated         = "payment_updated"
	MsgBankListCreated        = "bank_list_created"
	MsgEwalletCreated         = "ewallet_created"
	MsgBankEwalletUpdated     = "bank_ewallet_updated"
	MsgSubPaymentUpdated      = "sub_payment_updated"

	// ── Complain ─────────────────────────────────────────────────────────────
	MsgComplainCreated  = "complain_created"
	MsgComplainNotFound = "complain_not_found"
	MsgComplainUpdated  = "complain_updated"
	MsgComplainRemoved  = "complain_removed"

	// ── Category / Service / Sub-service ─────────────────────────────────────
	MsgCategoryCreated             = "category_created"
	MsgCategoryUpdated             = "category_updated"
	MsgCategoryDeleted             = "category_deleted"
	MsgServiceCreated              = "service_created"
	MsgServiceUpdated              = "service_updated"
	MsgServiceUpdatedWithImage     = "service_updated_with_image"
	MsgServiceDeleted              = "service_deleted"
	MsgSubServiceCreated           = "sub_service_created"
	MsgSubServiceUpdated           = "sub_service_updated"
	MsgSubServiceRemoved           = "sub_service_removed"
	MsgSubServiceAdditionalCreated = "sub_service_additional_created"
	MsgSubServiceAdditionalUpdated = "sub_service_additional_updated"
	MsgSubServiceAdditionalRemoved = "sub_service_additional_removed"
	MsgLayananCreated              = "layanan_created"
	MsgLayananUpdated              = "layanan_updated"
	MsgLayananRemoved              = "layanan_removed"
	MsgInvalidLayananID            = "invalid_layanan_id"

	// ── Banner / Content ─────────────────────────────────────────────────────
	MsgBannerCreated        = "banner_created"
	MsgBannerUpdated        = "banner_updated"
	MsgBannerRemoved        = "banner_removed"
	MsgNewsCreated          = "news_created"
	MsgNewsUpdated          = "news_updated"
	MsgNewsRemoved          = "news_removed"
	MsgBantuanCreated       = "bantuan_created"
	MsgBantuanUpdated       = "bantuan_updated"
	MsgBantuanRemoved       = "bantuan_removed"
	MsgPanduanCreated       = "panduan_created"
	MsgPanduanNotFound      = "panduan_not_found"
	MsgPanduanUpdated       = "panduan_updated"
	MsgPanduanRemoved       = "panduan_removed"
	MsgWatchingCountUpdated = "watching_count_updated"

	// ── Terms / TOC ───────────────────────────────────────────────────────────
	MsgTocCreated              = "toc_created"
	MsgTocUpdated              = "toc_updated"
	MsgTocDeleted              = "toc_deleted"
	MsgTermsNotFound           = "terms_not_found"
	MsgTermsUpdated            = "terms_updated"
	MsgTermsCreateFailed       = "terms_create_failed"
	MsgTermsDeleteFailed       = "terms_delete_failed"
	MsgTermsUpdateFailed       = "terms_update_failed"
	MsgTermsUpdateStatusFailed = "terms_update_status_failed"

	// ── Rating ────────────────────────────────────────────────────────────────
	MsgRatingSubmitted = "rating_submitted"
	MsgRatingInvalid   = "rating_invalid"

	// ── Schedule ─────────────────────────────────────────────────────────────
	MsgScheduleCreated = "schedule_created"
	MsgScheduleUpdated = "schedule_updated"
	MsgScheduleRemoved = "schedule_removed"
)
