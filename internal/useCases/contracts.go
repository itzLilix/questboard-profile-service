package useCases

import "github.com/itzLilix/QuestBoard/backend/internal/models"

type AuthRepository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id string) (*models.User, error)
	SaveRefreshToken(token *models.RefreshToken) error
	GetRefreshTokenByPrefix(prefix string) (*models.RefreshToken, error)
	DeleteRefreshToken(prefix string) error
	UpdateLastLogin(user *models.User) error
}