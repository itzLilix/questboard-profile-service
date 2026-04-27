package usecase

import "github.com/itzLilix/questboard-shared/dtos"

type CatalogUsecase interface {
	ListUsers(viewer *ViewerContext, filter ListUsersFilter) ([]dtos.ProfileCardData, error)
}

type ListUsersFilter struct {
	Search     		string
	Format    		dtos.SessionFormat
	Type      		dtos.SessionType
	City			string
	MinRating  		float64
	FollowedBy   	string
	OnlyGMs			bool
	Sort       		dtos.UserListSort
	Cursor     		string
	Limit      		int
}

type catalogUsecase struct {
	repo CatalogRepository
}

func NewCatalogUsecase(repo CatalogRepository) *catalogUsecase{
	return &catalogUsecase{repo:repo};
}

func (uc *catalogUsecase) ListUsers(viewer *ViewerContext, filter ListUsersFilter) ([]dtos.ProfileCardData, error) {
	var followedByID string
	if filter.FollowedBy == "me" {
		followedByID = viewer.UserID
	} else if filter.FollowedBy != "" {
		var err error
		followedByID, err = uc.repo.GetUserIDByUsername(filter.FollowedBy)
		if err != nil {
			return nil, mapRepoErr("list users: get followed by id", err)
		}
	}

	cards, err := uc.repo.GetUsersList(mapListUsersFilterToUserCatalogFilter(filter, followedByID))
	if err != nil {
		return nil, mapRepoErr("list users", err)
	}
	return mapUserCardRowToProfileCardData(cards), nil
}

