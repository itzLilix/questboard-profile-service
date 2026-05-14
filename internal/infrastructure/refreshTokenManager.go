package infrastructure

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"time"
)

const refreshTokenLength = 32

type refreshTokenManager struct {
	ttl time.Duration
}

func NewRefreshTokenManager(ttl time.Duration) *refreshTokenManager {
	return &refreshTokenManager{ttl: ttl}
}

func (m *refreshTokenManager) Generate() (string, string, time.Time, error) {
	tokenBytes := make([]byte, refreshTokenLength)
	n, err := rand.Read(tokenBytes)
	if err != nil || n != len(tokenBytes) {
		return "", "", time.Time{}, fmt.Errorf("generate refresh token: %w", err)
	}
	tokenString := hex.EncodeToString(tokenBytes)

	hash := sha256.Sum256([]byte(tokenString))
	hashString := hex.EncodeToString(hash[:])

	return tokenString, hashString, time.Now().Add(m.ttl), nil
}

func (m *refreshTokenManager) IsValid(clientToken, storedTokenHash string) bool {
	clientHashBytes := sha256.Sum256([]byte(clientToken))
	clientHash := hex.EncodeToString(clientHashBytes[:])

	return subtle.ConstantTimeCompare([]byte(clientHash), []byte(storedTokenHash)) == 1
}
