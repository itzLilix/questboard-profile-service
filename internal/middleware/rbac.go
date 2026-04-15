package middleware

import (
	"github.com/gofiber/fiber/v3"
	dtos "github.com/itzLilix/questboard-shared/DTOs"
	"github.com/rs/zerolog"
)

type TokenParser interface {
    ParseToken(tokenString string) (*dtos.TokenClaims, error)
}

const (
    LocalsUserID   = "userID"
    LocalsUserRole = "userRole"
)

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
        c.Locals(LocalsUserID, claims.UserID)
        c.Locals(LocalsUserRole, claims.Role)
        return c.Next()
    }
}