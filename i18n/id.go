package i18n

// idTranslations adalah terjemahan Bahasa Indonesia (gaul & proper).
var idTranslations = map[string]string{
	// ── Generic ──────────────────────────────────────────────────────────────
	MsgSuccess:            "Berhasil.",
	MsgOK:                 "OK.",
	MsgBadRequest:         "Permintaan tidak valid.",
	MsgInternalError:      "Terjadi kesalahan pada server.",
	MsgNotFound:           "Data tidak ditemukan.",
	MsgInvalidID:          "ID yang diberikan tidak valid.",
	MsgInvalidPayload:     "Data yang dikirim tidak valid.",
	MsgInvalidRequest:     "Permintaan tidak valid.",
	MsgMissingParams:      "Parameter yang diperlukan tidak lengkap.",
	MsgUnauthorized:       "Akses tidak diizinkan.",
	MsgTooManyRequests:    "Terlalu banyak permintaan, coba lagi sebentar lagi.",
	MsgRequestBodyLarge:   "Ukuran data yang dikirim terlalu besar.",
	MsgEventNotHandled:    "Tipe event tidak ditangani.",
	MsgInvalidCallbackTok: "Token callback tidak valid.",

	// ── Auth / Token ─────────────────────────────────────────────────────────
	MsgTokenExpired:              "Sesi kamu sudah kedaluwarsa. Silakan login kembali.",
	MsgTokenInvalidFormat:        "Format token tidak sesuai.",
	MsgTokenInvalidOrExpired:     "Token tidak valid atau sudah kedaluwarsa.",
	MsgTokenInvalidClaims:        "Informasi di dalam token tidak valid.",
	MsgTokenInvalidPayload:       "Payload token tidak valid.",
	MsgServerConfigError:         "Terjadi kesalahan konfigurasi server.",
	MsgAccountNotActive:          "Akun tidak aktif atau tidak terotorisasi.",
	MsgDeviceMismatch:            "Perangkat berbeda terdeteksi. Silakan login ulang.",
	MsgMissingRefreshToken:       "Refresh token tidak ditemukan.",
	MsgRefreshTokenInvalid:       "Refresh token tidak valid.",
	MsgRefreshTokenExpired:       "Refresh token sudah kedaluwarsa.",
	MsgRefreshTokenNotRecognized: "Refresh token tidak dikenali.",
	MsgRefreshTokenUpdated:       "Refresh token berhasil diperbarui.",

	// ── User / Auth ───────────────────────────────────────────────────────────
	MsgLoginSuccess:           "Login berhasil! Selamat datang kembali.",
	MsgLogoutSuccess:          "Logout berhasil. Sampai jumpa lagi!",
	MsgRegisterSuccess:        "Pendaftaran berhasil! Selamat bergabung.",
	MsgPasswordWrong:          "Password yang kamu masukkan salah.",
	MsgPasswordUpdated:        "Password berhasil diperbarui.",
	MsgPasswordChanged:        "Password berhasil diubah.",
	MsgPasswordNotMatch:       "Password tidak cocok.",
	MsgProfileUpdated:         "Profil kamu berhasil diperbarui.",
	MsgEmailChanged:           "Email berhasil diubah.",
	MsgEmailPasswordUpdated:   "Email dan password berhasil diperbarui.",
	MsgEmailNotRegistered:     "Email ini belum terdaftar.",
	MsgPhoneChanged:           "Nomor telepon berhasil diubah.",
	MsgPhoneAlreadyUsed:       "Nomor telepon ini sudah digunakan oleh akun lain.",
	MsgOtpSent:                "Kode OTP berhasil dikirim.",
	MsgOtpValid:               "OTP valid.",
	MsgOtpWrong:               "Kode OTP yang kamu masukkan salah.",
	MsgForgotPasswordSuccess:  "Permintaan reset password berhasil dikirim.",
	MsgChangeDataSuccess:      "Data berhasil diperbarui.",
	MsgRemoveAccountSuccess:   "Akun berhasil dihapus.",
	MsgFirebaseTokenUpdated:   "Firebase token berhasil diperbarui.",
	MsgAccountAlreadyLoggedIn: "Akun ini sudah aktif di perangkat lain.",

	// ── PIN ───────────────────────────────────────────────────────────────────
	MsgPinCheckSuccess: "PIN berhasil diverifikasi.",
	MsgOldPinDifferent: "PIN lama yang kamu masukkan tidak sesuai.",

	// ── Admin ─────────────────────────────────────────────────────────────────
	MsgAdminCreated:       "Admin berhasil dibuat.",
	MsgAdminNotFound:      "Admin tidak ditemukan.",
	MsgAdminStatusUpdated: "Status admin berhasil diperbarui.",

	// ── Customer ─────────────────────────────────────────────────────────────
	MsgCustomerNotFound: "Data customer tidak ditemukan.",

	// ── Mitra ────────────────────────────────────────────────────────────────
	MsgMitraNotFound:         "Mitra tidak ditemukan.",
	MsgMitraDataNotFound:     "Data mitra tidak ditemukan.",
	MsgMitraUpdated:          "Data mitra berhasil diperbarui.",
	MsgMitraStatusUpdated:    "Status mitra berhasil diperbarui.",
	MsgMitraActiveUpdated:    "Status aktif mitra berhasil diperbarui.",
	MsgMitraAutoBidUpdated:   "Status auto-bid mitra berhasil diperbarui.",
	MsgMitraCoordUpdated:     "Koordinat mitra berhasil diperbarui.",
	MsgMitraCandidateUpdated: "Data kandidat mitra berhasil diperbarui.",
	MsgMitraInvited:          "Mitra berhasil diundang.",
	MsgRejectionCountUpdated: "Jumlah penolakan berhasil diperbarui.",
	MsgDocumentStatusUpdated: "Status dokumen berhasil diperbarui.",
	MsgMitraActivated:        "Akun mitra berhasil diaktifkan.",
	MsgMitraRejected:         "Akun mitra telah ditolak.",

	// ── Order ─────────────────────────────────────────────────────────────────
	MsgOrderCreated:        "Order berhasil dibuat!",
	MsgOrderTookSuccess:    "Kamu berhasil mengambil orderan ini!",
	MsgTransactionNotFound: "Transaksi tidak ditemukan.",

	// ── Payment / Disbursement ────────────────────────────────────────────────
	MsgTopUpCreated:           "Top up berhasil dibuat.",
	MsgDisbursementCreated:    "Pencairan berhasil dibuat.",
	MsgDisbursementNotAllowed: "Metode pencairan ini tidak diizinkan.",
	MsgBankNotFound:           "Data bank tidak ditemukan.",
	MsgBankDataNotFound:       "Data bank tidak ditemukan.",
	MsgBankOrEwalletNotFound:  "Data bank atau e-wallet tidak ditemukan.",
	MsgAmountInsufficient:     "Jumlah saldo tidak mencukupi.",
	MsgPaymentCreated:         "Pembayaran berhasil dibuat.",
	MsgPaymentDeleted:         "Pembayaran berhasil dihapus.",
	MsgPaymentUpdated:         "Pembayaran berhasil diperbarui.",
	MsgBankListCreated:        "Daftar bank berhasil ditambahkan.",
	MsgEwalletCreated:         "E-wallet berhasil ditambahkan.",
	MsgBankEwalletUpdated:     "Data bank/e-wallet berhasil diperbarui.",
	MsgSubPaymentUpdated:      "Sub-pembayaran berhasil diperbarui.",

	// ── Complain ─────────────────────────────────────────────────────────────
	MsgComplainCreated:  "Komplain kamu berhasil dikirim.",
	MsgComplainNotFound: "Komplain tidak ditemukan.",
	MsgComplainUpdated:  "Komplain berhasil diperbarui.",
	MsgComplainRemoved:  "Komplain berhasil dihapus.",

	// ── Category / Service / Sub-service ─────────────────────────────────────
	MsgCategoryCreated:             "Kategori layanan berhasil dibuat.",
	MsgCategoryUpdated:             "Kategori layanan berhasil diperbarui.",
	MsgCategoryDeleted:             "Kategori layanan berhasil dihapus.",
	MsgServiceCreated:              "Layanan berhasil dibuat.",
	MsgServiceUpdated:              "Layanan berhasil diperbarui.",
	MsgServiceUpdatedWithImage:     "Layanan beserta gambar berhasil diperbarui.",
	MsgServiceDeleted:              "Layanan berhasil dihapus.",
	MsgSubServiceCreated:           "Sub-layanan berhasil dibuat.",
	MsgSubServiceUpdated:           "Sub-layanan berhasil diperbarui.",
	MsgSubServiceRemoved:           "Sub-layanan berhasil dihapus.",
	MsgSubServiceAdditionalCreated: "Tambahan sub-layanan berhasil dibuat.",
	MsgSubServiceAdditionalUpdated: "Tambahan sub-layanan berhasil diperbarui.",
	MsgSubServiceAdditionalRemoved: "Tambahan sub-layanan berhasil dihapus.",
	MsgLayananCreated:              "Layanan berhasil dibuat.",
	MsgLayananUpdated:              "Layanan berhasil diperbarui.",
	MsgLayananRemoved:              "Layanan berhasil dihapus.",
	MsgInvalidLayananID:            "ID layanan tidak valid.",

	// ── Banner / Content ─────────────────────────────────────────────────────
	MsgBannerCreated:        "Banner berhasil ditambahkan.",
	MsgBannerUpdated:        "Banner berhasil diperbarui.",
	MsgBannerRemoved:        "Banner berhasil dihapus.",
	MsgNewsCreated:          "Berita berhasil ditambahkan.",
	MsgNewsUpdated:          "Berita berhasil diperbarui.",
	MsgNewsRemoved:          "Berita berhasil dihapus.",
	MsgBantuanCreated:       "Bantuan berhasil ditambahkan.",
	MsgBantuanUpdated:       "Bantuan berhasil diperbarui.",
	MsgBantuanRemoved:       "Bantuan berhasil dihapus.",
	MsgPanduanCreated:       "Panduan berhasil ditambahkan.",
	MsgPanduanNotFound:      "Panduan tidak ditemukan.",
	MsgPanduanUpdated:       "Panduan berhasil diperbarui.",
	MsgPanduanRemoved:       "Panduan berhasil dihapus.",
	MsgWatchingCountUpdated: "Jumlah penonton berhasil diperbarui.",

	// ── Terms / TOC ───────────────────────────────────────────────────────────
	MsgTocCreated:              "TOC berhasil dibuat.",
	MsgTocUpdated:              "TOC berhasil diperbarui.",
	MsgTocDeleted:              "TOC berhasil dihapus.",
	MsgTermsNotFound:           "Syarat dan ketentuan tidak ditemukan.",
	MsgTermsUpdated:            "Syarat dan ketentuan berhasil diperbarui.",
	MsgTermsCreateFailed:       "Gagal membuat syarat dan ketentuan.",
	MsgTermsDeleteFailed:       "Gagal menghapus syarat dan ketentuan.",
	MsgTermsUpdateFailed:       "Gagal memperbarui syarat dan ketentuan.",
	MsgTermsUpdateStatusFailed: "Gagal memperbarui status syarat dan ketentuan.",

	// ── Rating ────────────────────────────────────────────────────────────────
	MsgRatingSubmitted: "Rating kamu berhasil diberikan.",
	MsgRatingInvalid:   "Nilai rating tidak valid.",

	// ── Schedule ─────────────────────────────────────────────────────────────
	MsgScheduleCreated: "Jadwal berhasil dibuat.",
	MsgScheduleUpdated: "Jadwal berhasil diperbarui.",
	MsgScheduleRemoved: "Jadwal berhasil dihapus.",
}
