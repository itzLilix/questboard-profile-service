package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/itzLilix/questboard-profile-service/internal/usecases"
)

type UsersHandler interface {
	RegisterRoutes(app *fiber.App)
}

type usersHandler struct {
	service usecases.UsersUseCase
}

func NewHandler(service usecases.UsersUseCase) UsersHandler {
	return &usersHandler{service: service}
}

func (h *usersHandler) RegisterRoutes(app *fiber.App) {
	//users := app.Group("/users")
	//users.Get("/:username", h.getUserByUsername)
}

