package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/itzLilix/questboard-profile-service/internal/entities"
	"github.com/itzLilix/questboard-profile-service/internal/infrastructure"
	"github.com/itzLilix/questboard-shared/dtos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newUsersUC(t *testing.T) (*usersUsecase, *MockUsersRepository, *MockImageUploader) {
	repo := NewMockUsersRepository(t)
	imgs := NewMockImageUploader(t)
	return NewUsersUsecase(repo, imgs), repo, imgs
}

// ---------- GetPublicProfile ----------

func TestUsersUsecase_GetPublicProfile_AnonymousViewer(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	user := &entities.User{ID: "u1", Username: "john"}
	repo.EXPECT().GetUserByUsername(ctx, "john").Return(user, nil)

	viewer := &Viewer{}
	profile, err := uc.GetPublicProfile(ctx, viewer, "john")
	require.NoError(t, err)
	assert.Equal(t, "john", profile.Username)
	assert.False(t, profile.IsFollowed)
}

func TestUsersUsecase_GetPublicProfile_SelfViewer(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	user := &entities.User{ID: "u1", Username: "john"}
	repo.EXPECT().GetUserByUsername(ctx, "john").Return(user, nil)

	viewer := &Viewer{UserID: "u1"}
	profile, err := uc.GetPublicProfile(ctx, viewer, "john")
	require.NoError(t, err)
	assert.False(t, profile.IsFollowed, "viewing self should not show as followed")
}

func TestUsersUsecase_GetPublicProfile_FollowedTrue(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	user := &entities.User{ID: "u1", Username: "john"}
	repo.EXPECT().GetUserByUsername(ctx, "john").Return(user, nil)
	repo.EXPECT().IsFollowing(ctx, "viewer", "u1").Return(true, nil)

	viewer := &Viewer{UserID: "viewer"}
	profile, err := uc.GetPublicProfile(ctx, viewer, "john")
	require.NoError(t, err)
	assert.True(t, profile.IsFollowed)
}

