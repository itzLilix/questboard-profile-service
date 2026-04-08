package jwt

import (
	"fmt"
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


func (tp *tokenProvider) GenerateToken(userID, role string) (string, error) {
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