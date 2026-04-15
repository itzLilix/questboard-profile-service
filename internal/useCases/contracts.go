package useCases

import (
	"time"

	"github.com/itzLilix/questboard-shared/models"
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

type UsersRepository interface {
	GetUserByUsername(username string) (*models.User, error)
}

type TokenProvider interface{
	GenerateAccessToken(userID string, role models.Role) (string, error)
	GenerateRefreshToken() (string, string, time.Time, error)
	ParseToken(tokenString string) (*models.TokenClaims, error)
	IsRefreshTokenValid(clientToken, storedTokenHash string) bool
}

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CompareHashAndPassword(hash, password string) error
}