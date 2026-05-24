package handlers

import (
	"context"
	"mime/multipart"

	uc "github.com/itzLilix/questboard-profile-service/internal/usecase"
	"github.com/itzLilix/questboard-shared/dtos"
)

type AuthUsecase interface {
	Register(ctx context.Context, username, displayname, email, password string) (*dtos.PrivateProfileData, string, string, error)
	Login(ctx context.Context, username, password string) (*dtos.PrivateProfileData, string, string, error)
	Logout(ctx context.Context, refreshToken string) error
	ValidateToken(ctx context.Context, tokenString string) (*dtos.PrivateProfileData, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*dtos.PrivateProfileData, string, string, error)
}

type CatalogUsecase interface {
	ListUsers(ctx context.Context, viewer *uc.Viewer, filter uc.ListUsersFilter) (*dtos.Page[dtos.ProfileCardData], error)
}

type UsersUsecase interface {
	GetPublicProfile(ctx context.Context, viewer *uc.Viewer, username string) (*dtos.PublicProfileData, error)
	GetPrivateProfile(ctx context.Context, viewer *uc.Viewer) (*dtos.PrivateProfileData, error)
	UpdateProfile(ctx context.Context, viewer *uc.Viewer, input *uc.UpdateProfileInput) (*dtos.PrivateProfileData, error)

	UpdateAvatar(ctx context.Context, viewer *uc.Viewer, file *multipart.FileHeader) (*dtos.PrivateProfileData, error)
	RemoveAvatar(ctx context.Context, viewer *uc.Viewer) (*dtos.PrivateProfileData, error)

	UpdateBanner(ctx context.Context, viewer *uc.Viewer, file *multipart.FileHeader) (*dtos.PrivateProfileData, error)
	RemoveBanner(ctx context.Context, viewer *uc.Viewer) (*dtos.PrivateProfileData, error)

	Follow(ctx context.Context, viewer *uc.Viewer, targetUsername string) error
	Unfollow(ctx context.Context, viewer *uc.Viewer, targetUsername string) error

	GetBriefs(ctx context.Context, ids []string) ([]dtos.UserBrief, error)
	UpdateStats(ctx context.Context, statName dtos.UserStatName, stats map[string]int) error
}