func TestUsersUsecase_GetPublicProfile_NotFound(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	repo.EXPECT().GetUserByUsername(ctx, "ghost").Return(nil, infrastructure.ErrUserNotFound)

	_, err := uc.GetPublicProfile(ctx, &Viewer{}, "ghost")
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestUsersUsecase_GetPublicProfile_IsFollowingError(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	user := &entities.User{ID: "u1", Username: "john"}
	repo.EXPECT().GetUserByUsername(ctx, "john").Return(user, nil)
	repo.EXPECT().IsFollowing(ctx, "viewer", "u1").Return(false, errors.New("db"))

	_, err := uc.GetPublicProfile(ctx, &Viewer{UserID: "viewer"}, "john")
	assert.Error(t, err)
}

// ---------- GetPrivateProfile ----------

func TestUsersUsecase_GetPrivateProfile_Success(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	user := &entities.User{ID: "u1", Username: "john", Email: "x@y.z"}
	repo.EXPECT().GetUserByID(ctx, "u1").Return(user, nil)

	profile, err := uc.GetPrivateProfile(ctx, &Viewer{UserID: "u1"})
	require.NoError(t, err)
	assert.Equal(t, "x@y.z", profile.Email)
}

func TestUsersUsecase_GetPrivateProfile_NotFound(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	repo.EXPECT().GetUserByID(ctx, "u1").Return(nil, infrastructure.ErrUserNotFound)

	_, err := uc.GetPrivateProfile(ctx, &Viewer{UserID: "u1"})
	assert.ErrorIs(t, err, ErrUserNotFound)
}

// ---------- UpdateProfile ----------

func TestUsersUsecase_UpdateProfile_Success(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	updated := &entities.User{ID: "u1", Username: "newname", DisplayName: "New"}
	repo.EXPECT().UpdateUser(ctx, mock.MatchedBy(func(p *infrastructure.UpdateUserParams) bool {
		return p.UserID == "u1" && p.Username != nil && *p.Username == "newname"
	})).Return(updated, nil)

	input := &UpdateProfileInput{
		Username:    strPtr("newname"),
		DisplayName: strPtr("New"),
		Bio:         strPtr("bio"),
		Links:       []dtos.Link{{Type: "telegram", URL: "https://t.me/foo"}},
	}
	profile, err := uc.UpdateProfile(ctx, &Viewer{UserID: "u1"}, input)
	require.NoError(t, err)
	assert.Equal(t, "newname", profile.Username)
}

func TestUsersUsecase_UpdateProfile_InvalidUsername(t *testing.T) {
	uc, _, _ := newUsersUC(t)
	_, err := uc.UpdateProfile(context.Background(), &Viewer{UserID: "u1"},
		&UpdateProfileInput{Username: strPtr("bad name!")})
	assert.ErrorIs(t, err, ErrInvalidData)
}

func TestUsersUsecase_UpdateProfile_InvalidDisplayName(t *testing.T) {
	uc, _, _ := newUsersUC(t)
	_, err := uc.UpdateProfile(context.Background(), &Viewer{UserID: "u1"},
		&UpdateProfileInput{DisplayName: strPtr("")})
	assert.ErrorIs(t, err, ErrInvalidData)
}

func TestUsersUsecase_UpdateProfile_InvalidBio(t *testing.T) {
	uc, _, _ := newUsersUC(t)
	long := make([]byte, maxBioLen+1)
	for i := range long {
		long[i] = 'a'
	}
	_, err := uc.UpdateProfile(context.Background(), &Viewer{UserID: "u1"},
		&UpdateProfileInput{Bio: strPtr(string(long))})
	assert.ErrorIs(t, err, ErrInvalidData)
}

func TestUsersUsecase_UpdateProfile_InvalidLink(t *testing.T) {
	uc, _, _ := newUsersUC(t)
	_, err := uc.UpdateProfile(context.Background(), &Viewer{UserID: "u1"},
		&UpdateProfileInput{Links: []dtos.Link{{Type: "unknown", URL: "https://bad/"}}})
	assert.ErrorIs(t, err, ErrInvalidData)
}

func TestUsersUsecase_UpdateProfile_RepoError(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	repo.EXPECT().UpdateUser(ctx, mock.Anything).Return(nil, infrastructure.ErrDuplicateUsername)

	_, err := uc.UpdateProfile(ctx, &Viewer{UserID: "u1"},
		&UpdateProfileInput{Username: strPtr("taken")})
	assert.ErrorIs(t, err, ErrUsernameExists)
}

// ---------- RemoveAvatar / RemoveBanner ----------

func TestUsersUsecase_RemoveAvatar_Success(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	updated := &entities.User{ID: "u1", Username: "john"}
	repo.EXPECT().UpdateUser(ctx, mock.MatchedBy(func(p *infrastructure.UpdateUserParams) bool {
		return p.UserID == "u1" && p.RemoveAvatar
	})).Return(updated, nil)

	profile, err := uc.RemoveAvatar(ctx, &Viewer{UserID: "u1"})
	require.NoError(t, err)
	assert.Equal(t, "john", profile.Username)
}

func TestUsersUsecase_RemoveAvatar_RepoError(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)
	repo.EXPECT().UpdateUser(ctx, mock.Anything).Return(nil, infrastructure.ErrUserNotFound)
	_, err := uc.RemoveAvatar(ctx, &Viewer{UserID: "u1"})
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestUsersUsecase_RemoveBanner_Success(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	updated := &entities.User{ID: "u1", Username: "john"}
	repo.EXPECT().UpdateUser(ctx, mock.MatchedBy(func(p *infrastructure.UpdateUserParams) bool {
		return p.UserID == "u1" && p.RemoveBanner
	})).Return(updated, nil)

	_, err := uc.RemoveBanner(ctx, &Viewer{UserID: "u1"})
	require.NoError(t, err)
}

func TestUsersUsecase_RemoveBanner_RepoError(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)
	repo.EXPECT().UpdateUser(ctx, mock.Anything).Return(nil, errors.New("db"))
	_, err := uc.RemoveBanner(ctx, &Viewer{UserID: "u1"})
	assert.ErrorIs(t, err, ErrInternal)
}

// ---------- Follow / Unfollow ----------

func TestUsersUsecase_Follow_Success(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	repo.EXPECT().GetUserIDByUsername(ctx, "alice").Return("alice-id", nil)
	repo.EXPECT().Follow(ctx, "u1", "alice-id").Return(nil)

	err := uc.Follow(ctx, &Viewer{UserID: "u1"}, "alice")
	assert.NoError(t, err)
}

