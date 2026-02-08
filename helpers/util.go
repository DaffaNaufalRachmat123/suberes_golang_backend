package helpers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"image"
	"math"
	"math/big"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	middleware "suberes_golang/middlewares"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

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
	s := fmt.Sprintf("%d", value)
	n := len(s)
	if n <= 3 {
		return "Rp " + s
	}

	var res []string
	for n > 3 {
		res = append([]string{s[n-3:]}, res...)
		s = s[:n-3]
		n = len(s)
	}
	if n > 0 {
		res = append([]string{s}, res...)
	}

	return "Rp " + strings.Join(res, ".")
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

const (
	CustomerRole = "customer"
	MitraRole    = "mitra"
	AdminRole    = "admin"
)

var AllRole = []string{CustomerRole, MitraRole, AdminRole}
