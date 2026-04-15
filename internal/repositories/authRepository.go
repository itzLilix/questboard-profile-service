package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/itzLilix/questboard-profile-service/internal/entities"
	dtos "github.com/itzLilix/questboard-shared/DTOs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type authRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *authRepository {
	return &authRepository{db: db}
}

func (r *authRepository) GetUserByID(id string) (*entities.User, error) {
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

func (r *authRepository) CreateUser(user *entities.User) error {
	row := r.db.QueryRow(context.Background(),
		"INSERT INTO users (username, display_name, password_hash, email, role) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at",
		user.Username, user.DisplayName, user.PasswordHash, user.Email, dtos.UserRole)

	err := row.Scan(&user.ID, &user.CreatedAt)
	if err != nil {
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
	row := r.db.QueryRow(context.Background(),
		"SELECT * FROM users WHERE email=$1", email)
	user := &entities.User{}
	err := scanUser(row, user)

    if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
        return nil, fmt.Errorf("get user by email: %w", err)
    }
	return user, nil
}

func (r *authRepository) SaveRefreshToken(token *entities.RefreshToken) error {
	row := r.db.QueryRow(context.Background(),
	"INSERT INTO refresh_tokens (user_id, token_prefix, token_hash, expires_at) VALUES ($1, $2, $3, $4) RETURNING id, created_at",
	token.UserID, token.TokenPrefix, token.TokenHash, token.ExpiresAt)
	err := row.Scan(&token.ID, &token.CreatedAt)
	if err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}

func (r *authRepository) GetRefreshTokenByPrefix(prefix string) (*entities.RefreshToken, error) {
	row := r.db.QueryRow(context.Background(),
	"SELECT * FROM refresh_tokens WHERE token_prefix=$1", prefix)
	token := &entities.RefreshToken{}
	err := row.Scan(&token.ID, &token.UserID, &token.TokenPrefix, &token.TokenHash, &token.ExpiresAt, &token.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("get refresh token by prefix: %w", err)
	}
	return token, nil
}

func (r *authRepository) DeleteRefreshToken(prefix string) error {
	_, err := r.db.Exec(context.Background(), "DELETE FROM refresh_tokens WHERE token_prefix=$1", prefix)
	return err
}

func (r *authRepository) UpdateLastLogin(user *entities.User) error {
	row := r.db.QueryRow(context.Background(), "UPDATE users SET last_login = NOW() WHERE id = $1 RETURNING last_login", user.ID)
	err := row.Scan(&user.LastLogin)
	return err
}