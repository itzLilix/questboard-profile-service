package usecases

import (
	"errors"
	"fmt"

	"github.com/itzLilix/questboard-profile-service/internal/repositories"
	"github.com/itzLilix/questboard-shared/dtos"
)

type UsersUsecase interface {
	GetPublicProfile(viewer *ViewerContext, username string) (*dtos.PublicProfileData, error)
	GetPrivateProfile(viewer *ViewerContext) (*dtos.PrivateProfileData, error)
	UpdateProfile(viewer *ViewerContext, input *UpdateProfileInput) (*dtos.PrivateProfileData, error)
	Follow(viewer *ViewerContext, targetUsername string) error
	Unfollow(viewer *ViewerContext, targetUsername string) error
}

type usersUsecase struct {
	repo UsersRepository
}

type UpdateProfileInput struct {
	Username     *string
	DisplayName  *string
	AvatarURL    *string
	RemoveAvatar bool
	BannerURL    *string
	RemoveBanner bool
	Bio          *string
}

type ViewerContext struct {
	UserID      string
	IsAdmin     bool
	IsFollowing bool
}

func (v *ViewerContext) IsAuthenticated() bool {
	return v != nil && v.UserID != ""
}

func NewUsersUsecase(repo UsersRepository) UsersUsecase {
	return &usersUsecase{repo: repo}
}

func (s *usersUsecase) GetPublicProfile(viewer *ViewerContext, username string) (*dtos.PublicProfileData, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get public profile: %w", err)
	}

	profile := mapUserToPublicProfile(user)

	if viewer.IsAuthenticated() {
		if viewer.UserID == user.ID {
			following, err := s.repo.IsFollowing(viewer.UserID, user.ID)
			if err != nil {
				return nil, fmt.Errorf("get public profile: %w", err)
			}
			viewer.IsFollowing = following
			(*profile).IsFollowed = following
		}
	} else {
		(*profile).IsFollowed = false
		fmt.Println((*profile).IsFollowed)
	}

	return profile, nil
}

func (s *usersUsecase) GetPrivateProfile(viewer *ViewerContext) (*dtos.PrivateProfileData, error) {
	user, err := s.repo.GetUserByID(viewer.UserID)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get private profile: %w", err)
	}
	return mapUserToPrivateProfile(user), nil
}

func (s *usersUsecase) UpdateProfile(viewer *ViewerContext, input *UpdateProfileInput) (*dtos.PrivateProfileData, error) {
	if input.Username != nil {
		if err := validateUsername(*input.Username); err != nil {
			return nil, wrapInvalidDataError(err)
		}
	}
	if input.DisplayName != nil {
		if err := validateDisplayName(*input.DisplayName); err != nil {
			return nil, wrapInvalidDataError(err)
		}
	}
	if input.AvatarURL != nil {
		if input.RemoveAvatar {
			return nil, ErrInvalidData
		}
		if err := validateURL(*input.AvatarURL); err != nil {
			return nil, wrapInvalidDataError(err)
		}
	}
	if input.BannerURL != nil {
		if input.RemoveBanner {
			return nil, ErrInvalidData
		}
		if err := validateURL(*input.BannerURL); err != nil {
			return nil, wrapInvalidDataError(err)
		}
	}
	if input.Bio != nil {
		if err := validateBio(*input.Bio); err != nil {
			return nil, wrapInvalidDataError(err)
		}
	}

	user, err := s.repo.UpdateUser(mapUpdateInputToRepoParams(viewer.UserID, input))
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

func (s *usersUsecase) Follow(viewer *ViewerContext, targetUsername string) error {
	followedID, err := s.repo.GetUserIDByUsername(targetUsername)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("follow: %w", err)
	}

	if followedID == viewer.UserID {
		return ErrCannotFollowSelf
	}

	if err := s.repo.Follow(viewer.UserID, followedID); err != nil {
		return fmt.Errorf("follow: %w", err)
	}
	viewer.IsFollowing = true
	return nil
}

func (s *usersUsecase) Unfollow(viewer *ViewerContext, targetUsername string) error {
	followedID, err := s.repo.GetUserIDByUsername(targetUsername)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("unfollow: %w", err)
	}

	if err := s.repo.Unfollow(viewer.UserID, followedID); err != nil {
		return fmt.Errorf("unfollow: %w", err)
	}
	viewer.IsFollowing = false
	return nil
}
