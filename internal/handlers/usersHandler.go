package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/itzLilix/questboard-profile-service/internal/middleware"
	"github.com/itzLilix/questboard-profile-service/internal/usecases"
	"github.com/rs/zerolog"
)

type UsersHandler interface {
	RegisterRoutes(app *fiber.App)
}

type usersHandler struct {
	rbac middleware.RBACMiddleware
	log zerolog.Logger
	usecase usecases.UsersUsecase
}

func NewHandler(usecase usecases.UsersUsecase, log zerolog.Logger, rbac middleware.RBACMiddleware) UsersHandler {
	return &usersHandler{usecase: usecase, log: log, rbac: rbac}
}

func (h *usersHandler) RegisterRoutes(app *fiber.App) {
	users := app.Group("/users")
	users.Get("/me", h.rbac.Protected(), h.getMyProfile)
	users.Patch("/me", h.rbac.Protected(), h.updateProfile)
	users.Get("/:username", h.getUserByUsername)
}

func (h *usersHandler) getUserByUsername(c fiber.Ctx) error {
	username := c.Params("username")
	profile, err := h.usecase.GetPublicProfile(username)
	if err != nil {
		if err == usecases.ErrUserNotFound {
			h.log.Error().Err(err).Str("username", username).Msg("user not found in getUserByUsername")
			return c.SendStatus(fiber.StatusNotFound)
		}
		h.log.Error().Err(err).Str("username", username).Msg("error getting public profile")
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Status(fiber.StatusOK).JSON(profile)
}

func (h *usersHandler) getMyProfile(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(string)
	profile, err := h.usecase.GetPrivateProfile(userID)
	if err != nil {
		if err == usecases.ErrUserNotFound {
			h.log.Error().Err(err).Str("userID", userID).Msg("user not found in getMyProfile")
			return c.SendStatus(fiber.StatusNotFound)
		}
		h.log.Error().Err(err).Str("userID", userID).Msg("error getting private profile")
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Status(fiber.StatusOK).JSON(profile)
}

func (h *usersHandler) updateProfile(c fiber.Ctx) error {
	userID := c.Locals(middleware.LocalsUserID).(string)
	
	req := &usecases.UpdateProfileInput{
		UserID: userID,
	}
	
	if err := c.Bind().Body(&req); err != nil {
		h.log.Error().Err(err).Str("userID", userID).Msg("invalid request body in updateProfile")
		return c.SendStatus(fiber.StatusBadRequest)
	}

	user, err := h.usecase.UpdateProfile(req)
	if err != nil {
		if errors.Is(err, usecases.ErrInvalidData) {
			h.log.Error().Err(err).Str("userID", userID).Msg("invalid data in updateProfile")
			return c.SendStatus(fiber.StatusConflict)
		}
		if errors.Is(err, usecases.ErrUserNotFound) {
			h.log.Error().Err(err).Str("userID", userID).Msg("user not found in updateProfile")
			return c.SendStatus(fiber.StatusNotFound)
		}
		if errors.Is(err, usecases.ErrUsernameExists) {
			h.log.Error().Err(err).Str("userID", userID).Msg("username already exists in updateProfile")
			return c.SendStatus(fiber.StatusConflict)
		}
		h.log.Error().Err(err).Str("userID", userID).Msg("error updating profile")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(user)
}