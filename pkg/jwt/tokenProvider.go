package jwt

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type tokenProvider struct {
	secretKey []byte
}

type TokenClaims struct {
	UserID string
    Role   string
}

type claims struct {
	UserID string `json:"user_id"`
    Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewTokenProvider() *tokenProvider {
	return &tokenProvider{
		secretKey: []byte(os.Getenv("JWT_SECRET")),
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

func (tp *tokenProvider) ParseToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (any, error) {
		return tp.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if c, ok := token.Claims.(*claims); ok && token.Valid {
		return &TokenClaims{
			UserID: c.UserID,
			Role:   c.Role,
		}, nil
	}

	return nil, fmt.Errorf("invalid claims")
}