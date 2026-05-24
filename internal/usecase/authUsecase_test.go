package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/itzLilix/questboard-profile-service/internal/entities"
	"github.com/itzLilix/questboard-profile-service/internal/infrastructure"
	"github.com/itzLilix/questboard-shared/dtos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type authMocks struct {
	repo     *MockAuthRepository
	tokens   *MockTokenProvider
	refresh  *MockRefreshTokenManager
	hasher   *MockPasswordHasher
}

func newAuthUC(t *testing.T) (*authUsecase, *authMocks) {
	m := &authMocks{
		repo:    NewMockAuthRepository(t),
		tokens:  NewMockTokenProvider(t),
		refresh: NewMockRefreshTokenManager(t),
		hasher:  NewMockPasswordHasher(t),
	}
	return NewAuthUsecase(m.repo, m.tokens, m.refresh, m.hasher), m
}

// ---------- Register ----------

func TestAuthUsecase_Register_Success(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	m.hasher.EXPECT().HashPassword("pw").Return("hash", nil)
	m.repo.EXPECT().CreateUser(ctx, mock.MatchedBy(func(u *entities.User) bool {
		return u.Username == "john" && u.Email == "john@example.com" && u.PasswordHash == "hash"
	})).Return(nil)
	m.tokens.EXPECT().GenerateAccessToken(mock.Anything, mock.Anything).Return("access-tok", nil)
	expiry := time.Now().Add(time.Hour)
	m.refresh.EXPECT().Generate().Return("refreshXXXXXXXX", "refresh-hash", expiry, nil)
	m.repo.EXPECT().SaveRefreshToken(ctx, mock.MatchedBy(func(rt *entities.RefreshToken) bool {
		return rt.TokenPrefix == "refreshX" && rt.TokenHash == "refresh-hash"
	})).Return(nil)

	profile, access, refresh, err := uc.Register(ctx, "john", "John", "john@example.com", "pw")
	require.NoError(t, err)
	assert.Equal(t, "access-tok", access)
	assert.Equal(t, "refreshXXXXXXXX", refresh)
	assert.Equal(t, "john", profile.Username)
}

func TestAuthUsecase_Register_HashError(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	m.hasher.EXPECT().HashPassword("pw").Return("", errors.New("hash failed"))

	_, _, _, err := uc.Register(ctx, "john", "John", "john@example.com", "pw")
	assert.Error(t, err)
}

func TestAuthUsecase_Register_InvalidUsername(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	m.hasher.EXPECT().HashPassword("pw").Return("hash", nil)

	_, _, _, err := uc.Register(ctx, "bad name!", "John", "john@example.com", "pw")
	assert.ErrorIs(t, err, ErrInvalidData)
}

func TestAuthUsecase_Register_InvalidEmail(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	m.hasher.EXPECT().HashPassword("pw").Return("hash", nil)

	_, _, _, err := uc.Register(ctx, "john", "John", "not-an-email", "pw")
	assert.ErrorIs(t, err, ErrInvalidData)
}

func TestAuthUsecase_Register_InvalidDisplayName(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	m.hasher.EXPECT().HashPassword("pw").Return("hash", nil)

	_, _, _, err := uc.Register(ctx, "john", "", "john@example.com", "pw")
	assert.ErrorIs(t, err, ErrInvalidData)
}

func TestAuthUsecase_Register_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	m.hasher.EXPECT().HashPassword("pw").Return("hash", nil)
	m.repo.EXPECT().CreateUser(ctx, mock.Anything).Return(infrastructure.ErrDuplicateEmail)

	_, _, _, err := uc.Register(ctx, "john", "John", "john@example.com", "pw")
	assert.ErrorIs(t, err, ErrEmailExists)
}

// ---------- Login ----------

