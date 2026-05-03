package helpers

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"image"
	"math"
	"math/big"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	middleware "suberes_golang/middlewares"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var RedisClient = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

func SetValue(key string, value string) error {
	return RedisClient.Set(ctx, key, value, 0).Err()
}

func GetValue(key string) (string, error) {
	return RedisClient.Get(ctx, key).Result()
}

func DeleteValue(key string) error {
	return RedisClient.Del(ctx, key).Err()
}

func GenerateRsaKey() (publicKeyPEM string, privateKeyPEM string, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", err
	}

	passphrase := []byte(os.Getenv("PASSPHRASE_RSA"))

	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	encBlock, err := x509.EncryptPEMBlock(
		rand.Reader,
		"RSA PRIVATE KEY",
		privBytes,
		passphrase,
		x509.PEMCipherAES256,
	)
	if err != nil {
		return "", "", err
	}

	privateKeyPEM = string(pem.EncodeToMemory(encBlock))

	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}

	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}

	publicKeyPEM = string(pem.EncodeToMemory(pubBlock))

	return
}

func GeneratePublicKey(sharedPrime, sharedBase, secret int64) string {
	p := big.NewInt(sharedPrime)
	g := big.NewInt(sharedBase)
	s := big.NewInt(secret)

	result := new(big.Int).Exp(g, s, p)
	return result.String()
}

func GenerateSharedSecret(publicKey string, secret int64, prime int64) string {
	pub := new(big.Int)
	pub.SetString(publicKey, 10)

	s := big.NewInt(secret)
	p := big.NewInt(prime)

	result := new(big.Int).Exp(pub, s, p)
	return Clean(result.String())
}

func GetFormattedYearMonthDateTimeZone(date interface{}, timezoneCode string) string {
	d, ok := date.(time.Time)
	if !ok {
		return ""
	}

	loc, err := time.LoadLocation(timezoneCode)
	if err != nil {
		return ""
	}

	converted := d.In(loc)
	return converted.Format("2006-01-02 15:04:05")
}

func Clean(number string) string {
	runes := []rune(number)
	n := len(runes)

	for i := 0; i < n/2; i++ {
		runes[i], runes[n-1-i] = runes[n-1-i], runes[i]
	}
	return string(runes)
}

func EncryptAES256(password string, data string) (string, error) {
	key := sha256.Sum256([]byte(password))
	iv := bytes.Repeat([]byte("0"), aes.BlockSize)

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	padding := aes.BlockSize - len(data)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	plaintext := append([]byte(data), padtext...)

	ciphertext := make([]byte, len(plaintext))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptAES256(password string, encrypted string) (string, error) {
	key := sha256.Sum256([]byte(password))
	iv := bytes.Repeat([]byte("0"), aes.BlockSize)

	cipherData, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherData, cipherData)

	padding := int(cipherData[len(cipherData)-1])
	plaintext := cipherData[:len(cipherData)-padding]

	return string(plaintext), nil
}

func DecryptRSA(privateKeyPEM string, encrypted string) ([]byte, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("invalid private key")
	}

	passphrase := []byte(os.Getenv("PASSPHRASE_RSA"))

	der, err := x509.DecryptPEMBlock(block, passphrase)
	if err != nil {
		return nil, err
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(der)
	if err != nil {
		return nil, err
	}

	cipherData, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, err
	}

	return rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherData)
}

// EncryptPinCbc replicates Node.js AES-256-CBC pin encryption:
//
//	key = sha512(SECRET_KEY).hex[:32], iv = sha512(SECRET_IV_KEY).hex[:16]
//	output = base64( hex( aes_cbc_encrypt(plaintext) ) )
func EncryptPinCbc(plaintext string) (string, error) {
	key := pinDeriveBytes(os.Getenv("SECRET_KEY"), 32)
	iv := pinDeriveBytes(os.Getenv("SECRET_IV_KEY"), 16)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// PKCS7 padding
	blockSize := aes.BlockSize
	padding := blockSize - len(plaintext)%blockSize
	padded := append([]byte(plaintext), bytes.Repeat([]byte{byte(padding)}, padding)...)

	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)

	// hex-encode the cipher bytes, then base64-encode the hex string (mirrors JS behaviour)
	hexStr := hex.EncodeToString(ciphertext)
	return base64.StdEncoding.EncodeToString([]byte(hexStr)), nil
}

