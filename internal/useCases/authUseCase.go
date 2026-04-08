package useCases

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
	"github.com/itzLilix/QuestBoard/backend/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

// type HashProvider interface{
// 	Hash(any) string
// }

// type TokenProvider interface{
// 	GenerateToken(user *models.User) (string, error)
// }

type AuthUseCase interface {
	Register(username, displayname, email, password string) (*models.User, string, string, error)
	Login(username, password string) (*models.User, string, string, error)
	Logout(refreshToken string) error
	ValidateToken(tokenString string) (*models.User, error)
	RefreshTokens(refreshToken string) (*models.User, string, string, error)
}

type authUseCase struct {
	repo AuthRepository
}

type claims struct {
	User *models.User `json:"user"`
	jwt.RegisteredClaims
}

func NewAuthUseCase(repo AuthRepository) AuthUseCase {
	return &authUseCase{repo: repo}
}

func (s *authUseCase) ValidateToken(tokenString string) (*models.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if c, ok := token.Claims.(*claims); ok && c.User != nil {
		return c.User, nil
	}

	return nil, fmt.Errorf("invalid claims")
}

func (s *authUseCase) Register(username, displayname, email, password string) (*models.User, string, string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", err
	}

	user := &models.User{
		Username:     username,
		DisplayName: displayname,
		Email:        email,
		PasswordHash: string(passwordHash),
	}
	err = s.repo.CreateUser(user)
	if err != nil {
		if errors.Is(err, repositories.ErrDuplicateEmail) {
			return nil, "", "", ErrEmailExists
		} else if errors.Is(err, repositories.ErrDuplicateUsername) {
			return nil, "", "", ErrUsernameExists
		}
		return nil, "", "", fmt.Errorf("register: %w", err)
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

func (s *authUseCase) Login(email, password string) (*models.User, string, string, error) {
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

func (s *authUseCase) Logout(refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	prefix := refreshToken[:8]
	if err := s.repo.DeleteRefreshToken(prefix); err != nil {
		return err
	}

	return nil
}

func (s *authUseCase) generateAccessToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)
	secretKey := []byte(os.Getenv("JWT_SECRET"))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	})

	return token.SignedString(secretKey)
}

func (s *authUseCase) generateRefreshToken(user *models.User) (string, error) {
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

func (s *authUseCase) RefreshTokens(clientToken string) (*models.User, string, string, error) {
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