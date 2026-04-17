package usecases

import (
	"github.com/itzLilix/questboard-profile-service/internal/entities"
	"github.com/itzLilix/questboard-shared/dtos"
)

func mapUserToPublicProfile(user *entities.User) *dtos.PublicProfile {
	if user == nil {
		return nil
	}

	return &dtos.PublicProfile{
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

func mapUserToPrivateProfile(user *entities.User) *dtos.PrivateProfile {
	if user == nil {
		return nil
	}

	return &dtos.PrivateProfile{
		PublicProfile:   *mapUserToPublicProfile(user),
		Email:           user.Email,
		CreatedAt:       user.CreatedAt,
		LastLogin:       user.LastLogin,
		Role:              user.Role,
		IsEmailVerified: user.IsEmailVerified,
	}
}