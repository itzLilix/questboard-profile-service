package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/itzLilix/QuestBoard/backend/internal/useCases"
	"github.com/rs/zerolog"
)

func Protected(service useCases.AuthUseCase, log zerolog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := c.Cookies("access_token")
		if token == "" {
			log.Warn().Str("path", c.Path()).Msg("unauthorized: missing token")
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		user, err := service.ValidateToken(token)
		if err != nil {
			log.Warn().Err(err).Str("path", c.Path()).Msg("unauthorized: invalid token")
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		c.Locals("user", user)
		return c.Next()
	}
}