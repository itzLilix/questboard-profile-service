package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/itzLilix/questboard-profile-service/internal/entities"
	"github.com/itzLilix/questboard-profile-service/internal/infrastructure"
	"github.com/itzLilix/questboard-shared/dtos"
)

type AuthUsecase interface {
	Register(ctx context.Context, username, displayname, email, password string) (*dtos.PrivateProfileData, string, string, error)
	Login(ctx context.Context, username, password string) (*dtos.PrivateProfileData, string, string, error)
	Logout(ctx context.Context, refreshToken string) error
	ValidateToken(ctx context.Context, tokenString string) (*dtos.PrivateProfileData, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*dtos.PrivateProfileData, string, string, error)
}

type authUsecase struct {
	repo           AuthRepository
	tokenProvider  TokenProvider
	refreshTokens  RefreshTokenManager
	passwordHasher PasswordHasher
}

const (
	refreshTokenPrefixLength = 8
)

func NewAuthUsecase(
	repo AuthRepository,
	tokenProvider TokenProvider,
	refreshTokens RefreshTokenManager,
	passwordHasher PasswordHasher,
) AuthUsecase {
	return &authUsecase{
		repo:           repo,
		tokenProvider:  tokenProvider,
		refreshTokens:  refreshTokens,
		passwordHasher: passwordHasher,
	}
}

func (s *authUsecase) ValidateToken(ctx context.Context, tokenString string) (*dtos.PrivateProfileData, error) {
	claims, err := s.tokenProvider.ParseToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("validateToken: parse token: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("validateToken: get user: %w", err)
	}

	return mapUserToPrivateProfile(user), nil
}

func (s *authUsecase) Register(ctx context.Context, username, displayname, email, password string) (*dtos.PrivateProfileData, string, string, error) {
	passwordHash, err := s.passwordHasher.HashPassword(password)
	if err != nil {
		return nil, "", "", fmt.Errorf("register: hash password: %w", err)
	}

	err = validateUsername(username)
	if err != nil {
		return nil, "", "", wrapInvalidDataError(err)
	}

	err = validateEmail(email)
	if err != nil {
		return nil, "", "", wrapInvalidDataError(err)
	}

	err = validateDisplayName(displayname)
	if err != nil {
		return nil, "", "", wrapInvalidDataError(err)
	}

	user := &entities.User{
		Username:     username,
		DisplayName:  displayname,
		Email:        email,
		PasswordHash: passwordHash,
	}

	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, "", "", mapRepoErr("register: create user", err)
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("register: generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(ctx, user)
	if err != nil {
		return nil, "", "", fmt.Errorf("register: generate refresh token: %w", err)
	}

	return mapUserToPrivateProfile(user), accessToken, refreshToken, nil
}

func (s *authUsecase) Login(ctx context.Context, email, password string) (*dtos.PrivateProfileData, string, string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserNotFound) {
			return nil, "", "", ErrWrongPassword
		}
		return nil, "", "", fmt.Errorf("login: get user: %w", ErrInternal)
	}

	err = s.passwordHasher.CompareHashAndPassword(user.PasswordHash, password)
	if err != nil {
		return nil, "", "", ErrWrongPassword
	}

	_ = s.repo.UpdateLastLogin(ctx, user)

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("login: generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(ctx, user)
	if err != nil {
		return nil, "", "", fmt.Errorf("login: generate refresh token: %w", err)
	}

	return mapUserToPrivateProfile(user), accessToken, refreshToken, nil
}

func (s *authUsecase) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	if len(refreshToken) < refreshTokenPrefixLength {
    	return ErrInvalidToken
	}

	prefix := refreshToken[:refreshTokenPrefixLength]
	if err := s.repo.DeleteRefreshToken(ctx, prefix); err != nil {
		return fmt.Errorf("logout: delete refresh token: %w", err)
	}

	return nil
}

func (s *authUsecase) generateAccessToken(user *entities.User) (string, error) {
	token, err := s.tokenProvider.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", fmt.Errorf("generateAccessToken: %w", err)
	}

	return token, nil
}

func (s *authUsecase) generateRefreshToken(ctx context.Context, user *entities.User) (string, error) {
	tokenString, hashString, expiresAt, err := s.refreshTokens.Generate()
	if err != nil {
		return "", fmt.Errorf("generateRefreshToken: generate: %w", err)
	}

	prefix := tokenString[:refreshTokenPrefixLength]

	token := &entities.RefreshToken{
		UserID:      user.ID,
		TokenPrefix: string(prefix),
		TokenHash:   hashString,
		ExpiresAt:   expiresAt,
	}
	err = s.repo.SaveRefreshToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("generateRefreshToken: save: %w", err)
	}

	return tokenString, nil
}

func (s *authUsecase) RefreshTokens(ctx context.Context, clientToken string) (*dtos.PrivateProfileData, string, string, error) {
	if len(clientToken) < refreshTokenPrefixLength {
    	return nil, "", "", ErrInvalidToken
	}
	
	prefix := clientToken[:refreshTokenPrefixLength]
	storedToken, err := s.repo.GetRefreshTokenByPrefix(ctx, prefix)
	if err != nil {
		if errors.Is(err, infrastructure.ErrRefreshTokenNotFound) {
			return nil, "", "", ErrInvalidToken
		}
		return nil, "", "", fmt.Errorf("refresh tokens: get refresh token: %w", err)
	}

	if !s.refreshTokens.IsValid(clientToken, storedToken.TokenHash) ||
		storedToken.ExpiresAt.Before(time.Now()) {
		return nil, "", "", ErrInvalidToken
	}

	if err := s.repo.DeleteRefreshToken(ctx, prefix); err != nil {
		return nil, "", "", fmt.Errorf("refresh tokens: delete refresh token: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, "", "", ErrUserNotFound
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("refresh tokens: generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(ctx, user)
	if err != nil {
		return nil, "", "", fmt.Errorf("refresh tokens: generate refresh token: %w", err)
	}

	return mapUserToPrivateProfile(user), accessToken, refreshToken, nil
}