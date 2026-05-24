package usecase

import (
	"context"

	"github.com/itzLilix/questboard-shared/dtos"
)

type ListUsersFilter struct {
	Search     		string
	Format    		dtos.SessionFormat
	Type      		dtos.SessionType
	City			string
	MinRating  		float64
	FollowedBy   	string
	OnlyGMs			bool
	Sort       		dtos.UserListSort
	SortOrder	 	dtos.SortOrder
	Cursor     		string
	Limit      		uint64
}

type catalogUsecase struct {
	repo CatalogRepository
}

func NewCatalogUsecase(repo CatalogRepository) *catalogUsecase{
	return &catalogUsecase{repo:repo};
}

func (uc *catalogUsecase) ListUsers(ctx context.Context, viewer *Viewer, filter ListUsersFilter) (*dtos.Page[dtos.ProfileCardData], error) {
	var followedByID string
	if filter.FollowedBy == "me" {
		followedByID = viewer.UserID
	} else if filter.FollowedBy != "" {
		var err error
		followedByID, err = uc.repo.GetUserIDByUsername(ctx, filter.FollowedBy)
		if err != nil {
			return nil, mapRepoErr("list users: get followed by id", err)
		}
	}

	rows, nextCursor, err := uc.repo.GetUsersList(ctx, mapListUsersFilterToUserCatalogFilter(filter, followedByID), viewer.UserID)
	if err != nil {
		return nil, mapRepoErr("list users", err)
	}
	return &dtos.Page[dtos.ProfileCardData]{
		Items:      mapUserCardRowToProfileCardData(rows),
		NextCursor: nextCursor,
	}, nil
}

