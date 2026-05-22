package middleware

import "github.com/gofiber/fiber/v3"

const internalTokenHeader = "X-Internal-Token"

func InternalOnly(token string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if token == "" || c.Get(internalTokenHeader) != token {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		return c.Next()
	}
}
