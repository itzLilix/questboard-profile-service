package usecase

import (
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/itzLilix/questboard-profile-service/internal/infrastructure"
	"github.com/itzLilix/questboard-shared/dtos"
	"github.com/itzLilix/questboard-shared/images"
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

func NewUsersUsecase(repo UsersRepository, images ImageUploader) UsersUsecase {
	return &usersUsecase{repo: repo, images: images}
}

func (s *usersUsecase) isFollowing(viewer *ViewerContext, targetID string) (bool, error) {
    if !viewer.IsAuthenticated() || viewer.UserID == targetID {
        return false, nil
    }
    return s.repo.IsFollowing(viewer.UserID, targetID)
}

func (s *usersUsecase) GetPublicProfile(viewer *ViewerContext, username string) (*dtos.PublicProfileData, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return nil, mapRepoErr("get public profile", err)
	}

	profile := mapUserToPublicProfile(user)
	if profile == nil {
		return nil, fmt.Errorf("get public profile: profile is nil")
	}

	(*profile).IsFollowed, err = s.isFollowing(viewer, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get public profile: %w", err)
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
	if len(input.Links) > 0 {
		links := make([]dtos.Link, 0, len(input.Links))
		for _, inLink := range input.Links {
			if link, err := validateAndNormalize(inLink); err != nil {
				links = nil
				return nil, wrapInvalidDataError(err)
			} else {
				links = append(links, link)
			}
		}
		if len(links) == 0 {
			return nil, ErrInvalidData
		}
		input.Links = links
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
		if errors.Is(err, images.ErrFileTooLarge) {
			return nil, ErrFileTooLarge
		}
		if errors.Is(err, images.ErrInvalidFileType) {
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
	return nil
}

func (s *usersUsecase) Unfollow(viewer *ViewerContext, targetUsername string) error {
	followedID, err := s.repo.GetUserIDByUsername(targetUsername)
	if err != nil {
		return mapRepoErr("unfollow", err)
	}

	if err := s.repo.Unfollow(viewer.UserID, followedID); err != nil {
		return fmt.Errorf("unfollow: %w", err)
	}
	return nil
}

