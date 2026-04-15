package useCases

import (
	"time"

	"github.com/itzLilix/questboard-profile-service/internal/entities"
	dtos "github.com/itzLilix/questboard-shared/DTOs"
)

type AuthRepository interface {
	CreateUser(user *entities.User) error
	GetUserByEmail(email string) (*entities.User, error)
	GetUserByID(id string) (*entities.User, error)
	SaveRefreshToken(token *entities.RefreshToken) error
	GetRefreshTokenByPrefix(prefix string) (*entities.RefreshToken, error)
	DeleteRefreshToken(prefix string) error
	UpdateLastLogin(user *entities.User) error
}

type UsersRepository interface {
	GetUserByUsername(username string) (*entities.User, error)
}

type TokenProvider interface{
	GenerateAccessToken(userID string, role dtos.Role) (string, error)
	GenerateRefreshToken() (string, string, time.Time, error)
	ParseToken(tokenString string) (*dtos.TokenClaims, error)
	IsRefreshTokenValid(clientToken, storedTokenHash string) bool
}

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CompareHashAndPassword(hash, password string) error
}