func TestUsersUsecase_Follow_TargetNotFound(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	repo.EXPECT().GetUserIDByUsername(ctx, "ghost").Return("", infrastructure.ErrUserNotFound)

	err := uc.Follow(ctx, &Viewer{UserID: "u1"}, "ghost")
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestUsersUsecase_Follow_CannotFollowSelf(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	repo.EXPECT().GetUserIDByUsername(ctx, "me").Return("u1", nil)

	err := uc.Follow(ctx, &Viewer{UserID: "u1"}, "me")
	assert.ErrorIs(t, err, ErrCannotFollowSelf)
}

func TestUsersUsecase_Follow_RepoError(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	repo.EXPECT().GetUserIDByUsername(ctx, "alice").Return("alice-id", nil)
	repo.EXPECT().Follow(ctx, "u1", "alice-id").Return(errors.New("db"))

	err := uc.Follow(ctx, &Viewer{UserID: "u1"}, "alice")
	assert.ErrorIs(t, err, ErrInternal)
}

func TestUsersUsecase_Unfollow_Success(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	repo.EXPECT().GetUserIDByUsername(ctx, "alice").Return("alice-id", nil)
	repo.EXPECT().Unfollow(ctx, "u1", "alice-id").Return(nil)

	err := uc.Unfollow(ctx, &Viewer{UserID: "u1"}, "alice")
	assert.NoError(t, err)
}

func TestUsersUsecase_Unfollow_TargetNotFound(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	repo.EXPECT().GetUserIDByUsername(ctx, "ghost").Return("", infrastructure.ErrUserNotFound)

	err := uc.Unfollow(ctx, &Viewer{UserID: "u1"}, "ghost")
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestUsersUsecase_Unfollow_RepoError(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	repo.EXPECT().GetUserIDByUsername(ctx, "alice").Return("alice-id", nil)
	repo.EXPECT().Unfollow(ctx, "u1", "alice-id").Return(errors.New("db"))

	err := uc.Unfollow(ctx, &Viewer{UserID: "u1"}, "alice")
	assert.Error(t, err)
}

// ---------- GetBriefs ----------

func TestUsersUsecase_GetBriefs_Empty(t *testing.T) {
	uc, _, _ := newUsersUC(t)
	briefs, err := uc.GetBriefs(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, briefs)
}

func TestUsersUsecase_GetBriefs_DedupesAndDropsEmpty(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	expected := []string{"a", "b", "c"}
	repo.EXPECT().GetBriefsByIDs(ctx, mock.MatchedBy(func(ids []string) bool {
		if len(ids) != len(expected) {
			return false
		}
		for i, v := range expected {
			if ids[i] != v {
				return false
			}
		}
		return true
	})).Return([]dtos.UserBrief{{ID: "a"}}, nil)

	_, err := uc.GetBriefs(ctx, []string{"a", "", "b", "a", "c", ""})
	require.NoError(t, err)
}

func TestUsersUsecase_GetBriefs_TruncatesTo100(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	ids := make([]string, 150)
	for i := range ids {
		ids[i] = string(rune('a'+(i%26))) + string(rune('0'+(i%10))) + string(rune(i)) // unique-ish
	}
	repo.EXPECT().GetBriefsByIDs(ctx, mock.MatchedBy(func(got []string) bool {
		return len(got) <= 100
	})).Return([]dtos.UserBrief{}, nil)

	_, err := uc.GetBriefs(ctx, ids)
	require.NoError(t, err)
}

func TestUsersUsecase_GetBriefs_RepoError(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	repo.EXPECT().GetBriefsByIDs(ctx, mock.Anything).Return(nil, errors.New("db"))
	_, err := uc.GetBriefs(ctx, []string{"a"})
	assert.ErrorIs(t, err, ErrInternal)
}

// ---------- UpdateStats ----------

func TestUsersUsecase_UpdateStats_Empty(t *testing.T) {
	uc, _, _ := newUsersUC(t)
	err := uc.UpdateStats(context.Background(), dtos.PlayedStatName, map[string]int{})
	assert.ErrorIs(t, err, ErrInvalidData)
}

func TestUsersUsecase_UpdateStats_TooMany(t *testing.T) {
	uc, _, _ := newUsersUC(t)
	stats := make(map[string]int, 51)
	for i := 0; i < 51; i++ {
		stats[string(rune('a'+i))+string(rune('0'+(i%10)))] = 1
	}
	err := uc.UpdateStats(context.Background(), dtos.PlayedStatName, stats)
	assert.ErrorIs(t, err, ErrInvalidData)
}

func TestUsersUsecase_UpdateStats_NegativeValue(t *testing.T) {
	uc, _, _ := newUsersUC(t)
	err := uc.UpdateStats(context.Background(), dtos.PlayedStatName, map[string]int{"u1": -1})
	assert.ErrorIs(t, err, ErrInvalidData)
}

func TestUsersUsecase_UpdateStats_Success(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	repo.EXPECT().UpdateStats(ctx, dtos.PlayedStatName, map[string]int{"u1": 2}).Return(nil)

	err := uc.UpdateStats(ctx, dtos.PlayedStatName, map[string]int{"u1": 2})
	assert.NoError(t, err)
}

func TestUsersUsecase_UpdateStats_RepoError(t *testing.T) {
	ctx := context.Background()
	uc, repo, _ := newUsersUC(t)

	repo.EXPECT().UpdateStats(ctx, dtos.PlayedStatName, mock.Anything).Return(errors.New("db"))

	err := uc.UpdateStats(ctx, dtos.PlayedStatName, map[string]int{"u1": 2})
	assert.ErrorIs(t, err, ErrInternal)
}
