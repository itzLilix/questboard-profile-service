package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	OnlyGMs		 bool
    Sort         dtos.UserListSort
	SortOrder	 dtos.SortOrder
    Cursor       string
    Limit        uint64
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
	// Sort-key columns — not rendered, used to build the next cursor.
	CreatedAt  time.Time  `db:"created_at"`
	FollowedAt *time.Time `db:"followed_at"` // f.created_at, NULL when no JOIN match
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

func (r *catalogRepository) GetUsersList(filter *UserCatalogFilter, viewerID string) ([]UserCardRow, string, error) {
	if filter.SortOrder == "" {
		filter.SortOrder = dtos.SortDesc
	}

	cursor, err := decodeCursor(filter.Cursor)
	if err != nil {
		return nil, "", err
	}

	q := r.psql.
		Select(UserCardRowCols...).
		From("users u")

	if viewerID != "" {
		q = q.
			LeftJoin("follows f ON f.followed_id = u.id AND f.follower_id = ?", viewerID).
			Where(sq.NotEq{"u.id": viewerID})
	} else {
		q = q.LeftJoin("follows f ON f.followed_id = u.id")
	}
	if filter.Search != "" {
		q = q.Where(sq.Or{
			sq.ILike{"u.username": "%" + filter.Search + "%"},
			sq.ILike{"u.display_name": "%" + filter.Search + "%"},
		})
	}
	if filter.Format != "" {
		q = q.Where(sq.Or{
			sq.Expr("u.preferred_format IS NULL"),
			sq.Eq{"u.preferred_format": filter.Format},
		})
	}
	if filter.Type != "" {
		q = q.Where(sq.Or{
			sq.Expr("u.preferred_type IS NULL"),
			sq.Eq{"u.preferred_type": filter.Type},
		})
	}
	if filter.City != "" {
		q = q.Where(sq.Expr("lower(u.city) = lower(?)", filter.City))
	}
	if filter.MinRating > 0 {
		q = q.Where(sq.GtOrEq{"u.rating": filter.MinRating})
	}
	if filter.FollowedByID != "" {
		q = q.Where(sq.Eq{"f.follower_id": filter.FollowedByID})
	}
	if filter.FollowedByID == "" {
		q = q.Where(sq.Eq{"u.is_visible_in_catalog": true})
	}
	if filter.OnlyGMs {
		q = q.Where(sq.Gt{"u.sessions_hosted": 0})
	}

	q, err = applyCursor(q, cursor, filter.Sort, filter.SortOrder)
	if err != nil {
		return nil, "", err
	}

	switch filter.Sort {
		case dtos.SortRating:
			q = q.OrderBy("u.rating "+string(filter.SortOrder), "u.id "+string(filter.SortOrder))
		case dtos.SortRecent:
			q = q.OrderBy("u.created_at "+string(filter.SortOrder), "u.id "+string(filter.SortOrder))
		case dtos.SortFollowedAt:
			q = q.OrderBy("f.created_at "+string(filter.SortOrder), "u.id "+string(filter.SortOrder))
		case dtos.SortReviewsCount:
			q = q.OrderBy("u.reviews_count "+string(filter.SortOrder), "u.id "+string(filter.SortOrder))
		default:
			q = q.OrderBy("u.id " + string(filter.SortOrder))
	}

	q = q.Limit(filter.Limit + 1)

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, "", fmt.Errorf("get users list: build sql: %w", err)
	}

	rows, err := r.db.Query(context.Background(), sql, args...)
	if err != nil {
		return nil, "", fmt.Errorf("get users list: query: %w", err)
	}
	defer rows.Close()

	out := make([]UserCardRow, 0, filter.Limit+1)
	for rows.Next() {
		var row UserCardRow
		if err := scanUserCardRow(rows, &row); err != nil {
			return nil, "", fmt.Errorf("get users list: scan: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("get users list: rows: %w", err)
	}

	hasMore := uint64(len(out)) > filter.Limit
	if hasMore {
		out = out[:filter.Limit]
	}

	var nextCursor string
	if hasMore && len(out) > 0 {
		nextCursor, err = buildNextCursor(out[len(out)-1], filter.Sort, filter.SortOrder)
		if err != nil {
			return nil, "", fmt.Errorf("get users list: build next cursor: %w", err)
		}
	}
	
	return out, nextCursor, nil
}