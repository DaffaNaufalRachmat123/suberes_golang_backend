// Package i18n menyediakan utilitas terjemahan untuk response HTTP.
// Bahasa ditentukan dari header "device_language" pada setiap request.
// Nilai yang valid: "id" (Bahasa Indonesia) dan "en" (English US).
// Default: "id" jika header tidak ada atau nilainya tidak dikenali.
package i18n

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// LangContextKey adalah key yang digunakan untuk menyimpan bahasa di gin.Context.
const LangContextKey = "lang"

// LangID dan LangEN adalah kode bahasa yang didukung.
const (
	LangID = "id"
	LangEN = "en"
)

// T mengembalikan string terjemahan berdasarkan bahasa dan kunci pesan.
// Jika kunci tidak ditemukan di bahasa yang diminta, fallback ke Bahasa Indonesia.
// Jika tetap tidak ditemukan, kembalikan kunci itu sendiri (agar tidak blank).
func T(lang, key string) string {
	var table map[string]string
	switch strings.ToLower(lang) {
	case LangEN:
		table = enTranslations
	default:
		table = idTranslations
	}
	if msg, ok := table[key]; ok {
		return msg
	}
	// fallback ke ID
	if msg, ok := idTranslations[key]; ok {
		return msg
	}
	// last resort: kembalikan key itu sendiri
	return key
}

// GetLang mengambil kode bahasa yang sudah diset oleh i18n middleware
// dari gin.Context. Jika tidak ada, default "id".
func GetLang(c *gin.Context) string {
	if lang, exists := c.Get(LangContextKey); exists {
		if s, ok := lang.(string); ok && s != "" {
			return s
		}
	}
	return LangID
}

// Tc adalah shortcut: mengambil bahasa dari gin.Context lalu memanggil T.
// Ini adalah fungsi yang paling sering dipakai di controller.
func Tc(c *gin.Context, key string) string {
	return T(GetLang(c), key)
}

// ParseLang menormalisasi nilai header device_language ke "id" atau "en".
// Nilai yang tidak dikenali dikembalikan sebagai "id" (default).
func ParseLang(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case LangEN, "en-us", "en_us", "english":
		return LangEN
	default:
		return LangID
	}
}
