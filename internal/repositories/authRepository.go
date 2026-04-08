package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/itzLilix/QuestBoard/backend/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type authRepository struct {
	db *pgxpool.Pool
}


func NewAuthRepository(db *pgxpool.Pool) *authRepository {
	return &authRepository{db: db}
}

func (r *authRepository) GetUserByID(id string) (*models.User, error) {
	row := r.db.QueryRow(context.Background(),
		"SELECT * FROM users WHERE id=$1", id)
	user := &models.User{}
	err := models.ScanUser(row, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *authRepository) CreateUser(user *models.User) error {
	row := r.db.QueryRow(context.Background(),
		"INSERT INTO users (username, display_name, password_hash, email, role) VALUES ($1, $2, $3, $4, 'user') RETURNING id, created_at",
		user.Username, user.DisplayName, user.PasswordHash, user.Email)

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

func (r *authRepository) GetUserByEmail(email string) (*models.User, error) {
	row := r.db.QueryRow(context.Background(),
		"SELECT * FROM users WHERE email=$1", email)
	user := &models.User{}
	err := models.ScanUser(row, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *authRepository) SaveRefreshToken(token *models.RefreshToken) error {
	row := r.db.QueryRow(context.Background(),
	"INSERT INTO refresh_tokens (user_id, token_prefix, token_hash, expires_at) VALUES ($1, $2, $3, $4) RETURNING id, created_at",
	token.UserID, token.TokenPrefix, token.TokenHash, token.ExpiresAt)
	err := row.Scan(&token.ID, &token.CreatedAt)
	return err
}

func (r *authRepository) GetRefreshTokenByPrefix(prefix string) (*models.RefreshToken, error) {
	row := r.db.QueryRow(context.Background(),
	"SELECT * FROM refresh_tokens WHERE token_prefix=$1", prefix)
	token := &models.RefreshToken{}
	err := row.Scan(&token.ID, &token.UserID, &token.TokenPrefix, &token.TokenHash, &token.ExpiresAt, &token.CreatedAt)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *authRepository) DeleteRefreshToken(prefix string) error {
	_, err := r.db.Exec(context.Background(), "DELETE FROM refresh_tokens WHERE token_prefix=$1", prefix)
	return err
}

func (r *authRepository) UpdateLastLogin(user *models.User) error {
	row := r.db.QueryRow(context.Background(), "UPDATE users SET last_login = NOW() WHERE id = $1 RETURNING last_login", user.ID)
	err := row.Scan(&user.LastLogin)
	return err
}