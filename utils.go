package main

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// generateUserID generates a unique user ID
func generateUserID() string {
	return uuid.New().String()[:12]
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), randomString(8))
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}

// ValidateUsername checks if username is valid
func ValidateUsername(username string) bool {
	if len(username) < 1 || len(username) > 100 {
		return false
	}
	return true
}

// ValidateRoom checks if room name is valid
func ValidateRoom(room string) bool {
	if len(room) < 1 || len(room) > 100 {
		return false
	}
	return true
}

// ValidateContent checks if message content is valid
func ValidateContent(content string) bool {
	if len(content) < 1 || len(content) > 10000 {
		return false
	}
	return true
}
