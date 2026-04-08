package models

import (
	"time"
)

type RefreshToken struct {
	ID string `json:"id"`
	UserID string `json:"userId"`
	TokenPrefix string `json:"tokenPrefix"`
	TokenHash string `json:"tokenHash"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

type TokenClaims struct {
	UserID string
    Role   string
}