package infrastructure

import (
	"github.com/itzLilix/questboard-profile-service/internal/entities"
	"github.com/jackc/pgx/v5"
)

const userCols = "id, username, password_hash, email, created_at, last_login, profile_picture, banner_picture, role, display_name, is_email_verified, sessions_played, sessions_hosted, rating, reviews_count, bio, links"


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
	)
}