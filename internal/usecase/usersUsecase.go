package usecase

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/itzLilix/questboard-profile-service/internal/infrastructure"
	"github.com/itzLilix/questboard-shared/dtos"
	"github.com/itzLilix/questboard-shared/images"
)

type UsersUsecase interface {
	GetPublicProfile(ctx context.Context, viewer *Viewer, username string) (*dtos.PublicProfileData, error)
	GetPrivateProfile(ctx context.Context, viewer *Viewer) (*dtos.PrivateProfileData, error)
	UpdateProfile(ctx context.Context, viewer *Viewer, input *UpdateProfileInput) (*dtos.PrivateProfileData, error)
	
	UpdateAvatar(ctx context.Context, viewer *Viewer, file *multipart.FileHeader) (*dtos.PrivateProfileData, error)
	RemoveAvatar(ctx context.Context, viewer *Viewer) (*dtos.PrivateProfileData, error)
	
	UpdateBanner(ctx context.Context, viewer *Viewer, file *multipart.FileHeader) (*dtos.PrivateProfileData, error)
	RemoveBanner(ctx context.Context, viewer *Viewer) (*dtos.PrivateProfileData, error)
	
	Follow(ctx context.Context, viewer *Viewer, targetUsername string) error
	Unfollow(ctx context.Context, viewer *Viewer, targetUsername string) error

	GetBriefs(ctx context.Context, ids []string) ([]dtos.UserBrief, error)
	UpdateStats(ctx context.Context, statName dtos.UserStatName, stats map[string]int) error
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

func (s *usersUsecase) isFollowing(ctx context.Context, viewer *Viewer, targetID string) (bool, error) {
    if !viewer.IsAuthenticated() || viewer.Is(targetID) {
        return false, nil
    }
    return s.repo.IsFollowing(ctx, viewer.UserID, targetID)
}

func (s *usersUsecase) GetPublicProfile(ctx context.Context, viewer *Viewer, username string) (*dtos.PublicProfileData, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, mapRepoErr("get public profile", err)
	}

	profile := mapUserToPublicProfile(user)
	if profile == nil {
		return nil, fmt.Errorf("get public profile: profile is nil")
	}

	(*profile).IsFollowed, err = s.isFollowing(ctx, viewer, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get public profile: %w", err)
	}

	return profile, nil
}

func (s *usersUsecase) GetPrivateProfile(ctx context.Context, viewer *Viewer) (*dtos.PrivateProfileData, error) {
	user, err := s.repo.GetUserByID(ctx, viewer.UserID)
	if err != nil {
		return nil, mapRepoErr("get private profile", err)
	}
	return mapUserToPrivateProfile(user), nil
}

func (s *usersUsecase) UpdateProfile(ctx context.Context, viewer *Viewer, input *UpdateProfileInput) (*dtos.PrivateProfileData, error) {
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

	user, err := s.repo.UpdateUser(ctx, mapUpdateInputToRepoParams(viewer.UserID, input))
	if err != nil {
		return nil, mapRepoErr("update profile", err)
	}
	return mapUserToPrivateProfile(user), nil
}

func (s *usersUsecase) UpdateAvatar(ctx context.Context, viewer *Viewer, file *multipart.FileHeader) (*dtos.PrivateProfileData, error) {
	return s.uploadImage(ctx, viewer.UserID, file, "avatars", func(url *string) *infrastructure.UpdateUserParams {
		return &infrastructure.UpdateUserParams{UserID: viewer.UserID, AvatarURL: url}
	})
}

func (s *usersUsecase) RemoveAvatar(ctx context.Context, viewer *Viewer) (*dtos.PrivateProfileData, error) {
	user, err := s.repo.UpdateUser(ctx, &infrastructure.UpdateUserParams{UserID: viewer.UserID, RemoveAvatar: true})
	if err != nil {
		return nil, mapRepoErr("remove avatar", err)
	}
	return mapUserToPrivateProfile(user), nil
}

func (s *usersUsecase) UpdateBanner(ctx context.Context, viewer *Viewer, file *multipart.FileHeader) (*dtos.PrivateProfileData, error) {
	return s.uploadImage(ctx, viewer.UserID, file, "banners", func(url *string) *infrastructure.UpdateUserParams {
		return &infrastructure.UpdateUserParams{UserID: viewer.UserID, BannerURL: url}
	})
}

func (s *usersUsecase) RemoveBanner(ctx context.Context, viewer *Viewer) (*dtos.PrivateProfileData, error) {
	user, err := s.repo.UpdateUser(ctx, &infrastructure.UpdateUserParams{UserID: viewer.UserID, RemoveBanner: true})
	if err != nil {
		return nil, mapRepoErr("remove banner", err)
	}
	return mapUserToPrivateProfile(user), nil
}

func (s *usersUsecase) uploadImage(
	ctx context.Context,
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

	user, err := s.repo.UpdateUser(ctx, buildParams(&url)); 
	if err != nil {
		return nil, mapRepoErr("upload "+subdir, err)
	}
	return mapUserToPrivateProfile(user), nil
}

func (s *usersUsecase) Follow(ctx context.Context, viewer *Viewer, targetUsername string) error {
	followedID, err := s.repo.GetUserIDByUsername(ctx, targetUsername)
	if err != nil {
		return mapRepoErr("follow", err)
	}

	if followedID == viewer.UserID {
		return ErrCannotFollowSelf
	}

	if err := s.repo.Follow(ctx, viewer.UserID, followedID); err != nil {
		return mapRepoErr("follow", err)
	}
	return nil
}

func (s *usersUsecase) Unfollow(ctx context.Context, viewer *Viewer, targetUsername string) error {
	followedID, err := s.repo.GetUserIDByUsername(ctx, targetUsername)
	if err != nil {
		return mapRepoErr("unfollow", err)
	}

	if err := s.repo.Unfollow(ctx, viewer.UserID, followedID); err != nil {
		return fmt.Errorf("unfollow: %w", err)
	}
	return nil
}

func (s *usersUsecase) GetBriefs(ctx context.Context, ids []string) ([]dtos.UserBrief, error) {
	if len(ids) == 0 {
		return []dtos.UserBrief{}, nil
	}
	if len(ids) > 100 {
		ids = ids[:100]
	}
	seen := make(map[string]struct{}, len(ids))
	deduped := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		deduped = append(deduped, id)
	}
	briefs, err := s.repo.GetBriefsByIDs(ctx, deduped)
	if err != nil {
		return nil, mapRepoErr("get briefs", err)
	}
	return briefs, nil
}

func (s *usersUsecase) UpdateStats(ctx context.Context, statName dtos.UserStatName, stats map[string]int) error {
	if len(stats) == 0 || len(stats) > 50 {
		return ErrInvalidData
	}

	for _, v := range stats {
		if v < 0 { return ErrInvalidData }
	} 

	err := s.repo.UpdateStats(ctx, statName, stats)
	if err != nil {
		return mapRepoErr("update session stats", err)
	}
	return nil
}