package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/itzLilix/QuestBoard/backend/internal/useCases"
)

type UsersHandler interface {
	RegisterRoutes(app *fiber.App)
}

type usersHandler struct {
	service useCases.UsersUseCase
}

func NewHandler(service useCases.UsersUseCase) UsersHandler {
	return &usersHandler{service: service}
}

func (h *usersHandler) RegisterRoutes(app *fiber.App) {
	//users := app.Group("/users")
	//users.Get("/:username", h.getUserByUsername)
}

