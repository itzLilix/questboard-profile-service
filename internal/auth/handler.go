package auth

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

type Handler interface {
	RegisterRoutes(app *fiber.App)
}

type handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return &handler{service: service}
}

func (h *handler) RegisterRoutes(app *fiber.App) {
	auth := app.Group("/auth")
	auth.Post("/login", h.login)
	auth.Post("/signup", h.signup)
	auth.Post("/logout", h.logout)
	auth.Get("/activate/:link", h.activate)
	auth.Get("/refresh", h.refresh)
	auth.Get("/me", h.restoreSession)
}

func (h *handler) login(c fiber.Ctx) error {
	type request struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	var req request
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	user, accessToken, refreshToken, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) || errors.Is(err, ErrWrongPassword) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Неверная почта или пароль",
			})
		}
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

func (h *handler) signup(c fiber.Ctx) error {
	type request struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req request
	if err := c.Bind().Body(&req); err != nil {
        return err
    }
	user, accessToken, refreshToken, err := h.service.Register(req.Username, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message": "Почта уже зарегестрирована",
			})
		} else if errors.Is(err, ErrUsernameExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message": "Пользователь с таким именем уже существует",
			})
		}
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

func (h *handler) logout(c fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
    h.service.Logout(refreshToken)
    c.ClearCookie("access_token")
    c.ClearCookie("refresh_token")
    return c.SendStatus(fiber.StatusOK)
}

func (h *handler) activate(c fiber.Ctx) error {
	return nil;
}

func (h *handler) refresh(c fiber.Ctx) error {
	oldRefreshToken := c.Cookies("refresh_token")
	if oldRefreshToken == "" {
    	return c.SendStatus(fiber.StatusUnauthorized)
	}
	
	user, accessToken, refreshToken, err := h.service.RefreshTokens(oldRefreshToken)
	if err != nil {
		return  c.SendStatus(fiber.StatusUnauthorized)
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

func (h *handler) restoreSession(c fiber.Ctx) error {
	tokenString := c.Cookies("access_token")

	if tokenString == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	user, err := h.service.ValidateToken(tokenString)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	return c.Status(fiber.StatusOK).JSON(user)
}