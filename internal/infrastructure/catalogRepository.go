package infrastructure

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/itzLilix/questboard-shared/dtos"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserCatalogFilter struct {
    Search       string
    Format       dtos.SessionFormat
    Type         dtos.SessionType
	City		 string
    MinRating    float64
    FollowedByID string
    Sort         dtos.UserListSort
	OnlyGMs		 bool
    Cursor       string
    Limit        int
}

type UserCardRow struct {
	ID              string              `db:"id"`
	Username        string              `db:"username"`
	DisplayName     string              `db:"display_name"`
	AvatarURL       *string             `db:"avatar_url"`
	BannerURL       *string             `db:"banner_url"`
	Rating          float64             `db:"rating"`
	ReviewsCount    int                 `db:"reviews_count"`
	SessionsPlayed  int                 `db:"sessions_played"`
	SessionsHosted  int                 `db:"sessions_hosted"`
	PreferredFormat *dtos.SessionFormat `db:"preferred_format"`
	PreferredType   *dtos.SessionType   `db:"preferred_type"`
	IsFollowed      bool                `db:"is_followed"`
}

type catalogRepository struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewCatalogRepository (db *pgxpool.Pool, psql sq.StatementBuilderType) *catalogRepository {
	return &catalogRepository{db:db, psql: psql}
}

func (r *catalogRepository) GetUserIDByUsername(username string) (string, error) {
	sql, args, err := r.psql.Select("id").From("users").Where(sq.Eq{"username": username}).ToSql()
	if err != nil {
		return "", fmt.Errorf("get user id by username: %w", err)
	}
	var id string
	if err := r.db.QueryRow(context.Background(), sql, args...).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", fmt.Errorf("get user id by username: %w", err)
	}
	return id, nil
}

func (r *catalogRepository) GetUsersList(filter *UserCatalogFilter) ([]UserCardRow, error) {
	
}