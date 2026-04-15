package repositories

import (
	"github.com/itzLilix/questboard-profile-service/internal/entities"
	"github.com/jackc/pgx/v5"
)

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