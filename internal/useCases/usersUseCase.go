package usecases

import (
	"errors"
	"fmt"

	"github.com/itzLilix/questboard-profile-service/internal/repositories"
	"github.com/itzLilix/questboard-shared/dtos"
)

type UsersUsecase interface {
	GetPublicProfile(username string) (*dtos.PublicProfile, error)
	GetPrivateProfile(username string) (*dtos.PrivateProfile, error)
	UpdateProfile(input *UpdateProfileInput) (*dtos.PrivateProfile, error)
}

type usersUsecase struct {
	repo UsersRepository
}

type UpdateProfileInput struct {
    UserID      string
    Username    *string
    DisplayName *string
    AvatarURL   *string
    BannerURL   *string
    Bio         *string
}

func NewUsersUsecase(repo UsersRepository) UsersUsecase {
	return &usersUsecase{repo: repo}
}

func (s *usersUsecase) GetPublicProfile(username string) (*dtos.PublicProfile, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get public profile: %w", err)
	}
	return mapUserToPublicProfile(user), nil
}

func (s *usersUsecase) GetPrivateProfile(userID string) (*dtos.PrivateProfile, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get private profile: %w", err)
	}
	return mapUserToPrivateProfile(user), nil
}

func (s *usersUsecase) UpdateProfile(input *UpdateProfileInput) (*dtos.PrivateProfile, error) {

	if input.Username != nil {
		err := validateUsername(*input.Username)
		if err != nil {
			return nil, wrapInvalidDataError(err)
		}
	}

	if input.DisplayName != nil {
		err := validateDisplayName(*input.DisplayName)
		if err != nil {
			return nil, wrapInvalidDataError(err)
		}
	}

	if input.AvatarURL != nil {
		err := validateURL(*input.AvatarURL)
		if err != nil {
			return nil, wrapInvalidDataError(err)
		}
	}

	if input.BannerURL != nil {
		err := validateURL(*input.BannerURL)
		if err != nil {
			return nil, wrapInvalidDataError(err)
		}
	}

	if input.Bio != nil {
		err := validateBio(*input.Bio)
		if err != nil {
			return nil, wrapInvalidDataError(err)
		}
	}

	user, err := s.repo.UpdateUser(mapUpdateInputToRepoParams(input))
	if err != nil {
		if errors.Is(err, repositories.ErrDuplicateUsername) {
			return nil, ErrUsernameExists
		}
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return mapUserToPrivateProfile(user), nil
}