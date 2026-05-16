package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/itzLilix/questboard-profile-service/internal/middleware"
	"github.com/itzLilix/questboard-profile-service/internal/usecase"
	"github.com/itzLilix/questboard-shared/dtos"
)

func viewerFromCtx(c fiber.Ctx) *usecase.Viewer {
	v := &usecase.Viewer{}
	if id, ok := c.Locals(middleware.LocalsUserID).(string); ok {
		v.UserID = id
	}
	if role, ok := c.Locals(middleware.LocalsUserRole).(dtos.Role); ok {
		v.Role = role
	}
	return v
}