func TestAuthUsecase_Login_Success(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	user := &entities.User{ID: "u1", Email: "x@y.z", PasswordHash: "h", Role: dtos.UserRole}
	m.repo.EXPECT().GetUserByEmail(ctx, "x@y.z").Return(user, nil)
	m.hasher.EXPECT().CompareHashAndPassword("h", "pw").Return(nil)
	m.repo.EXPECT().UpdateLastLogin(ctx, user).Return(nil)
	m.tokens.EXPECT().GenerateAccessToken("u1", dtos.UserRole).Return("acc", nil)
	m.refresh.EXPECT().Generate().Return("rrrrrrrrAB", "rhash", time.Now().Add(time.Hour), nil)
	m.repo.EXPECT().SaveRefreshToken(ctx, mock.Anything).Return(nil)

	profile, access, refresh, err := uc.Login(ctx, "x@y.z", "pw")
	require.NoError(t, err)
	assert.Equal(t, "acc", access)
	assert.Equal(t, "rrrrrrrrAB", refresh)
	assert.Equal(t, "u1", profile.ID)
}

func TestAuthUsecase_Login_UserNotFound(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	m.repo.EXPECT().GetUserByEmail(ctx, "x@y.z").Return(nil, infrastructure.ErrUserNotFound)

	_, _, _, err := uc.Login(ctx, "x@y.z", "pw")
	assert.ErrorIs(t, err, ErrWrongPassword)
}

func TestAuthUsecase_Login_RepoError(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	m.repo.EXPECT().GetUserByEmail(ctx, "x@y.z").Return(nil, errors.New("db down"))

	_, _, _, err := uc.Login(ctx, "x@y.z", "pw")
	assert.ErrorIs(t, err, ErrInternal)
}

func TestAuthUsecase_Login_WrongPassword(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	user := &entities.User{ID: "u1", PasswordHash: "h"}
	m.repo.EXPECT().GetUserByEmail(ctx, "x@y.z").Return(user, nil)
	m.hasher.EXPECT().CompareHashAndPassword("h", "pw").Return(errors.New("mismatch"))

	_, _, _, err := uc.Login(ctx, "x@y.z", "pw")
	assert.ErrorIs(t, err, ErrWrongPassword)
}

// ---------- Logout ----------

func TestAuthUsecase_Logout_Empty(t *testing.T) {
	uc, _ := newAuthUC(t)
	assert.NoError(t, uc.Logout(context.Background(), ""))
}

func TestAuthUsecase_Logout_TooShort(t *testing.T) {
	uc, _ := newAuthUC(t)
	err := uc.Logout(context.Background(), "abc")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestAuthUsecase_Logout_Success(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)
	m.repo.EXPECT().DeleteRefreshToken(ctx, "prefix12").Return(nil)
	assert.NoError(t, uc.Logout(ctx, "prefix12tail"))
}

func TestAuthUsecase_Logout_RepoError(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)
	m.repo.EXPECT().DeleteRefreshToken(ctx, "prefix12").Return(errors.New("db"))
	err := uc.Logout(ctx, "prefix12tail")
	assert.Error(t, err)
}

// ---------- ValidateToken ----------

func TestAuthUsecase_ValidateToken_Success(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	m.tokens.EXPECT().ParseToken("jwt").Return(&dtos.TokenClaims{UserID: "u1", Role: dtos.UserRole}, nil)
	m.repo.EXPECT().GetUserByID(ctx, "u1").Return(&entities.User{ID: "u1", Username: "john"}, nil)

	profile, err := uc.ValidateToken(ctx, "jwt")
	require.NoError(t, err)
	assert.Equal(t, "john", profile.Username)
}

func TestAuthUsecase_ValidateToken_ParseError(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	m.tokens.EXPECT().ParseToken("jwt").Return(nil, errors.New("bad token"))

	_, err := uc.ValidateToken(ctx, "jwt")
	assert.Error(t, err)
}

func TestAuthUsecase_ValidateToken_UserLookupError(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	m.tokens.EXPECT().ParseToken("jwt").Return(&dtos.TokenClaims{UserID: "u1"}, nil)
	m.repo.EXPECT().GetUserByID(ctx, "u1").Return(nil, infrastructure.ErrUserNotFound)

	_, err := uc.ValidateToken(ctx, "jwt")
	assert.Error(t, err)
}

