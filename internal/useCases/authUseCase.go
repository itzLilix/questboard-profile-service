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

const (
	refreshTokenPrefixLength = 8
)

func NewAuthUseCase(repo AuthRepository, tokenProvider TokenProvider, passwordHasher PasswordHasher) AuthUseCase {
	return &authUseCase{repo: repo, tokenProvider: tokenProvider, passwordHasher: passwordHasher}
}

func (s *authUseCase) ValidateToken(tokenString string) (*models.User, error) {
	claims, err := s.tokenProvider.ParseToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("validateToken: parse token: %w", err)
	}

	user, err := s.repo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("validateToken: get user: %w", err)
	}

	return user, nil
}

func (s *authUseCase) Register(username, displayname, email, password string) (*models.User, string, string, error) {
	passwordHash, err := s.passwordHasher.HashPassword(password)
	if err != nil {
		return nil, "", "", fmt.Errorf("register: hash password: %w", err)
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
		return nil, "", "", fmt.Errorf("register: create user: %w", err)
	}

	_ = s.repo.UpdateLastLogin(user)

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("register: generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("register: generate refresh token: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

func (s *authUseCase) Login(email, password string) (*models.User, string, string, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, "", "", ErrUserNotFound
		}
		return nil, "", "", fmt.Errorf("login: get user: %w", err)
	}

	err = s.passwordHasher.CompareHashAndPassword(user.PasswordHash, password)
	if err != nil {
		return nil, "", "", ErrWrongPassword
	}

	_ = s.repo.UpdateLastLogin(user)

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("login: generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("login: generate refresh token: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

func (s *authUseCase) Logout(refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	if len(refreshToken) < refreshTokenPrefixLength {
    	return ErrInvalidToken
	}

	prefix := refreshToken[:refreshTokenPrefixLength]
	if err := s.repo.DeleteRefreshToken(prefix); err != nil {
		return fmt.Errorf("logout: delete refresh token: %w", err)
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
	tokenString, hashString, expiresAt, err := s.tokenProvider.GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generateRefreshToken: generate: %w", err)
	}

	prefix := tokenString[:refreshTokenPrefixLength]

	token := &models.RefreshToken{
		UserID:      user.ID,
		TokenPrefix: string(prefix),
		TokenHash:   hashString,
		ExpiresAt:   expiresAt,
	}
	err = s.repo.SaveRefreshToken(token)
	if err != nil {
		return "", fmt.Errorf("generateRefreshToken: save: %w", err)
	}

	return tokenString, nil
}

func (s *authUseCase) RefreshTokens(clientToken string) (*models.User, string, string, error) {
	if len(clientToken) < refreshTokenPrefixLength {
    	return nil, "", "", ErrInvalidToken
	}
	
	prefix := clientToken[:refreshTokenPrefixLength]
	storedToken, err := s.repo.GetRefreshTokenByPrefix(prefix)
	if err != nil {
		if errors.Is(err, repositories.ErrRefreshTokenNotFound) {
			return nil, "", "", ErrInvalidToken
		}
		return nil, "", "", fmt.Errorf("refresh tokens: get refresh token: %w", err)
	}

	if !s.tokenProvider.IsRefreshTokenValid(clientToken, storedToken.TokenHash) || 
		storedToken.ExpiresAt.Before(time.Now()) {
		return nil, "", "", ErrInvalidToken
	}

	if err := s.repo.DeleteRefreshToken(prefix); err != nil {
		return nil, "", "", fmt.Errorf("refresh tokens: delete refresh token: %w", err)
	}

	user, err := s.repo.GetUserByID(storedToken.UserID)
	if err != nil {
		return nil, "", "", ErrUserNotFound
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("refresh tokens: generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("refresh tokens: generate refresh token: %w", err)
	}

	return user, accessToken, refreshToken, nil
}