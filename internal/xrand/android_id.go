package xrand

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateAndroidID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
