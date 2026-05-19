package usecase

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/itzLilix/questboard-profile-service/internal/entities"
	"github.com/itzLilix/questboard-profile-service/internal/infrastructure"
	"github.com/itzLilix/questboard-shared/dtos"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user *entities.User) error
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUserByID(ctx context.Context, id string) (*entities.User, error)
	SaveRefreshToken(ctx context.Context, token *entities.RefreshToken) error
	GetRefreshTokenByPrefix(ctx context.Context, prefix string) (*entities.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, prefix string) error
	UpdateLastLogin(ctx context.Context, user *entities.User) error
}

type UsersRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*entities.User, error)
	GetUserByID(ctx context.Context, id string) (*entities.User, error)
	GetUserIDByUsername(ctx context.Context, username string) (string, error)
	UpdateUser(ctx context.Context, input *infrastructure.UpdateUserParams) (*entities.User, error)
	Follow(ctx context.Context, followerID, followedID string) error
	Unfollow(ctx context.Context, followerID, followedID string) error
	IsFollowing(ctx context.Context, followerID, followedID string) (bool, error)
}

type CatalogRepository interface {
	GetUserIDByUsername(ctx context.Context, username string) (string, error)
	GetUsersList(ctx context.Context, filter *infrastructure.UserCatalogFilter, viewerID string) ([]infrastructure.UserCardRow, string, error)
}

type TokenProvider interface {
	GenerateAccessToken(userID string, role dtos.Role) (string, error)
	ParseToken(tokenString string) (*dtos.TokenClaims, error)
}

type RefreshTokenManager interface {
	Generate() (string, string, time.Time, error)
	IsValid(clientToken, storedTokenHash string) bool
}

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CompareHashAndPassword(hash, password string) error
}

type ImageUploader interface {
	Upload(file *multipart.FileHeader, subdir string) (string, error)
}