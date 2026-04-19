package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/itzLilix/questboard-profile-service/internal/entities"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type usersRepository struct {
	db *pgxpool.Pool
}

type UpdateUserParams struct {
    UserID       string
    Username     *string
    DisplayName  *string
    AvatarURL    *string
	RemoveAvatar bool
    BannerURL    *string
	RemoveBanner bool
    Bio          *string
}


func NewUsersRepository(db *pgxpool.Pool) *usersRepository {
	return &usersRepository{db: db}
}

func (r *usersRepository) GetUserByUsername(username string) (*entities.User, error) {
	row := r.db.QueryRow(context.Background(),
		"SELECT * FROM users WHERE username=$1", username)
	user := &entities.User{}
	err := scanUser(row, user)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
        return nil, fmt.Errorf("get user by username: %w", err)
	}
	return user, nil
}

func (r *usersRepository) GetUserByID(id string) (*entities.User, error) {
	row := r.db.QueryRow(context.Background(),
		"SELECT * FROM users WHERE id=$1", id)
	user := &entities.User{}
	err := scanUser(row, user)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
        return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *usersRepository) GetUserIDByUsername(username string) (string, error) {
	row := r.db.QueryRow(context.Background(),
		"SELECT id FROM users WHERE username=$1", username)
	var id string
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", fmt.Errorf("get user id by username: %w", err)
	}
	return id, nil
}

func (r *usersRepository) UpdateUser(input *UpdateUserParams) (*entities.User, error) {
	q := `UPDATE users SET `
	qParts := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)

	if input.Username != nil {
		qParts = append(qParts, fmt.Sprintf("username=$%d", len(args)+1))
		args = append(args, *input.Username)
	}
	if input.DisplayName != nil {
		qParts = append(qParts, fmt.Sprintf("display_name=$%d", len(args)+1))
		args = append(args, *input.DisplayName)
	}
	if input.AvatarURL != nil && !input.RemoveAvatar {
		qParts = append(qParts, fmt.Sprintf("avatar_url=$%d", len(args)+1))
		args = append(args, *input.AvatarURL)
	}
	if input.RemoveAvatar {
		qParts = append(qParts, "avatar_url=NULL")
	}
	if input.BannerURL != nil && !input.RemoveBanner {
		qParts = append(qParts, fmt.Sprintf("banner_url=$%d", len(args)+1))
		args = append(args, *input.BannerURL)
	}
	if input.RemoveBanner {
		qParts = append(qParts, "banner_url=NULL")
	}
	if input.Bio != nil {
		qParts = append(qParts, fmt.Sprintf("bio=$%d", len(args)+1))
		args = append(args, *input.Bio)
	}

	if len(qParts) == 0 {
		return nil, nil
	}

	q += strings.Join(qParts, ", ") + ` WHERE id = $` + fmt.Sprint(len(args)+1) + ` RETURNING *`
	args = append(args, input.UserID)

    row := r.db.QueryRow(context.Background(), q, args...)
	user := &entities.User{}
	err := scanUser(row, user)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "username") {
				return nil, ErrDuplicateUsername
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

func (r *usersRepository) Follow(followerID, followedID string) error {
    _, err := r.db.Exec(context.Background(),
        `INSERT INTO follows (follower_id, followed_id) VALUES ($1, $2)
         ON CONFLICT DO NOTHING`,
        followerID, followedID)
    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23514" {
            return ErrCannotFollowSelf
        }
        return fmt.Errorf("follow user: %w", err)
    }
    return nil
}

func (r *usersRepository) Unfollow(followerID, followedID string) error {
    _, err := r.db.Exec(context.Background(),
        `DELETE FROM follows WHERE follower_id=$1 AND followed_id=$2`,
        followerID, followedID)
    if err != nil {
        return fmt.Errorf("unfollow user: %w", err)
    }
    return nil
}

func (r *usersRepository) IsFollowing(followerID, followedID string) (bool, error) {
    var exists bool
    err := r.db.QueryRow(context.Background(),
        `SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id=$1 AND followed_id=$2)`,
        followerID, followedID).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("is following: %w", err)
    }
    return exists, nil
}