func pinDeriveBytes(secret string, length int) []byte {
	sum := sha512.Sum512([]byte(secret))
	hexStr := hex.EncodeToString(sum[:])
	return []byte(hexStr[:length])
}

func FindPercentageFromDifferentValue(number1, number2 float64) float64 {
	if number1-number2 == 0 {
		return 0
	}

	result := 100 * math.Abs((number1-number2)/number2)
	result = math.Round(result*100) / 100

	if number1 > number2 {
		return -result
	}
	return result
}

var Days = []string{
	"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu",
}

func GetRandomPrimeNumber() int64 {
	for {
		// range: 1000 - 5000 (cukup untuk shared key ringan)
		nBig, _ := rand.Int(rand.Reader, big.NewInt(4000))
		n := nBig.Int64() + 1000

		if isPrime(n) {
			return n
		}
	}
}

func isPrime(n int64) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}

	sqrtN := int64(math.Sqrt(float64(n)))
	for i := int64(3); i <= sqrtN; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func FindPrimitiveRoot(p int64) int64 {
	phi := p - 1
	factors := primeFactors(phi)

	for g := int64(2); g < p; g++ {
		isPrimitive := true
		for _, factor := range factors {
			if modPow(g, phi/factor, p) == 1 {
				isPrimitive = false
				break
			}
		}
		if isPrimitive {
			return g
		}
	}

	return -1
}

func primeFactors(n int64) []int64 {
	factors := []int64{}

	for n%2 == 0 {
		factors = appendUnique(factors, 2)
		n /= 2
	}

	for i := int64(3); i*i <= n; i += 2 {
		for n%i == 0 {
			factors = appendUnique(factors, i)
			n /= i
		}
	}

	if n > 2 {
		factors = appendUnique(factors, n)
	}

	return factors
}

func appendUnique(slice []int64, val int64) []int64 {
	for _, v := range slice {
		if v == val {
			return slice
		}
	}
	return append(slice, val)
}

func modPow(base, exp, mod int64) int64 {
	result := int64(1)
	base = base % mod

	for exp > 0 {
		if exp%2 == 1 {
			result = (result * base) % mod
		}
		exp >>= 1
		base = (base * base) % mod
	}

	return result
}

func ConvertDayToString(t time.Time) string {
	return Days[int(t.Weekday())]
}

func ConvertMonthToString(m int) string {
	months := []string{
		"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	return months[m]
}

func FormatRupiah(value int64) string {
	isNegative := value < 0
	if isNegative {
		value = -value
	}

	s := fmt.Sprintf("%d", value)
	n := len(s)

	if n <= 3 {
		if isNegative {
			return "-Rp " + s
		}
		return "Rp " + s
	}

	var result []string
	for n > 3 {
		result = append([]string{s[n-3:]}, result...)
		s = s[:n-3]
		n = len(s)
	}

	if n > 0 {
		result = append([]string{s}, result...)
	}

	formatted := "Rp " + strings.Join(result, ".")

	if isNegative {
		return "-" + formatted
	}

	return formatted
}

func GetTimezoneNowDateReturnDate(timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}

	now := time.Now().In(loc)
	return now, nil
}
func GetTimezoneNowDate(timezone string) string {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)

	year := now.Year()
	month := int(now.Month())
	day := now.Day()
	hour := now.Hour()
	minute := now.Minute()
	second := now.Second()

	return fmt.Sprintf(
		"%04d-%02d-%02d %02d:%02d:%02d",
		year,
		month,
		day,
		hour,
		minute,
		second,
	)
}

