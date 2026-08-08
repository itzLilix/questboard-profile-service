package infrastructure

import (
	"github.com/itzLilix/questboard-profile-service/internal/entities"
	"github.com/jackc/pgx/v5"
)

var userCols = []string {
	"id", 
	"username", 
	"password_hash", 
	"email", 
	"created_at", 
	"last_login", 
	"profile_picture", 
	"banner_picture", 
	"role", 
	"display_name", 
	"is_email_verified", 
	"sessions_played", 
	"sessions_hosted", 
	"rating", 
	"reviews_count", 
	"bio", 
	"links", 
	"preferred_type", 
	"preferred_format", 
	"city", 
	"is_visible_in_catalog",
}


func scanUser(row pgx.Row, user *entities.User) error {
	return row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Email,
		&user.CreatedAt,
		&user.LastLogin,
		&user.AvatarURL,
		&user.BannerURL,
		&user.Role,
		&user.DisplayName,
		&user.IsEmailVerified,
		&user.SessionsPlayed,
		&user.SessionsHosted,
		&user.Rating,
		&user.ReviewsCount,
		&user.Bio,
		&user.Links,
		&user.PreferredType,
		&user.PreferredFormat,
		&user.City,
		&user.IsVisibleInCatalog,
	)
}

// userCardBaseCols are the static columns of a catalog row. The two follow-edge
// columns (is_followed, followed_at) depend on which joins the query has, so the
// repository appends them per query; scanUserCardRow expects them in the order
// base..., is_followed, u.created_at, followed_at.
var userCardBaseCols = []string {
	"u.id",
	"u.username",
	"u.display_name",
	"u.profile_picture",
	"u.banner_picture",
	"u.rating",
	"u.reviews_count",
	"u.sessions_played",
	"u.sessions_hosted",
	"u.preferred_format",
	"u.preferred_type",
}

func scanUserCardRow(row pgx.Row, user *UserCardRow) error {
	return row.Scan(
		&user.ID,
		&user.Username,
		&user.DisplayName,
		&user.AvatarURL,
		&user.BannerURL,
		&user.Rating,
		&user.ReviewsCount,
		&user.SessionsPlayed,
		&user.SessionsHosted,
		&user.PreferredFormat,
		&user.PreferredType,
		&user.IsFollowed,
		&user.CreatedAt,
		&user.FollowedAt,
	)
}