// ---------- RefreshTokens ----------

func TestAuthUsecase_RefreshTokens_TooShort(t *testing.T) {
	uc, _ := newAuthUC(t)
	_, _, _, err := uc.RefreshTokens(context.Background(), "abc")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestAuthUsecase_RefreshTokens_NotFound(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	m.repo.EXPECT().GetRefreshTokenByPrefix(ctx, "prefix12").
		Return(nil, infrastructure.ErrRefreshTokenNotFound)

	_, _, _, err := uc.RefreshTokens(ctx, "prefix12tail")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestAuthUsecase_RefreshTokens_InvalidHash(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	stored := &entities.RefreshToken{
		UserID: "u1", TokenHash: "stored", ExpiresAt: time.Now().Add(time.Hour),
	}
	m.repo.EXPECT().GetRefreshTokenByPrefix(ctx, "prefix12").Return(stored, nil)
	m.refresh.EXPECT().IsValid("prefix12tail", "stored").Return(false)

	_, _, _, err := uc.RefreshTokens(ctx, "prefix12tail")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestAuthUsecase_RefreshTokens_Expired(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	stored := &entities.RefreshToken{
		UserID: "u1", TokenHash: "stored", ExpiresAt: time.Now().Add(-time.Hour),
	}
	m.repo.EXPECT().GetRefreshTokenByPrefix(ctx, "prefix12").Return(stored, nil)
	m.refresh.EXPECT().IsValid("prefix12tail", "stored").Return(true)

	_, _, _, err := uc.RefreshTokens(ctx, "prefix12tail")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestAuthUsecase_RefreshTokens_Success(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	stored := &entities.RefreshToken{
		UserID: "u1", TokenHash: "stored", ExpiresAt: time.Now().Add(time.Hour),
	}
	user := &entities.User{ID: "u1", Username: "john", Role: dtos.UserRole}

	m.repo.EXPECT().GetRefreshTokenByPrefix(ctx, "prefix12").Return(stored, nil)
	m.refresh.EXPECT().IsValid("prefix12tail", "stored").Return(true)
	m.repo.EXPECT().DeleteRefreshToken(ctx, "prefix12").Return(nil)
	m.repo.EXPECT().GetUserByID(ctx, "u1").Return(user, nil)
	m.tokens.EXPECT().GenerateAccessToken("u1", dtos.UserRole).Return("new-acc", nil)
	m.refresh.EXPECT().Generate().Return("newrefrshTTTT", "newhash", time.Now().Add(time.Hour), nil)
	m.repo.EXPECT().SaveRefreshToken(ctx, mock.Anything).Return(nil)

	profile, access, refresh, err := uc.RefreshTokens(ctx, "prefix12tail")
	require.NoError(t, err)
	assert.Equal(t, "new-acc", access)
	assert.Equal(t, "newrefrshTTTT", refresh)
	assert.Equal(t, "john", profile.Username)
}

func TestAuthUsecase_RefreshTokens_UserNotFound(t *testing.T) {
	ctx := context.Background()
	uc, m := newAuthUC(t)

	stored := &entities.RefreshToken{
		UserID: "u1", TokenHash: "stored", ExpiresAt: time.Now().Add(time.Hour),
	}
	m.repo.EXPECT().GetRefreshTokenByPrefix(ctx, "prefix12").Return(stored, nil)
	m.refresh.EXPECT().IsValid("prefix12tail", "stored").Return(true)
	m.repo.EXPECT().DeleteRefreshToken(ctx, "prefix12").Return(nil)
	m.repo.EXPECT().GetUserByID(ctx, "u1").Return(nil, infrastructure.ErrUserNotFound)

	_, _, _, err := uc.RefreshTokens(ctx, "prefix12tail")
	assert.ErrorIs(t, err, ErrUserNotFound)
}
