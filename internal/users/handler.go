package users

import "github.com/gofiber/fiber/v3"

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
	//users := app.Group("/users")
	//users.Get("/:username", h.getUserByUsername)

}