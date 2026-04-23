package usecase

import (
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/itzLilix/questboard-profile-service/internal/infrastructure"
	"github.com/itzLilix/questboard-shared/dtos"
)

type UsersUsecase interface {
	GetPublicProfile(viewer *ViewerContext, username string) (*dtos.PublicProfileData, error)
	GetPrivateProfile(viewer *ViewerContext) (*dtos.PrivateProfileData, error)
	UpdateProfile(viewer *ViewerContext, input *UpdateProfileInput) (*dtos.PrivateProfileData, error)
	UpdateAvatar(viewer *ViewerContext, file *multipart.FileHeader) (*dtos.PrivateProfileData, error)
	RemoveAvatar(viewer *ViewerContext) (*dtos.PrivateProfileData, error)
	UpdateBanner(viewer *ViewerContext, file *multipart.FileHeader) (*dtos.PrivateProfileData, error)
	RemoveBanner(viewer *ViewerContext) (*dtos.PrivateProfileData, error)
	Follow(viewer *ViewerContext, targetUsername string) error
	Unfollow(viewer *ViewerContext, targetUsername string) error
}

type usersUsecase struct {
	repo   UsersRepository
	images ImageUploader
}

type UpdateProfileInput struct {
	Username    *string     `json:"username"`
	DisplayName *string     `json:"displayName"`
	Bio         *string     `json:"bio"`
	Links       []dtos.Link `json:"links"`
}

type ViewerContext struct {
	UserID      string
	IsAdmin     bool
	IsFollowing bool
}

func (v *ViewerContext) IsAuthenticated() bool {
	return v != nil && v.UserID != ""
}

func NewUsersUsecase(repo UsersRepository, images ImageUploader) UsersUsecase {
	return &usersUsecase{repo: repo, images: images}
}

func (s *usersUsecase) GetPublicProfile(viewer *ViewerContext, username string) (*dtos.PublicProfileData, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return nil, mapRepoErr("get public profile", err)
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
	}

	return profile, nil
}

func (s *usersUsecase) GetPrivateProfile(viewer *ViewerContext) (*dtos.PrivateProfileData, error) {
	user, err := s.repo.GetUserByID(viewer.UserID)
	if err != nil {
		return nil, mapRepoErr("get private profile", err)
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
	if input.Bio != nil {
		if err := validateBio(*input.Bio); err != nil {
			return nil, wrapInvalidDataError(err)
		}
	}

	user, err := s.repo.UpdateUser(mapUpdateInputToRepoParams(viewer.UserID, input))
	if err != nil {
		if errors.Is(err, infrastructure.ErrNoNewData) {
			return nil, wrapInvalidDataError(err)
		}
		return nil, mapRepoErr("update profile", err)
	}
	return mapUserToPrivateProfile(user), nil
}

func (s *usersUsecase) UpdateAvatar(viewer *ViewerContext, file *multipart.FileHeader) (*dtos.PrivateProfileData, error) {
	return s.uploadImage(viewer.UserID, file, "avatars", func(url *string) *infrastructure.UpdateUserParams {
		return &infrastructure.UpdateUserParams{UserID: viewer.UserID, AvatarURL: url}
	})
}

func (s *usersUsecase) RemoveAvatar(viewer *ViewerContext) (*dtos.PrivateProfileData, error) {
	user, err := s.repo.UpdateUser(&infrastructure.UpdateUserParams{UserID: viewer.UserID, RemoveAvatar: true})
	if err != nil {
		return nil, mapRepoErr("remove avatar", err)
	}
	return mapUserToPrivateProfile(user), nil
}

func (s *usersUsecase) UpdateBanner(viewer *ViewerContext, file *multipart.FileHeader) (*dtos.PrivateProfileData, error) {
	return s.uploadImage(viewer.UserID, file, "banners", func(url *string) *infrastructure.UpdateUserParams {
		return &infrastructure.UpdateUserParams{UserID: viewer.UserID, BannerURL: url}
	})
}

func (s *usersUsecase) RemoveBanner(viewer *ViewerContext) (*dtos.PrivateProfileData, error) {
	user, err := s.repo.UpdateUser(&infrastructure.UpdateUserParams{UserID: viewer.UserID, RemoveBanner: true})
	if err != nil {
		return nil, mapRepoErr("remove banner", err)
	}
	return mapUserToPrivateProfile(user), nil
}

func (s *usersUsecase) uploadImage(
	userID string,
	file *multipart.FileHeader,
	subdir string,
	buildParams func(url *string) *infrastructure.UpdateUserParams,
) (*dtos.PrivateProfileData, error) {
	url, err := s.images.Upload(file, subdir)
	if err != nil {
		if errors.Is(err, infrastructure.ErrFileTooLarge) {
			return nil, ErrFileTooLarge
		}
		if errors.Is(err, infrastructure.ErrInvalidFileType) {
			return nil, ErrInvalidFileType
		}
		return nil, fmt.Errorf("upload %s: %w", subdir, err)
	}

	user, err := s.repo.UpdateUser(buildParams(&url)); 
	if err != nil {
		return nil, mapRepoErr("upload "+subdir, err)
	}
	return mapUserToPrivateProfile(user), nil
}

func (s *usersUsecase) Follow(viewer *ViewerContext, targetUsername string) error {
	followedID, err := s.repo.GetUserIDByUsername(targetUsername)
	if err != nil {
		return mapRepoErr("follow", err)
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
		mapRepoErr("unfollow", err)
	}

	if err := s.repo.Unfollow(viewer.UserID, followedID); err != nil {
		return fmt.Errorf("unfollow: %w", err)
	}
	viewer.IsFollowing = false
	return nil
}
