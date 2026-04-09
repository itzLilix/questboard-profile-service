package jwt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/itzLilix/QuestBoard/backend/internal/models"
)

type tokenProvider struct {
	secretKey []byte
}

type claims struct {
	UserID string `json:"user_id"`
    Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewTokenProvider(secretKey []byte) *tokenProvider {
	return &tokenProvider{
		secretKey: secretKey,
	}
}

func (tp *tokenProvider) GenerateAccessToken(userID, role string) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	})

	return token.SignedString(tp.secretKey)
}

func (tp *tokenProvider) GenerateRefreshToken() (string, string, error) {
	tokenBytes := make([]byte, 32)
	n, err := rand.Read(tokenBytes)
	if err != nil || n != len(tokenBytes) {
		return "", "", fmt.Errorf("generateRefreshToken: %w", err)
	}
	tokenString := hex.EncodeToString(tokenBytes)

	hash := sha256.Sum256([]byte(tokenString))
	hashString := hex.EncodeToString(hash[:])

	return tokenString, hashString,nil
}

func (tp *tokenProvider) ParseToken(tokenString string) (*models.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (any, error) {
		return tp.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if c, ok := token.Claims.(*claims); ok && token.Valid {
		return &models.TokenClaims{
			UserID: c.UserID,
			Role:   c.Role,
		}, nil
	}

	return nil, fmt.Errorf("invalid claims")
}

func (tp *tokenProvider) IsRefreshTokenValid(clientToken, storedTokenHash string) bool {
	clientHashBytes := sha256.Sum256([]byte(clientToken))
	clientHash := hex.EncodeToString(clientHashBytes[:])

	if !strings.EqualFold(storedTokenHash, clientHash) {
		return false
	}
	return true
}