package usecase

import (
	"github.com/itzLilix/questboard-profile-service/internal/entities"
	"github.com/itzLilix/questboard-profile-service/internal/infrastructure"
	"github.com/itzLilix/questboard-shared/dtos"
)

func mapUserToPublicProfile(user *entities.User) *dtos.PublicProfileData {
	if user == nil {
		return nil
	}

	return &dtos.PublicProfileData{
		ID:             user.ID,
		Username:       user.Username,
		DisplayName:    user.DisplayName,
		AvatarURL:      user.AvatarURL,
		BannerURL:      user.BannerURL,
		SessionsPlayed: user.SessionsPlayed,
		SessionsHosted: user.SessionsHosted,
		Rating:         user.Rating,
		ReviewsCount:   user.ReviewsCount,
		Bio:            user.Bio,
		Links:          user.Links,
	}
}

func mapUserToPrivateProfile(user *entities.User) *dtos.PrivateProfileData {
	if user == nil {
		return nil
	}

	return &dtos.PrivateProfileData{
		PublicProfileData:   *mapUserToPublicProfile(user),
		Email:           user.Email,
		CreatedAt:       user.CreatedAt,
		LastLogin:       user.LastLogin,
		Role:            user.Role,
		IsEmailVerified: user.IsEmailVerified,
	}
}

func mapUpdateInputToRepoParams(userID string, in *UpdateProfileInput) *infrastructure.UpdateUserParams {
    return &infrastructure.UpdateUserParams{
        UserID:       userID,
        Username:     in.Username,
        DisplayName:  in.DisplayName,
        AvatarURL:    in.AvatarURL,
		RemoveAvatar: in.RemoveAvatar,
        BannerURL:    in.BannerURL,
		RemoveBanner: in.RemoveBanner,
        Bio:          in.Bio,
    }
}