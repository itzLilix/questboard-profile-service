package useCases

import (
	"errors"
	"fmt"
	"time"

	"github.com/itzLilix/QuestBoard/backend/internal/models"
	"github.com/itzLilix/QuestBoard/backend/internal/repositories"
)

type AuthUseCase interface {
	Register(username, displayname, email, password string) (*models.User, string, string, error)
	Login(username, password string) (*models.User, string, string, error)
	Logout(refreshToken string) error
	ValidateToken(tokenString string) (*models.User, error)
	RefreshTokens(refreshToken string) (*models.User, string, string, error)
}

type authUseCase struct {
	repo AuthRepository
	tokenProvider TokenProvider
	passwordHasher PasswordHasher
}

func NewAuthUseCase(repo AuthRepository, tokenProvider TokenProvider, passwordHasher PasswordHasher) AuthUseCase {
	return &authUseCase{repo: repo, tokenProvider: tokenProvider, passwordHasher: passwordHasher}
}

func (s *authUseCase) ValidateToken(tokenString string) (*models.User, error) {
	claims, err := s.tokenProvider.ParseToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("validateToken: %w", err)
	}

	user, err := s.repo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("validateToken: %w", err)
	}

	return user, nil
}

func (s *authUseCase) Register(username, displayname, email, password string) (*models.User, string, string, error) {
	passwordHash, err := s.passwordHasher.HashPassword(password)
	if err != nil {
		return nil, "", "", fmt.Errorf("register: %w", err)
	}

	user := &models.User{
		Username:     username,
		DisplayName: displayname,
		Email:        email,
		PasswordHash: passwordHash,
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
		return nil, "", "", fmt.Errorf("register: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("register: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

func (s *authUseCase) Login(email, password string) (*models.User, string, string, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, "", "", ErrUserNotFound
	}

	err = s.passwordHasher.CompareHashAndPassword(user.PasswordHash, password)
	if err != nil {
		return nil, "", "", ErrWrongPassword
	}

	_ = s.repo.UpdateLastLogin(user)

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("login: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("login: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

func (s *authUseCase) Logout(refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	prefix := refreshToken[:8]
	if err := s.repo.DeleteRefreshToken(prefix); err != nil {
		return fmt.Errorf("logout: %w", err)
	}

	return nil
}

func (s *authUseCase) generateAccessToken(user *models.User) (string, error) {
	token, err := s.tokenProvider.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", fmt.Errorf("generateAccessToken: %w", err)
	}

	return token, nil
}

func (s *authUseCase) generateRefreshToken(user *models.User) (string, error) {
	tokenString, hashString, err := s.tokenProvider.GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generateRefreshToken: %w", err)
	}

	prefix := tokenString[:8]
	

	token := &models.RefreshToken{
		UserID:      user.ID,
		TokenPrefix: string(prefix),
		TokenHash:   hashString,
		ExpiresAt:   time.Now().AddDate(0, 0, 30),
	}
	err = s.repo.SaveRefreshToken(token)
	if err != nil {
		return "", fmt.Errorf("generateRefreshToken: %w", err)
	}

	return tokenString, nil
}

func (s *authUseCase) RefreshTokens(clientToken string) (*models.User, string, string, error) {
	prefix := clientToken[:8]
	storedToken, err := s.repo.GetRefreshTokenByPrefix(prefix)
	if err != nil {
		return nil, "", "", ErrInvalidToken
	}

	if !s.tokenProvider.IsRefreshTokenValid(clientToken, storedToken.TokenHash) || 
		storedToken.ExpiresAt.Before(time.Now()) {
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