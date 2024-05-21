package utils

import (
	"math/rand"

	"github.com/google/uuid"
)

func randomString(length int, charset string) string {
	b := make([]byte, length)

	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

func GenerateToken() string {
	return uuid.New().String()
}
