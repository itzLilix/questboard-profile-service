package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/itzLilix/questboard-profile-service/internal/entities"
	"github.com/itzLilix/questboard-shared/dtos"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type usersRepository struct {
	db *pgxpool.Pool
	psql sq.StatementBuilderType
}

type UpdateUserParams struct {
	UserID       string     `db:"id"`
	Username     *string    `db:"username"`
	DisplayName  *string    `db:"display_name"`
	AvatarURL    *string    `db:"avatar_url"`
	RemoveAvatar bool       `db:"-"`
	BannerURL    *string    `db:"banner_url"`
	RemoveBanner bool       `db:"-"`
	Bio          *string    `db:"bio"`
	Links        []dtos.Link `db:"links"`
}

func NewUsersRepository(db *pgxpool.Pool, psql sq.StatementBuilderType) *usersRepository {
	return &usersRepository{db: db, psql: psql}
}

func (r *usersRepository) GetUserByUsername(ctx context.Context, username string) (*entities.User, error) {
	sql, args, err := r.psql.Select(userCols...).From("users").Where(sq.Eq{"username": username}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	row := r.db.QueryRow(ctx, sql, args...)
	user := &entities.User{}
	if err := scanUser(row, user); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return user, nil
}

func (r *usersRepository) GetUserByID(ctx context.Context, id string) (*entities.User, error) {
	sql, args, err := r.psql.Select(userCols...).From("users").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	row := r.db.QueryRow(ctx, sql, args...)
	user := &entities.User{}
	if err := scanUser(row, user); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *usersRepository) GetUserIDByUsername(ctx context.Context, username string) (string, error) {
	sql, args, err := r.psql.Select("id").From("users").Where(sq.Eq{"username": username}).ToSql()
	if err != nil {
		return "", fmt.Errorf("get user id by username: %w", err)
	}
	var id string
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", fmt.Errorf("get user id by username: %w", err)
	}
	return id, nil
}

func (r *usersRepository) UpdateUser(ctx context.Context, input *UpdateUserParams) (*entities.User, error) {
	setCount := 0
	builder := r.psql.Update("users")
	if input.Username != nil {
		builder = builder.Set("username", *input.Username)
		setCount++
	}
	if input.DisplayName != nil {
		builder = builder.Set("display_name", *input.DisplayName)
		setCount++
	}
	if input.AvatarURL != nil && !input.RemoveAvatar {
		builder = builder.Set("profile_picture", *input.AvatarURL)
		setCount++
	}
	if input.RemoveAvatar {
		builder = builder.Set("profile_picture", nil)
		setCount++
	}
	if input.BannerURL != nil && !input.RemoveBanner {
		builder = builder.Set("banner_picture", *input.BannerURL)
		setCount++
	}
	if input.RemoveBanner {
		builder = builder.Set("banner_picture", nil)
		setCount++
	}
	if input.Bio != nil {
		builder = builder.Set("bio", *input.Bio)
		setCount++
	}
	if input.Links != nil {
		builder = builder.Set("links", input.Links)
		setCount++
	}

	if setCount == 0 {
		return nil, ErrNoNewData
	}

	sql, args, err := builder.Where(sq.Eq{"id": input.UserID}).Suffix("RETURNING " + strings.Join(userCols, ", ")).ToSql()
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	user := &entities.User{}
	if err := scanUser(row, user); err != nil {
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

func (r *usersRepository) Follow(ctx context.Context, followerID, followedID string) error {
	sql, args, err := r.psql.Insert("follows").
		Columns("follower_id", "followed_id").
		Values(followerID, followedID).
		Suffix("ON CONFLICT DO NOTHING").
		ToSql()
	if err != nil {
		return fmt.Errorf("follow user: %w", err)
	}
	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23514" {
			return ErrCannotFollowSelf
		}
		return fmt.Errorf("follow user: %w", err)
	}
	return nil
}

func (r *usersRepository) Unfollow(ctx context.Context, followerID, followedID string) error {
	sql, args, err := r.psql.Delete("follows").
		Where(sq.And{sq.Eq{"follower_id": followerID}, sq.Eq{"followed_id": followedID}}).
		ToSql()
	if err != nil {
		return fmt.Errorf("unfollow user: %w", err)
	}
	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("unfollow user: %w", err)
	}
	return nil
}

func (r *usersRepository) IsFollowing(ctx context.Context, followerID, followedID string) (bool, error) {
	subSql, args, err := r.psql.Select("1").From("follows").
		Where(sq.Eq{"follower_id": followerID}).
		Where(sq.Eq{"followed_id": followedID}).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("is following: %w", err)
	}
	var exists bool
	if err := r.db.QueryRow(ctx, "SELECT EXISTS("+subSql+")", args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("is following: %w", err)
	}
	return exists, nil
}

func (r *usersRepository) GetBriefsByIDs(ctx context.Context, ids []string) ([]dtos.UserBrief, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	sql, args, err := r.psql.Select("id", "username", "display_name", "profile_picture", "rating", "sessions_played", "sessions_hosted").
		From("users").
		Where(sq.Eq{"id": ids}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("get briefs by ids: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get briefs by ids: %w", err)
	}
	defer rows.Close()

	briefs := []dtos.UserBrief{}
	for rows.Next() {
		var brief dtos.UserBrief
		if err := rows.Scan(&brief.ID, &brief.Username, &brief.DisplayName, &brief.AvatarURL, &brief.Rating, &brief.Played, &brief.Hosted); err != nil {
			return nil, fmt.Errorf("get briefs by ids: %w", err)
		}
		fmt.Println(brief.Rating)
		briefs = append(briefs, brief)
	}
	return briefs, nil
}

func (r *usersRepository) UpdateStats(ctx context.Context, statName dtos.UserStatName, values map[string]int) error {
	var statCol string
	switch statName {
		case dtos.HostedStatName: statCol = "sessions_hosted"
		case dtos.PlayedStatName: statCol = "sessions_played"
		default: return fmt.Errorf("invalid statname")
	} 

	tuples := make([]string, 0, len(values))
    tArgs := make([]any, 0, len(values)*2)
    for id, v := range values {
        tuples = append(tuples, "(?::uuid, ?::int)")
        tArgs = append(tArgs, id, v)
    }

	sql, args, err := r.psql.
        Update("users AS u").
        Prefix("WITH v(id, val) AS (VALUES "+strings.Join(tuples, ",")+")", tArgs...).
        Set(statCol, sq.Expr("v.val")).
        From("v").
        Where("u.id = v.id").
        ToSql()
    if err != nil {
        return fmt.Errorf("update stats: %w", err)
    }

    if _, err := r.db.Exec(ctx, sql, args...); err != nil {
        return fmt.Errorf("update stats: %w", err)
    }
    return nil
}