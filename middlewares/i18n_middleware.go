package middleware

import (
	"suberes_golang/i18n"

	"github.com/gin-gonic/gin"
)

// I18nMiddleware membaca header "device_language" dari setiap request
// dan menyimpan kode bahasa yang sudah dinormalisasi ke gin.Context
// dengan key i18n.LangContextKey ("lang").
//
// Nilai yang didukung:
//   - "id" (atau tidak ada header)  → Bahasa Indonesia (default)
//   - "en" / "en-US" / "en_US"     → English (US)
//
// Middleware ini harus dipasang secara global SEBELUM middleware auth
// agar pesan error dari auth pun bisa diterjemahkan.
func I18nMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("device_language")
		lang := i18n.ParseLang(raw)
		c.Set(i18n.LangContextKey, lang)
		c.Next()
	}
}
