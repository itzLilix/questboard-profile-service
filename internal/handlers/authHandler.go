package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/itzLilix/QuestBoard/backend/internal/useCases"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

type AuthHandler interface {
	RegisterRoutes(app *fiber.App)
}

type authHandler struct {
	useCase useCases.AuthUseCase
	log     zerolog.Logger
}

func NewAuthHandler(useCase useCases.AuthUseCase, log zerolog.Logger) AuthHandler {
	return &authHandler{useCase: useCase, log: log}
}

func (h *authHandler) RegisterRoutes(app *fiber.App) {
	auth := app.Group("/auth")
	auth.Post("/login", h.login)
	auth.Post("/signup", h.signup)
	auth.Post("/logout", h.logout)
	auth.Get("/activate/:link", h.activate)
	auth.Get("/refresh", h.refresh)
	auth.Get("/me", h.restoreSession)
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
	user, accessToken, refreshToken, err := h.useCase.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, useCases.ErrUserNotFound) || errors.Is(err, useCases.ErrWrongPassword) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Неверная почта или пароль",
			})
		}
		h.log.Error().Err(err).Str("email", req.Email).Msg("login failed")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})

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

	user, accessToken, refreshToken, err := h.useCase.Register(req.Username, req.DisplayName, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, useCases.ErrEmailExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message": "Почта уже зарегестрирована",
			})
		} else if errors.Is(err, useCases.ErrUsernameExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message": "Пользователь с таким именем уже существует",
			})
		}
		h.log.Error().Err(err).Str("email", req.Email).Msg("signup failed")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})

	return c.Status(fiber.StatusCreated).JSON(user)
}

func (h *authHandler) logout(c fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
    err := h.useCase.Logout(refreshToken)

    c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  fasthttp.CookieExpireDelete,
		HTTPOnly: true,
		Path:     "/",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  fasthttp.CookieExpireDelete,
		HTTPOnly: true,
		Path:     "/",
	})

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
	
	user, accessToken, refreshToken, err := h.useCase.RefreshTokens(oldRefreshToken)
	if err != nil {
		h.log.Warn().Err(err).Msg("token refresh failed")
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})

	return c.Status(fiber.StatusOK).JSON(user)
}

func (h *authHandler) restoreSession(c fiber.Ctx) error {
	tokenString := c.Cookies("access_token")

	if tokenString == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	user, err := h.useCase.ValidateToken(tokenString)
	if err != nil {
		h.log.Warn().Err(err).Msg("session restore failed")
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	return c.Status(fiber.StatusOK).JSON(user)
}