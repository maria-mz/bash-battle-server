package main

import (
	"fmt"
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

func GenerateNewToken() string {
	return uuid.New().String()
}

func GenerateGameID() GameID {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXZYabcdefghijklmnopqrstuvwxyz0123456789"

	firstHalf := randomString(3, chars)
	secondHalf := randomString(3, chars)

	id := fmt.Sprintf("%s-%s", firstHalf, secondHalf)

	return GameID(id)
}

func GenerateGameCode() string {
	chars := "0123456789"
	return randomString(4, chars)
}