// NormalizeDateTimeString ensures the date portion of a "YYYY-M-D HH:MM" or
// "YYYY-M-D HH:MM:SS" string is zero-padded to "YYYY-MM-DD HH:MM[:SS]" so
// that Go's strict time.ParseInLocation does not fail on single-digit months
// or days sent by mobile/web clients.
// If the time component only has hours and minutes (HH:MM), ":00" is appended
// so the string is also compatible with layouts that include seconds.
func NormalizeDateTimeString(s string) string {
	parts := strings.SplitN(s, " ", 2)
	if len(parts) != 2 {
		return s
	}
	dateParts := strings.Split(parts[0], "-")
	if len(dateParts) != 3 {
		return s
	}
	month, _ := strconv.Atoi(dateParts[1])
	day, _ := strconv.Atoi(dateParts[2])
	timePart := parts[1]
	// Add seconds if only HH:MM was provided
	if strings.Count(timePart, ":") == 1 {
		timePart += ":00"
	}
	return fmt.Sprintf("%s-%02d-%02d %s", dateParts[0], month, day, timePart)
}

// StringPtr returns a pointer to the given string value.
func StringPtr(s string) *string {
	return &s
}

// DerefStr safely dereferences a *string, returning "" if nil.
func DerefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// NormalizeIsAdditional converts "0"/"1" (sent by mobile clients) or
// "true"/"false" (any case) to the canonical "false"/"true" stored in DB.
func NormalizeIsAdditional(v string) string {
	if v == "1" || strings.EqualFold(v, "true") {
		return "true"
	}
	return "false"
}

func RandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}

	return string(b)
}

var monthMappingLetters = map[int]string{
	1:  "Januari",
	2:  "Feberuari",
	3:  "Maret",
	4:  "April",
	5:  "Mei",
	6:  "Juni",
	7:  "Juli",
	8:  "Agustus",
	9:  "September",
	10: "Oktober",
	11: "November",
	12: "Desember",
}

func ConvertNumberToMonthString(number string) string {
	n, err := strconv.Atoi(number)
	if err != nil {
		return ""
	}

	if month, ok := monthMappingLetters[n]; ok {
		return month
	}

	return ""
}
func GetFormattedYearMonthDate(date time.Time) string {
	if !date.IsZero() {
		year := date.Year()
		month := int(date.Month())
		day := date.Day()
		hour := date.Hour()
		minute := date.Minute()
		second := date.Second()

		return fmt.Sprintf(
			"%04d-%02d-%02d %02d:%02d:%02d",
			year,
			month,
			day,
			hour,
			minute,
			second,
		)
	}
	return ""
}
func GenerateRandomAlphaNum(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[num.Int64()]
	}
	return string(b)
}

func ExtractImagesFromText(text string) []string {
	re := regexp.MustCompile(`https?://\S+\.(jpg|jpeg|png|webp)`)
	return re.FindAllString(text, -1)
}

func BlurImage(src string, dst string) error {
	img, err := imaging.Open(src)
	if err != nil {
		return err
	}
	blur := imaging.Blur(img, 5)
	return imaging.Save(blur, dst)
}

func GenerateRSAKeyPair() (string, string, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	priv := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	pubBytes, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	pub := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	return string(priv), string(pub), nil
}

func GetOtpDuration() time.Duration {
	envVal := os.Getenv("OTP_TIMEOUT")
	if envVal == "" {
		return 1 * time.Minute
	}

	lowerVal := strings.ToLower(envVal)

	lowerVal = strings.TrimSuffix(lowerVal, "s")

	parts := strings.Fields(lowerVal)
	if len(parts) != 2 {
		return 1 * time.Minute
	}

	number, err := strconv.Atoi(parts[0])
	if err != nil {
		return 1 * time.Minute
	}

	unit := parts[1]
	fmt.Println("Duration", number)
	switch {
	case strings.Contains(unit, "minute"): // handle "minute"
		return time.Duration(number) * time.Minute
	case strings.Contains(unit, "second"): // handle "second"
		return time.Duration(number) * time.Second
	case strings.Contains(unit, "hour"): // handle "hour"
		return time.Duration(number) * time.Hour
	default:
		return 1 * time.Minute
	}
}

