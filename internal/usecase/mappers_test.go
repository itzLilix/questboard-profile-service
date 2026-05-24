package usecase

import (
	"testing"
	"time"

	"github.com/itzLilix/questboard-profile-service/internal/entities"
	"github.com/itzLilix/questboard-profile-service/internal/infrastructure"
	"github.com/itzLilix/questboard-shared/dtos"
	"github.com/stretchr/testify/assert"
)

func strPtr(s string) *string { return &s }

func sampleUser() *entities.User {
	avatar := "https://cdn/a.png"
	banner := "https://cdn/b.png"
	bio := "hello"
	return &entities.User{
		ID:              "u1",
		Username:        "john",
		DisplayName:     "John",
		Email:           "john@example.com",
		PasswordHash:    "hash",
		CreatedAt:       time.Unix(1000, 0),
		LastLogin:       time.Unix(2000, 0),
		AvatarURL:       &avatar,
		BannerURL:       &banner,
		Role:            dtos.UserRole,
		IsEmailVerified: true,
		SessionsPlayed:  3,
		SessionsHosted:  2,
		Rating:          4.5,
		ReviewsCount:    10,
		Bio:             &bio,
		Links:           []dtos.Link{{Type: "telegram", URL: "https://t.me/john"}},
	}
}

func TestMapUserToPublicProfile_Nil(t *testing.T) {
	assert.Nil(t, mapUserToPublicProfile(nil))
}

func TestMapUserToPublicProfile_CopiesFields(t *testing.T) {
	u := sampleUser()
	p := mapUserToPublicProfile(u)
	assert.Equal(t, u.ID, p.ID)
	assert.Equal(t, u.Username, p.Username)
	assert.Equal(t, u.DisplayName, p.DisplayName)
	assert.Equal(t, u.AvatarURL, p.AvatarURL)
	assert.Equal(t, u.BannerURL, p.BannerURL)
	assert.Equal(t, u.SessionsPlayed, p.SessionsPlayed)
	assert.Equal(t, u.SessionsHosted, p.SessionsHosted)
	assert.Equal(t, u.Rating, p.Rating)
	assert.Equal(t, u.ReviewsCount, p.ReviewsCount)
	assert.Equal(t, u.Bio, p.Bio)
	assert.Equal(t, u.Links, p.Links)
	assert.False(t, p.IsFollowed)
}

func TestMapUserToPrivateProfile_Nil(t *testing.T) {
	assert.Nil(t, mapUserToPrivateProfile(nil))
}

func TestMapUserToPrivateProfile_CopiesFields(t *testing.T) {
	u := sampleUser()
	p := mapUserToPrivateProfile(u)
	assert.Equal(t, u.Email, p.Email)
	assert.Equal(t, u.CreatedAt, p.CreatedAt)
	assert.Equal(t, u.LastLogin, p.LastLogin)
	assert.Equal(t, u.Role, p.Role)
	assert.Equal(t, u.IsEmailVerified, p.IsEmailVerified)
	// Embedded public
	assert.Equal(t, u.ID, p.ID)
	assert.Equal(t, u.Username, p.Username)
}

func TestMapUpdateInputToRepoParams(t *testing.T) {
	input := &UpdateProfileInput{
		Username:    strPtr("new"),
		DisplayName: strPtr("New"),
		Bio:         strPtr("bio"),
		Links:       []dtos.Link{{Type: "x", URL: "https://x.com/foo"}},
	}
	out := mapUpdateInputToRepoParams("u1", input)
	assert.Equal(t, "u1", out.UserID)
	assert.Equal(t, input.Username, out.Username)
	assert.Equal(t, input.DisplayName, out.DisplayName)
	assert.Equal(t, input.Bio, out.Bio)
	assert.Equal(t, input.Links, out.Links)
}

func TestMapListUsersFilterToUserCatalogFilter(t *testing.T) {
	filter := ListUsersFilter{
		Search:    "rogue",
		Format:    dtos.Online,
		Type:      dtos.CampaignType,
		City:      "NYC",
		MinRating: 4.0,
		OnlyGMs:   true,
		Sort:      dtos.SortRating,
		SortOrder: dtos.SortDesc,
		Cursor:    "abc",
		Limit:     25,
	}
	out := mapListUsersFilterToUserCatalogFilter(filter, "fid")
	assert.Equal(t, filter.Search, out.Search)
	assert.Equal(t, filter.Format, out.Format)
	assert.Equal(t, filter.Type, out.Type)
	assert.Equal(t, filter.City, out.City)
	assert.Equal(t, filter.MinRating, out.MinRating)
	assert.Equal(t, "fid", out.FollowedByID)
	assert.Equal(t, filter.OnlyGMs, out.OnlyGMs)
	assert.Equal(t, filter.Sort, out.Sort)
	assert.Equal(t, filter.SortOrder, out.SortOrder)
	assert.Equal(t, filter.Cursor, out.Cursor)
	assert.Equal(t, filter.Limit, out.Limit)
}

func TestMapUserCardRowToProfileCardData_Nil(t *testing.T) {
	assert.Nil(t, mapUserCardRowToProfileCardData(nil))
}

func TestMapUserCardRowToProfileCardData_Maps(t *testing.T) {
	avatar := "a"
	banner := "b"
	format := dtos.Online
	typ := dtos.CampaignType
	rows := []infrastructure.UserCardRow{
		{
			ID: "u1", Username: "john", DisplayName: "John",
			AvatarURL: &avatar, BannerURL: &banner,
			Rating: 4.5, ReviewsCount: 5,
			SessionsPlayed: 2, SessionsHosted: 3,
			PreferredFormat: &format, PreferredType: &typ,
			IsFollowed: true,
		},
		{ID: "u2", Username: "jane", DisplayName: "Jane"},
	}
	out := mapUserCardRowToProfileCardData(rows)
	assert.Len(t, out, 2)
	assert.Equal(t, "u1", out[0].ID)
	assert.True(t, out[0].IsFollowed)
	assert.Equal(t, &format, out[0].PreferredFormat)
	assert.Equal(t, &typ, out[0].PreferredType)
	assert.Equal(t, "u2", out[1].ID)
	assert.False(t, out[1].IsFollowed)
}
