package usecase

import (
	"errors"
	"fmt"

	"github.com/itzLilix/questboard-profile-service/internal/entities"
	"github.com/itzLilix/questboard-profile-service/internal/infrastructure"
	"github.com/itzLilix/questboard-shared/dtos"
)

func mapRepoErr(op string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, infrastructure.ErrUserNotFound) {
		return ErrUserNotFound
	}
	if errors.Is(err, infrastructure.ErrDuplicateUsername) {
		return ErrUsernameExists
	}
	if errors.Is(err, infrastructure.ErrDuplicateEmail) {
		return ErrEmailExists
	}
	return fmt.Errorf("%s: %w", op, err)
}


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
		UserID:      userID,
		Username:    in.Username,
		DisplayName: in.DisplayName,
		Bio:         in.Bio,
		Links:       in.Links,
	}
}

func mapListUsersFilterToUserCatalogFilter(filter ListUsersFilter, followedByID string) *infrastructure.UserCatalogFilter {
	return &infrastructure.UserCatalogFilter{
		Search:       filter.Search,
		Format:       filter.Format,
		Type:         filter.Type,
		City:         filter.City,
		MinRating:    filter.MinRating,
		FollowedByID: followedByID,
		Sort:         filter.Sort,
		OnlyGMs:      filter.OnlyGMs,
		Cursor:       filter.Cursor,
		Limit:        filter.Limit,
	}
}

func mapUserCardRowToProfileCardData(rows []infrastructure.UserCardRow) []dtos.ProfileCardData {
	if rows == nil {
		return nil
	}
	
	result := make([]dtos.ProfileCardData, len(rows))
	for i, row := range rows {
		result[i] = dtos.ProfileCardData{
			ID:              row.ID,
			Username:        row.Username,
			DisplayName:     row.DisplayName,
			AvatarURL:       row.AvatarURL,
			BannerURL:       row.BannerURL,
			Rating:          row.Rating,
			ReviewsCount:    row.ReviewsCount,
			SessionsPlayed:  row.SessionsPlayed,
			SessionsHosted:  row.SessionsHosted,
			PreferredFormat: row.PreferredFormat,
			PreferredType:   row.PreferredType,
			IsFollowed:      row.IsFollowed,
		}
	}
	return result
}
