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

type authRepository struct {
	db *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewAuthRepository(db *pgxpool.Pool, psql sq.StatementBuilderType) *authRepository {
	return &authRepository{db: db, psql: psql}
}

func (r *authRepository) GetUserByID(id string) (*entities.User, error) {
	sql, args, err := r.psql.Select(userCols).From("users").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	row := r.db.QueryRow(context.Background(), sql, args...)
	user := &entities.User{}
	if err := scanUser(row, user); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *authRepository) CreateUser(user *entities.User) error {
	sql, args, err := r.psql.Insert("users").
		Columns("username", "display_name", "password_hash", "email", "role", "last_login").
		Values(user.Username, user.DisplayName, user.PasswordHash, user.Email, dtos.UserRole, sq.Expr("NOW()")).
		Suffix("RETURNING id, created_at, last_login").
		ToSql()
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	if err := r.db.QueryRow(context.Background(), sql, args...).Scan(&user.ID, &user.CreatedAt, &user.LastLogin); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "username") {
				return ErrDuplicateUsername
			}
			if strings.Contains(pgErr.ConstraintName, "email") {
				return ErrDuplicateEmail
			}
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *authRepository) GetUserByEmail(email string) (*entities.User, error) {
	sql, args, err := r.psql.Select(userCols).From("users").Where(sq.Eq{"email": email}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	row := r.db.QueryRow(context.Background(), sql, args...)
	user := &entities.User{}
	if err := scanUser(row, user); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *authRepository) SaveRefreshToken(token *entities.RefreshToken) error {
	sql, args, err := r.psql.Insert("refresh_tokens").
		Columns("user_id", "token_prefix", "token_hash", "expires_at").
		Values(token.UserID, token.TokenPrefix, token.TokenHash, token.ExpiresAt).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	if err := r.db.QueryRow(context.Background(), sql, args...).Scan(&token.ID, &token.CreatedAt); err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}

func (r *authRepository) GetRefreshTokenByPrefix(prefix string) (*entities.RefreshToken, error) {
	sql, args, err := r.psql.Select("id, user_id, token_prefix, token_hash, expires_at, created_at").
		From("refresh_tokens").
		Where(sq.Eq{"token_prefix": prefix}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("get refresh token by prefix: %w", err)
	}
	token := &entities.RefreshToken{}
	if err := r.db.QueryRow(context.Background(), sql, args...).Scan(
		&token.ID, &token.UserID, &token.TokenPrefix, &token.TokenHash, &token.ExpiresAt, &token.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("get refresh token by prefix: %w", err)
	}
	return token, nil
}

func (r *authRepository) DeleteRefreshToken(prefix string) error {
	sql, args, err := r.psql.Delete("refresh_tokens").Where(sq.Eq{"token_prefix": prefix}).ToSql()
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	_, err = r.db.Exec(context.Background(), sql, args...)
	return err
}

func (r *authRepository) UpdateLastLogin(user *entities.User) error {
	sql, args, err := r.psql.Update("users").
		Set("last_login", sq.Expr("NOW()")).
		Where(sq.Eq{"id": user.ID}).
		Suffix("RETURNING last_login").
		ToSql()
	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}
	return r.db.QueryRow(context.Background(), sql, args...).Scan(&user.LastLogin)
}
