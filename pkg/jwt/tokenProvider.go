package jwt

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/itzLilix/QuestBoard/backend/internal/models"
)

type tokenProvider struct {
	secretKey []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type claims struct {
	UserID string `json:"user_id"`
    Role   models.Role `json:"role"`
	jwt.RegisteredClaims
}

const (
	refreshTokenLength = 32
)

func NewTokenProvider(secretKey []byte, accessTokenTTL, refreshTokenTTL time.Duration) *tokenProvider {
	return &tokenProvider{
		secretKey: secretKey,
		accessTokenTTL: accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (tp *tokenProvider) GenerateAccessToken(userID string, role models.Role) (string, error) {
	expirationTime := time.Now().Add(tp.accessTokenTTL)
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	})

	return token.SignedString(tp.secretKey)
}

func (tp *tokenProvider) GenerateRefreshToken() (string, string, time.Time, error) {
	tokenBytes := make([]byte, refreshTokenLength)
	n, err := rand.Read(tokenBytes)
	if err != nil || n != len(tokenBytes) {
		return "", "", time.Time{}, fmt.Errorf("generateRefreshToken: %w", err)
	}
	tokenString := hex.EncodeToString(tokenBytes)

	hash := sha256.Sum256([]byte(tokenString))
	hashString := hex.EncodeToString(hash[:])

	return tokenString, hashString, time.Now().Add(tp.refreshTokenTTL), nil
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

	return subtle.ConstantTimeCompare([]byte(clientHash), []byte(storedTokenHash)) == 1
}