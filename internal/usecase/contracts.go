package usecase

import (
	"mime/multipart"
	"time"

	"github.com/itzLilix/questboard-profile-service/internal/entities"
	"github.com/itzLilix/questboard-profile-service/internal/infrastructure"
	"github.com/itzLilix/questboard-shared/dtos"
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
	GetUserByID(id string) (*entities.User, error)
	GetUserIDByUsername(username string) (string, error)
	UpdateUser(input *infrastructure.UpdateUserParams) (*entities.User, error)
	Follow(followerID, followedID string) error
	Unfollow(followerID, followedID string) error
	IsFollowing(followerID, followedID string) (bool, error)
}

type CatalogRepository interface {
	GetUserIDByUsername(username string) (string, error)
	GetUsersList(filter *infrastructure.UserCatalogFilter) ([]infrastructure.UserCardRow, error)
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

type ImageUploader interface {
	Upload(file *multipart.FileHeader, subdir string) (string, error)
}