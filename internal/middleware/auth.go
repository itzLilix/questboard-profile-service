package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/itzLilix/QuestBoard/backend/internal/auth"
)

func Protected(service auth.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
        token := c.Cookies("access_token")
        if token == "" {
            return c.SendStatus(fiber.StatusUnauthorized)
        }
        user, err := service.ValidateToken(token)
        if err != nil {
            return c.SendStatus(fiber.StatusUnauthorized)
        }
        c.Locals("user", user)
        return c.Next()
    }
}