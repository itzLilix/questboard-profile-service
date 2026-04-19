package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/itzLilix/questboard-profile-service/internal/middleware"
	"github.com/itzLilix/questboard-profile-service/internal/usecases"
	"github.com/itzLilix/questboard-shared/dtos"
	"github.com/rs/zerolog"
)

type UsersHandler interface {
	RegisterRoutes(app *fiber.App)
}

type usersHandler struct {
	rbac    middleware.RBACMiddleware
	log     zerolog.Logger
	usecase usecases.UsersUsecase
}

func NewHandler(usecase usecases.UsersUsecase, log zerolog.Logger, rbac middleware.RBACMiddleware) UsersHandler {
	return &usersHandler{usecase: usecase, log: log, rbac: rbac}
}

func (h *usersHandler) RegisterRoutes(app *fiber.App) {
	users := app.Group("/users")
	users.Get("/:username", h.rbac.Optional(), h.getProfileByUsername)
	//users.Get("/me", h.rbac.Protected(), h.getMyProfile)
	users.Patch("/me", h.rbac.Protected(), h.updateProfile)

	users.Post("/:username/follow", h.rbac.Protected(), h.followUser)
	users.Delete("/:username/follow", h.rbac.Protected(), h.unfollowUser)
}

func viewerFromCtx(c fiber.Ctx) *usecases.ViewerContext {
	v := &usecases.ViewerContext{}
	if id, ok := c.Locals(middleware.LocalsUserID).(string); ok {
		v.UserID = id
	}
	if role, ok := c.Locals(middleware.LocalsUserRole).(dtos.Role); ok {
		v.IsAdmin = role == dtos.AdminRole
	}
	return v
}

func (h *usersHandler) getProfileByUsername(c fiber.Ctx) error {
	username := c.Params("username")
	viewer := viewerFromCtx(c)
	profile, err := h.usecase.GetPublicProfile(viewer, username)
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
	viewer := viewerFromCtx(c)
	profile, err := h.usecase.GetPrivateProfile(viewer)
	if err != nil {
		if err == usecases.ErrUserNotFound {
			h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("user not found in getMyProfile")
			return c.SendStatus(fiber.StatusNotFound)
		}
		h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("error getting private profile")
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Status(fiber.StatusOK).JSON(profile)
}

func (h *usersHandler) updateProfile(c fiber.Ctx) error {
	viewer := viewerFromCtx(c)

	req := &usecases.UpdateProfileInput{}

	if err := c.Bind().Body(&req); err != nil {
		h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("invalid request body in updateProfile")
		return c.SendStatus(fiber.StatusBadRequest)
	}

	user, err := h.usecase.UpdateProfile(viewer, req)
	if err != nil {
		if errors.Is(err, usecases.ErrInvalidData) {
			h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("invalid data in updateProfile")
			return c.SendStatus(fiber.StatusBadRequest)
		}
		if errors.Is(err, usecases.ErrUserNotFound) {
			h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("user not found in updateProfile")
			return c.SendStatus(fiber.StatusNotFound)
		}
		if errors.Is(err, usecases.ErrUsernameExists) {
			h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("username already exists in updateProfile")
			return c.SendStatus(fiber.StatusConflict)
		}
		h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("error updating profile")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

func (h *usersHandler) followUser(c fiber.Ctx) error {
	viewer := viewerFromCtx(c)
	targetUsername := c.Params("username")

	err := h.usecase.Follow(viewer, targetUsername)
	if err != nil {
		if errors.Is(err, usecases.ErrUserNotFound) {
			h.log.Error().Err(err).Str("followerID", viewer.UserID).Str("targetUsername", targetUsername).Msg("user not found in followUser")
			return c.SendStatus(fiber.StatusNotFound)
		}
		if errors.Is(err, usecases.ErrCannotFollowSelf) {
			h.log.Error().Err(err).Str("followerID", viewer.UserID).Str("targetUsername", targetUsername).Msg("cannot follow yourself")
			return c.SendStatus(fiber.StatusNoContent)
		}
		h.log.Error().Err(err).Str("followerID", viewer.UserID).Str("targetUsername", targetUsername).Msg("error following user")
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *usersHandler) unfollowUser(c fiber.Ctx) error {
	viewer := viewerFromCtx(c)
	targetUsername := c.Params("username")

	err := h.usecase.Unfollow(viewer, targetUsername)
	if err != nil {
		if errors.Is(err, usecases.ErrUserNotFound) {
			h.log.Error().Err(err).Str("followerID", viewer.UserID).Str("targetUsername", targetUsername).Msg("user not found in unfollowUser")
			return c.SendStatus(fiber.StatusNotFound)
		}
		h.log.Error().Err(err).Str("followerID", viewer.UserID).Str("targetUsername", targetUsername).Msg("error unfollowing user")
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
