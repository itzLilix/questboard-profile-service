package useCases

import (
	"github.com/itzLilix/QuestBoard/backend/internal/models"
	"github.com/itzLilix/QuestBoard/backend/pkg/jwt"
)

type AuthRepository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id string) (*models.User, error)
	SaveRefreshToken(token *models.RefreshToken) error
	GetRefreshTokenByPrefix(prefix string) (*models.RefreshToken, error)
	DeleteRefreshToken(prefix string) error
	UpdateLastLogin(user *models.User) error
}

type TokenProvider interface{
	GenerateToken(userID, role string) (string, error)
	ParseToken(tokenString string) (*jwt.TokenClaims, error)
}