package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/itzLilix/QuestBoard/backend/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(username, email, password string) (*models.User, string, string, error)
	Login(username, password string) (*models.User, string, string, error)
	Logout(refreshToken string) error
	ValidateToken(tokenString string) (*models.User, error)
	RefreshTokens(refreshToken string) (*models.User, string, string, error)
}

type service struct {
	repo Repository
}

type claims struct {
	ID       string
	Username string
	Role     string
	jwt.RegisteredClaims
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ValidateToken(tokenString string) (*models.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	var user *models.User

	if err != nil {
		return nil, ErrInvalidToken
	} else if claims, ok := token.Claims.(*claims); ok {
		user, err = s.repo.GetUserByID(claims.ID)
		if err != nil {
			return nil, ErrUserNotFound
		}
	} else {
		return nil, fmt.Errorf("unknown claims type")
	}
	return user, nil
}

func (s *service) Register(username, email, password string) (*models.User, string, string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", err
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
	}
	err = s.repo.CreateUser(user)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "username") {
				return nil, "", "", ErrUsernameExists
			}
			if strings.Contains(pgErr.ConstraintName, "email") {
				return nil, "", "", ErrEmailExists
			}
		}
		return nil, "", "", err
	}

	s.repo.UpdateLastLogin(user)

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *service) Login(email, password string) (*models.User, string, string, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, "", "", ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, "", "", ErrWrongPassword
	}

	s.repo.UpdateLastLogin(user)

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *service) Logout(refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	prefix := refreshToken[:8]
	if err := s.repo.DeleteRefreshToken(prefix); err != nil {
		return err
	}

	return nil
}

func (s *service) generateAccessToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(15 * time.Second)
	secretKey := []byte(os.Getenv("JWT_SECRET"))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	})

	return token.SignedString(secretKey)
}

func (s *service) generateRefreshToken(user *models.User) (string, error) {
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	tokenString := hex.EncodeToString(tokenBytes)

	prefix := tokenString[:8]
	hash := sha256.Sum256([]byte(tokenString))
	hashString := hex.EncodeToString(hash[:])

	token := &models.RefreshToken{
		UserID:      user.ID,
		TokenPrefix: string(prefix),
		TokenHash:   hashString,
		ExpiresAt:   time.Now().AddDate(0, 0, 30),
	}
	err := s.repo.SaveRefreshToken(token)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *service) RefreshTokens(clientToken string) (*models.User, string, string, error) {
	prefix := clientToken[:8]
	storedToken, err := s.repo.GetRefreshTokenByPrefix(prefix)
	if err != nil {
		return nil, "", "", ErrInvalidToken
	}

	clientHashBytes := sha256.Sum256([]byte(clientToken))
	clientHash := hex.EncodeToString(clientHashBytes[:])

	if !strings.EqualFold(storedToken.TokenHash, clientHash) {
		return nil, "", "", ErrInvalidToken
	}
	if storedToken.ExpiresAt.Before(time.Now()){
		return nil, "", "", ErrInvalidToken
	}
	
	if err := s.repo.DeleteRefreshToken(prefix); err != nil {
		return nil, "", "", err
	}

	user, err := s.repo.GetUserByID(storedToken.UserID)
	if err != nil {
		return nil, "", "", ErrUserNotFound
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}