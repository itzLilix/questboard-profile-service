package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/itzLilix/questboard-profile-service/internal/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCatalogUsecase_ListUsers_NoFollowFilter(t *testing.T) {
	ctx := context.Background()
	repo := NewMockCatalogRepository(t)
	uc := NewCatalogUsecase(repo)

	viewer := &Viewer{UserID: "viewer1"}
	filter := ListUsersFilter{Search: "rogue", Limit: 10}

	rows := []infrastructure.UserCardRow{{ID: "u1", Username: "john", DisplayName: "John"}}
	repo.EXPECT().
		GetUsersList(ctx, mock.MatchedBy(func(f *infrastructure.UserCatalogFilter) bool {
			return f.Search == "rogue" && f.FollowedByID == "" && f.Limit == 10
		}), "viewer1").
		Return(rows, "next", nil)

	page, err := uc.ListUsers(ctx, viewer, filter)
	require.NoError(t, err)
	assert.Equal(t, "next", page.NextCursor)
	assert.Len(t, page.Items, 1)
	assert.Equal(t, "u1", page.Items[0].ID)
}

func TestCatalogUsecase_ListUsers_FollowedByMe(t *testing.T) {
	ctx := context.Background()
	repo := NewMockCatalogRepository(t)
	uc := NewCatalogUsecase(repo)

	viewer := &Viewer{UserID: "viewer1"}
	filter := ListUsersFilter{FollowedBy: "me"}

	repo.EXPECT().
		GetUsersList(ctx, mock.MatchedBy(func(f *infrastructure.UserCatalogFilter) bool {
			return f.FollowedByID == "viewer1"
		}), "viewer1").
		Return([]infrastructure.UserCardRow{}, "", nil)

	page, err := uc.ListUsers(ctx, viewer, filter)
	require.NoError(t, err)
	assert.Empty(t, page.Items)
}

func TestCatalogUsecase_ListUsers_FollowedByUsername(t *testing.T) {
	ctx := context.Background()
	repo := NewMockCatalogRepository(t)
	uc := NewCatalogUsecase(repo)

	viewer := &Viewer{UserID: "viewer1"}
	filter := ListUsersFilter{FollowedBy: "alice"}

	repo.EXPECT().GetUserIDByUsername(ctx, "alice").Return("alice-id", nil)
	repo.EXPECT().
		GetUsersList(ctx, mock.MatchedBy(func(f *infrastructure.UserCatalogFilter) bool {
			return f.FollowedByID == "alice-id"
		}), "viewer1").
		Return([]infrastructure.UserCardRow{}, "", nil)

	_, err := uc.ListUsers(ctx, viewer, filter)
	require.NoError(t, err)
}

func TestCatalogUsecase_ListUsers_FollowedByUsername_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := NewMockCatalogRepository(t)
	uc := NewCatalogUsecase(repo)

	viewer := &Viewer{UserID: "viewer1"}
	filter := ListUsersFilter{FollowedBy: "ghost"}

	repo.EXPECT().GetUserIDByUsername(ctx, "ghost").Return("", infrastructure.ErrUserNotFound)

	_, err := uc.ListUsers(ctx, viewer, filter)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestCatalogUsecase_ListUsers_RepoError(t *testing.T) {
	ctx := context.Background()
	repo := NewMockCatalogRepository(t)
	uc := NewCatalogUsecase(repo)

	viewer := &Viewer{UserID: "viewer1"}
	repo.EXPECT().GetUsersList(ctx, mock.Anything, "viewer1").
		Return(nil, "", errors.New("db down"))

	_, err := uc.ListUsers(ctx, viewer, ListUsersFilter{})
	assert.ErrorIs(t, err, ErrInternal)
}
