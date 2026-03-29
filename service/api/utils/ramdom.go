package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateOrderID(length int) string {
	b := make([]byte, length)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[num.Int64()]
	}
	return "NL" + string(b)
}

func GenerateUniqueOrderID() string {
	// Format: NL + YYYYMMDDHHMMSS + 8 random chars
	timestamp := time.Now().Format("20060102150405")

	b := make([]byte, 8)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[num.Int64()]
	}

	return fmt.Sprintf("NL%s%s", timestamp, string(b))
}
