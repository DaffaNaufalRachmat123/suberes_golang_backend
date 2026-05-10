package i18n

// enTranslations is the English (US) translation map.
var enTranslations = map[string]string{
	// ── Generic ──────────────────────────────────────────────────────────────
	MsgSuccess:            "Success.",
	MsgOK:                 "OK.",
	MsgBadRequest:         "Invalid request.",
	MsgInternalError:      "An internal server error occurred.",
	MsgNotFound:           "Data not found.",
	MsgInvalidID:          "The provided ID is invalid.",
	MsgInvalidPayload:     "Invalid request payload.",
	MsgInvalidRequest:     "Invalid request.",
	MsgMissingParams:      "Required parameters are missing.",
	MsgUnauthorized:       "Unauthorized access.",
	MsgTooManyRequests:    "Too many requests. Please try again later.",
	MsgRequestBodyLarge:   "Request body is too large.",
	MsgEventNotHandled:    "Event type not handled.",
	MsgInvalidCallbackTok: "Invalid callback token.",

	// ── Auth / Token ─────────────────────────────────────────────────────────
	MsgTokenExpired:              "Your session has expired. Please log in again.",
	MsgTokenInvalidFormat:        "Invalid token format.",
	MsgTokenInvalidOrExpired:     "Invalid or expired token.",
	MsgTokenInvalidClaims:        "Invalid token claims.",
	MsgTokenInvalidPayload:       "Invalid token payload.",
	MsgServerConfigError:         "Server configuration error.",
	MsgAccountNotActive:          "Account is not active or authorized.",
	MsgDeviceMismatch:            "Device mismatch detected. Please log in again.",
	MsgMissingRefreshToken:       "Refresh token is missing.",
	MsgRefreshTokenInvalid:       "Invalid refresh token.",
	MsgRefreshTokenExpired:       "Refresh token has expired.",
	MsgRefreshTokenNotRecognized: "Refresh token not recognized.",
	MsgRefreshTokenUpdated:       "Refresh token updated successfully.",

	// ── User / Auth ───────────────────────────────────────────────────────────
	MsgLoginSuccess:           "Logged in successfully. Welcome back!",
	MsgLogoutSuccess:          "Logged out successfully. See you next time!",
	MsgRegisterSuccess:        "Registration successful. Welcome aboard!",
	MsgPasswordWrong:          "Incorrect password.",
	MsgPasswordUpdated:        "Password updated successfully.",
	MsgPasswordChanged:        "Password changed successfully.",
	MsgPasswordNotMatch:       "Passwords do not match.",
	MsgProfileUpdated:         "Profile updated successfully.",
	MsgEmailChanged:           "Email address updated successfully.",
	MsgEmailPasswordUpdated:   "Email and password updated successfully.",
	MsgEmailNotRegistered:     "This email address is not registered.",
	MsgPhoneChanged:           "Phone number updated successfully.",
	MsgPhoneAlreadyUsed:       "This phone number is already in use by another account.",
	MsgOtpSent:                "OTP code sent successfully.",
	MsgOtpValid:               "OTP is valid.",
	MsgOtpWrong:               "Incorrect OTP code.",
	MsgForgotPasswordSuccess:  "Password reset request submitted successfully.",
	MsgChangeDataSuccess:      "Data updated successfully.",
	MsgRemoveAccountSuccess:   "Account removed successfully.",
	MsgFirebaseTokenUpdated:   "Firebase token updated successfully.",
	MsgAccountAlreadyLoggedIn: "This account is already logged in on another device.",

	// ── PIN ───────────────────────────────────────────────────────────────────
	MsgPinCheckSuccess: "PIN verified successfully.",
	MsgOldPinDifferent: "The old PIN you entered is incorrect.",

	// ── Admin ─────────────────────────────────────────────────────────────────
	MsgAdminCreated:       "Admin created successfully.",
	MsgAdminNotFound:      "Admin not found.",
	MsgAdminStatusUpdated: "Admin status updated successfully.",

	// ── Customer ─────────────────────────────────────────────────────────────
	MsgCustomerNotFound: "Customer not found.",

	// ── Mitra ────────────────────────────────────────────────────────────────
	MsgMitraNotFound:         "Mitra not found.",
	MsgMitraDataNotFound:     "Mitra data not found.",
	MsgMitraUpdated:          "Mitra updated successfully.",
	MsgMitraStatusUpdated:    "Mitra status updated successfully.",
	MsgMitraActiveUpdated:    "Mitra active status updated successfully.",
	MsgMitraAutoBidUpdated:   "Mitra auto-bid status updated successfully.",
	MsgMitraCoordUpdated:     "Mitra coordinates updated successfully.",
	MsgMitraCandidateUpdated: "Mitra candidate data updated successfully.",
	MsgMitraInvited:          "Mitra invited successfully.",
	MsgRejectionCountUpdated: "Rejection count updated successfully.",
	MsgDocumentStatusUpdated: "Document status updated successfully.",
	MsgMitraActivated:        "Mitra account activated successfully.",
	MsgMitraRejected:         "Mitra account has been rejected.",

	// ── Order ─────────────────────────────────────────────────────────────────
	MsgOrderCreated:        "Order created successfully.",
	MsgOrderTookSuccess:    "Order accepted successfully.",
	MsgTransactionNotFound: "Transaction not found.",

	// ── Payment / Disbursement ────────────────────────────────────────────────
	MsgTopUpCreated:           "Top-up created successfully.",
	MsgDisbursementCreated:    "Disbursement created successfully.",
	MsgDisbursementNotAllowed: "This disbursement method is not allowed.",
	MsgBankNotFound:           "Bank not found.",
	MsgBankDataNotFound:       "Bank data not found.",
	MsgBankOrEwalletNotFound:  "Bank or e-wallet data not found.",
	MsgAmountInsufficient:     "Insufficient balance.",
	MsgPaymentCreated:         "Payment created successfully.",
	MsgPaymentDeleted:         "Payment deleted successfully.",
	MsgPaymentUpdated:         "Payment updated successfully.",
	MsgBankListCreated:        "Bank list created successfully.",
	MsgEwalletCreated:         "E-wallet created successfully.",
	MsgBankEwalletUpdated:     "Bank/e-wallet updated successfully.",
	MsgSubPaymentUpdated:      "Sub-payment updated successfully.",

	// ── Complain ─────────────────────────────────────────────────────────────
	MsgComplainCreated:  "Complaint submitted successfully.",
	MsgComplainNotFound: "Complaint not found.",
	MsgComplainUpdated:  "Complaint updated successfully.",
	MsgComplainRemoved:  "Complaint removed successfully.",

	// ── Category / Service / Sub-service ─────────────────────────────────────
	MsgCategoryCreated:             "Category created successfully.",
	MsgCategoryUpdated:             "Category updated successfully.",
	MsgCategoryDeleted:             "Category deleted successfully.",
	MsgServiceCreated:              "Service created successfully.",
	MsgServiceUpdated:              "Service updated successfully.",
	MsgServiceUpdatedWithImage:     "Service updated with image successfully.",
	MsgServiceDeleted:              "Service deleted successfully.",
	MsgSubServiceCreated:           "Sub-service created successfully.",
	MsgSubServiceUpdated:           "Sub-service updated successfully.",
	MsgSubServiceRemoved:           "Sub-service removed successfully.",
	MsgSubServiceAdditionalCreated: "Sub-service additional created successfully.",
	MsgSubServiceAdditionalUpdated: "Sub-service additional updated successfully.",
	MsgSubServiceAdditionalRemoved: "Sub-service additional removed successfully.",
	MsgLayananCreated:              "Service created successfully.",
	MsgLayananUpdated:              "Service updated successfully.",
	MsgLayananRemoved:              "Service removed successfully.",
	MsgInvalidLayananID:            "Invalid service ID.",

	// ── Banner / Content ─────────────────────────────────────────────────────
	MsgBannerCreated:        "Banner created successfully.",
	MsgBannerUpdated:        "Banner updated successfully.",
	MsgBannerRemoved:        "Banner removed successfully.",
	MsgNewsCreated:          "News created successfully.",
	MsgNewsUpdated:          "News updated successfully.",
	MsgNewsRemoved:          "News removed successfully.",
	MsgBantuanCreated:       "Help item created successfully.",
	MsgBantuanUpdated:       "Help item updated successfully.",
	MsgBantuanRemoved:       "Help item removed successfully.",
	MsgPanduanCreated:       "Guide created successfully.",
	MsgPanduanNotFound:      "Guide not found.",
	MsgPanduanUpdated:       "Guide updated successfully.",
	MsgPanduanRemoved:       "Guide removed successfully.",
	MsgWatchingCountUpdated: "Watch count updated successfully.",

	// ── Terms / TOC ───────────────────────────────────────────────────────────
	MsgTocCreated:              "Terms of condition created successfully.",
	MsgTocUpdated:              "Terms of condition updated successfully.",
	MsgTocDeleted:              "Terms of condition deleted successfully.",
	MsgTermsNotFound:           "Terms and conditions not found.",
	MsgTermsUpdated:            "Terms and conditions updated successfully.",
	MsgTermsCreateFailed:       "Failed to create terms and conditions.",
	MsgTermsDeleteFailed:       "Failed to delete terms and conditions.",
	MsgTermsUpdateFailed:       "Failed to update terms and conditions.",
	MsgTermsUpdateStatusFailed: "Failed to update terms and conditions status.",

	// ── Rating ────────────────────────────────────────────────────────────────
	MsgRatingSubmitted: "Rating submitted successfully.",
	MsgRatingInvalid:   "Invalid rating value.",

	// ── Schedule ─────────────────────────────────────────────────────────────
	MsgScheduleCreated: "Schedule created successfully.",
	MsgScheduleUpdated: "Schedule updated successfully.",
	MsgScheduleRemoved: "Schedule removed successfully.",
}
