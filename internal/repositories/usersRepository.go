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
    UserID      string
    Username    *string
    DisplayName *string
    AvatarURL   *string
    BannerURL   *string
    Bio         *string
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
	if input.AvatarURL != nil {
		qParts = append(qParts, fmt.Sprintf("avatar_url=$%d", len(args)+1))
		args = append(args, *input.AvatarURL)
	}
	if input.BannerURL != nil {
		qParts = append(qParts, fmt.Sprintf("banner_url=$%d", len(args)+1))
		args = append(args, *input.BannerURL)
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