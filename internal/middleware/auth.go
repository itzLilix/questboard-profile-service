// internal/middleware/auth.go
package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/itzLilix/QuestBoard/backend/internal/models"
	"github.com/rs/zerolog"
)

type TokenParser interface {
    ParseToken(tokenString string) (*models.TokenClaims, error)
}

func Protected(parser TokenParser, log zerolog.Logger) fiber.Handler {
    return func(c fiber.Ctx) error {
        token := c.Cookies("access_token")
        if token == "" {
            return c.SendStatus(fiber.StatusUnauthorized)
        }
        claims, err := parser.ParseToken(token)
        if err != nil {
            log.Warn().Err(err).Str("path", c.Path()).Msg("unauthorized")
            return c.SendStatus(fiber.StatusUnauthorized)
        }
        c.Locals("userID", claims.UserID)
        c.Locals("userRole", claims.Role)
        return c.Next()
    }
}