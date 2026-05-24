package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/itzLilix/questboard-profile-service/internal/config"
	"github.com/itzLilix/questboard-profile-service/internal/usecase"
	"github.com/itzLilix/questboard-shared/dtos"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

type AuthHandler interface {
	RegisterRoutes(router fiber.Router)
}

type authHandler struct {
	usecase AuthUsecase
	log     zerolog.Logger
	cfg     *config.Config
}

// LoginRequest is the body for POST /v1/auth/login.
type LoginRequest struct {
	Email    string `json:"email"    example:"user@example.com"`
	Password string `json:"password" example:"Secret123!"`
}

// SignupRequest is the body for POST /v1/auth/signup.
type SignupRequest struct {
	Username    string `json:"username"    example:"johndoe"`
	DisplayName string `json:"displayName" example:"John Doe"`
	Email       string `json:"email"       example:"user@example.com"`
	Password    string `json:"password"    example:"Secret123!"`
}

// authResponse aliases dtos.PrivateProfileData for swagger type resolution.
type authResponse = dtos.PrivateProfileData

const (
	accessCookie  = "access_token"
	refreshCookie = "refresh_token"
)

func NewAuthHandler(usecase AuthUsecase, log zerolog.Logger, cfg *config.Config) AuthHandler {
	return &authHandler{usecase: usecase, log: log, cfg: cfg}
}

func (h *authHandler) RegisterRoutes(router fiber.Router) {
	auth := router.Group("/auth")
	auth.Post("/login", h.login)
	auth.Post("/signup", h.signup)
	auth.Post("/logout", h.logout)
	auth.Get("/activate/:link", h.activate)
	auth.Get("/refresh", h.refresh)
}

// @summary      Login
// @tags         auth
// @accept       json
// @produce      json
// @param        request  body      LoginRequest   true  "Credentials"
// @success      200      {object}  authResponse
// @failure      400      {object}  ErrorResponse
// @failure      401      {object}  ErrorResponse
// @failure      500      {object}  ErrorResponse
// @router       /v1/auth/login [post]
func (h *authHandler) login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return handleErr(c, usecase.ErrInvalidData)
	}
	user, accessToken, refreshToken, err := h.usecase.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		h.log.Error().Err(err).Str("email", req.Email).Msg("login failed")
		return handleErr(c, err)
	}
	h.setAuthCookies(c, accessToken, refreshToken)
	return c.Status(fiber.StatusOK).JSON(user)
}

// @summary      Sign up
// @tags         auth
// @accept       json
// @produce      json
// @param        request  body      SignupRequest  true  "Registration data"
// @success      201      {object}  authResponse
// @failure      400      {object}  ErrorResponse
// @failure      409      {object}  ErrorResponse
// @failure      500      {object}  ErrorResponse
// @router       /v1/auth/signup [post]
func (h *authHandler) signup(c fiber.Ctx) error {
	var req SignupRequest
	if err := c.Bind().Body(&req); err != nil {
		return handleErr(c, usecase.ErrInvalidData)
	}
	user, accessToken, refreshToken, err := h.usecase.Register(c.Context(), req.Username, req.DisplayName, req.Email, req.Password)
	if err != nil {
		h.log.Error().Err(err).Str("email", req.Email).Msg("signup failed")
		return handleErr(c, err)
	}
	h.setAuthCookies(c, accessToken, refreshToken)
	return c.Status(fiber.StatusCreated).JSON(user)
}

// @summary      Logout
// @tags         auth
// @success      200
// @failure      500  {object}  ErrorResponse
// @router       /v1/auth/logout [post]
func (h *authHandler) logout(c fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	err := h.usecase.Logout(c.Context(), refreshToken)
	h.clearAuthCookies(c)
	if err != nil {
		h.log.Error().Err(err).Msg("logout failed")
		return handleErr(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

// @summary      Activate account
// @tags         auth
// @param        link  path  string  true  "Activation link"
// @success      200
// @router       /v1/auth/activate/{link} [get]
func (h *authHandler) activate(c fiber.Ctx) error {
	return nil
}

// @summary      Refresh tokens
// @tags         auth
// @produce      json
// @success      200  {object}  authResponse
// @failure      401  {object}  ErrorResponse
// @router       /v1/auth/refresh [get]
func (h *authHandler) refresh(c fiber.Ctx) error {
	oldRefreshToken := c.Cookies("refresh_token")
	if oldRefreshToken == "" {
		return handleErr(c, usecase.ErrInvalidToken)
	}
	user, accessToken, refreshToken, err := h.usecase.RefreshTokens(c.Context(), oldRefreshToken)
	if err != nil {
		h.log.Warn().Err(err).Msg("token refresh failed")
		return handleErr(c, err)
	}
	h.setAuthCookies(c, accessToken, refreshToken)
	return c.Status(fiber.StatusOK).JSON(user)
}

func (h *authHandler) setAuthCookies(c fiber.Ctx, accessToken, refreshToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     accessCookie,
		Value:    accessToken,
		Path:     "/",
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookie,
		Value:    refreshToken,
		Path:     "/",
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})
}

func (h *authHandler) clearAuthCookies(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     accessCookie,
		Value:    "",
		Path:     "/",
		Expires:  fasthttp.CookieExpireDelete,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookie,
		Value:    "",
		Path:     "/",
		Expires:  fasthttp.CookieExpireDelete,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})
}