func GetImageDimension(file *multipart.FileHeader) (int, int, error) {
	f, err := file.Open()
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

func GenerateFilename(alias, original string) string {
	ext := filepath.Ext(original)
	t := time.Now()
	return fmt.Sprintf("%s%d-%02d-%02d_%02d-%02d-%d%s", alias,
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Nanosecond(), ext)
}
func GetHostURL(c *gin.Context) string {
	// 1. Check reverse proxy headers
	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}

	return scheme + "://" + host
}

type ProtectedRoute struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
	Roles   []string
}

func RegisterProtectedRoutes(
	group *gin.RouterGroup,
	routes []ProtectedRoute,
) {
	for _, rt := range routes {
		group.Handle(
			rt.Method,
			rt.Path,
			middleware.RoleMiddleware(rt.Roles...),
			rt.Handler,
		)
	}
}

func FormatDate(input string, format string, locale string) (string, error) {
	// parse ISO (support nano juga)
	t, err := time.Parse(time.RFC3339Nano, input)
	if err != nil {
		return "", err
	}

	// bulan Indonesia
	monthID := map[time.Month]string{
		time.January:   "Januari",
		time.February:  "Februari",
		time.March:     "Maret",
		time.April:     "April",
		time.May:       "Mei",
		time.June:      "Juni",
		time.July:      "Juli",
		time.August:    "Agustus",
		time.September: "September",
		time.October:   "Oktober",
		time.November:  "November",
		time.December:  "Desember",
	}

	switch format {

	case "DD-MM-YYYY":
		return t.Format("02-01-2006"), nil

	case "DD Month YYYY":
		if locale == "id" {
			return fmt.Sprintf("%d %s %d", t.Day(), monthID[t.Month()], t.Year()), nil
		}
		return t.Format("2 January 2006"), nil

	default:
		// custom format (Go layout)
		return t.Format(format), nil
	}
}

func RootPath() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}

	return ""
}

func DeleteImageIfExists(imagePath string) error {
	if imagePath == "" {
		return nil
	}

	fullPath := filepath.Join(
		RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
		imagePath,
	)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil // file tidak ada → aman
	}

	return os.Remove(fullPath)
}
func JSONError(ctx *gin.Context, code int, err error) {
	statusCode := mapErrorCode(code)

	ctx.JSON(statusCode, gin.H{
		"server_message": err.Error(),
		"status":         "failed",
	})
}

// Map service code ke HTTP status
func mapErrorCode(code int) int {
	switch code {
	case 400:
		return http.StatusBadRequest
	case 401:
		return http.StatusUnauthorized
	case 403:
		return http.StatusForbidden
	case 404:
		return http.StatusNotFound
	case 409:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
func CopyFields(data map[string]interface{}, target map[string]interface{}) {
	fields := []string{
		"customer_id",
		"mitra_id",
		"order_id",
		"sub_order_id",
		"service_id",
		"sub_service_id",
		"transaction_id",
		"notification_type",
		"title",
		"message",
		"notif_type",
	}

	for _, f := range fields {
		if val, ok := data[f]; ok {
			switch f {
			case "title":
				target["notification_title"] = val
			case "message":
				target["notification_message"] = val
			default:
				target[f] = val
			}
		}
	}
}

const (
	CustomerRole   = "customer"
	MitraRole      = "mitra"
	AdminRole      = "admin"
	SuperAdminRole = "superadmin"
)

var AllRole = []string{CustomerRole, MitraRole, AdminRole, SuperAdminRole}

// GenerateMitraPassword generates a 10-char random password from the mitra wishlist charset.
// Uses crypto/rand for cryptographic randomness (matches Node.js crypto.randomFillSync).
func GenerateMitraPassword() string {
	const wishlist = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz~!@-_+=#$"
	const length = 10
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(wishlist))))
		b[i] = wishlist[n.Int64()]
	}
	return string(b)
}
