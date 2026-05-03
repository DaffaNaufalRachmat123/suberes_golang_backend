package helpers

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(999999))
}

func GenerateInvoice(prefix string) string {
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 4)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return fmt.Sprintf("%s-%s", prefix, string(b))
}
