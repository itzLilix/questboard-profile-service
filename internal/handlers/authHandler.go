package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/itzLilix/questboard-profile-service/internal/config"
	"github.com/itzLilix/questboard-profile-service/internal/usecase"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

type AuthHandler interface {
	RegisterRoutes(app *fiber.App)
}

type authHandler struct {
	usecase usecase.AuthUsecase
	log     zerolog.Logger
	cfg    *config.Config
}

const (
	accessCookie = "access_token"
	refreshCookie = "refresh_token"
)

func NewAuthHandler(usecase usecase.AuthUsecase, log zerolog.Logger, cfg *config.Config) AuthHandler {
	return &authHandler{usecase: usecase, log: log, cfg: cfg}
}

func (h *authHandler) RegisterRoutes(app *fiber.App) {
	auth := app.Group("/auth")
	auth.Post("/login", h.login)
	auth.Post("/signup", h.signup)
	auth.Post("/logout", h.logout)
	auth.Get("/activate/:link", h.activate)
	auth.Get("/refresh", h.refresh)
}

func (h *authHandler) login(c fiber.Ctx) error {
	type request struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	var req request
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	user, accessToken, refreshToken, err := h.usecase.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) || errors.Is(err, usecase.ErrWrongPassword) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Неверная почта или пароль",
			})
		}
		h.log.Error().Err(err).Str("email", req.Email).Msg("login failed")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	h.setAuthCookies(c, accessToken, refreshToken)

	return c.Status(fiber.StatusOK).JSON(user)
}

func (h *authHandler) signup(c fiber.Ctx) error {
	type request struct {
		Username string `json:"username"`
		DisplayName string `json:"displayName"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req request
	if err := c.Bind().Body(&req); err != nil {
        return err
    }

	user, accessToken, refreshToken, err := h.usecase.Register(req.Username, req.DisplayName, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrEmailExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message": "Почта уже зарегестрирована",
			})
		} else if errors.Is(err, usecase.ErrUsernameExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message": "Пользователь с таким именем уже существует",
			})
		}
		h.log.Error().Err(err).Str("email", req.Email).Msg("signup failed")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	h.setAuthCookies(c, accessToken, refreshToken)

	return c.Status(fiber.StatusCreated).JSON(user)
}

func (h *authHandler) logout(c fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
    err := h.usecase.Logout(refreshToken)

	h.clearAuthCookies(c)

	if err != nil {
		h.log.Error().Err(err).Msg("logout failed")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *authHandler) activate(c fiber.Ctx) error {
	return nil;
}

func (h *authHandler) refresh(c fiber.Ctx) error {
	oldRefreshToken := c.Cookies("refresh_token")
	if oldRefreshToken == "" {
    	return c.SendStatus(fiber.StatusUnauthorized)
	}
	
	user, accessToken, refreshToken, err := h.usecase.RefreshTokens(oldRefreshToken)
	if err != nil {
		h.log.Warn().Err(err).Msg("token refresh failed")
		return c.SendStatus(fiber.StatusUnauthorized)
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