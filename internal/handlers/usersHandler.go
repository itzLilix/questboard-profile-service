package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/itzLilix/questboard-profile-service/internal/middleware"
	"github.com/itzLilix/questboard-profile-service/internal/usecase"
	"github.com/itzLilix/questboard-shared/dtos"
	"github.com/rs/zerolog"
)

type UsersHandler interface {
	RegisterRoutes(app *fiber.App)
}

type usersHandler struct {
	rbac    middleware.RBACMiddleware
	log     zerolog.Logger
	usecase usecase.UsersUsecase
}

func NewHandler(usecase usecase.UsersUsecase, log zerolog.Logger, rbac middleware.RBACMiddleware) UsersHandler {
	return &usersHandler{usecase: usecase, log: log, rbac: rbac}
}

func (h *usersHandler) RegisterRoutes(app *fiber.App) {
	users := app.Group("/users")
	users.Get("/:username", h.rbac.Optional(), h.getProfileByUsername)
	//users.Get("/me", h.rbac.Protected(), h.getMyProfile)
	users.Patch("/me", h.rbac.Protected(), h.updateProfile)

	users.Put("/me/avatar", h.rbac.Protected(), h.updateAvatar)
	users.Delete("/me/avatar", h.rbac.Protected(), h.removeAvatar)
	users.Put("/me/banner", h.rbac.Protected(), h.updateBanner)
	users.Delete("/me/banner", h.rbac.Protected(), h.removeBanner)

	users.Post("/:username/follow", h.rbac.Protected(), h.followUser)
	users.Delete("/:username/follow", h.rbac.Protected(), h.unfollowUser)
}

func viewerFromCtx(c fiber.Ctx) *usecase.ViewerContext {
	v := &usecase.ViewerContext{}
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
		if err == usecase.ErrUserNotFound {
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
		if err == usecase.ErrUserNotFound {
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

	req := &usecase.UpdateProfileInput{}

	if err := c.Bind().Body(&req); err != nil {
		h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("invalid request body in updateProfile")
		return c.SendStatus(fiber.StatusBadRequest)
	}

	user, err := h.usecase.UpdateProfile(viewer, req)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidData) {
			h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("invalid data in updateProfile")
			return c.SendStatus(fiber.StatusBadRequest)
		}
		if errors.Is(err, usecase.ErrUserNotFound) {
			h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("user not found in updateProfile")
			return c.SendStatus(fiber.StatusNotFound)
		}
		if errors.Is(err, usecase.ErrUsernameExists) {
			h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("username already exists in updateProfile")
			return c.SendStatus(fiber.StatusConflict)
		}
		h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("error updating profile")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

func (h *usersHandler) updateAvatar(c fiber.Ctx) error {
	viewer := viewerFromCtx(c)

	file, err := c.FormFile("avatar")
	if err != nil {
		h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("missing avatar file in updateAvatar")
		return c.SendStatus(fiber.StatusBadRequest)
	}

	user, err := h.usecase.UpdateAvatar(viewer, file)
	if err != nil {
		return h.handleImageErr(c, viewer.UserID, err, "avatar")
	}
	return c.Status(fiber.StatusOK).JSON(user)
}

func (h *usersHandler) removeAvatar(c fiber.Ctx) error {
	viewer := viewerFromCtx(c)
	user, err := h.usecase.RemoveAvatar(viewer)
	if err != nil {
		return h.handleImageErr(c, viewer.UserID, err, "avatar")
	}
	return c.Status(fiber.StatusOK).JSON(user)
}

func (h *usersHandler) updateBanner(c fiber.Ctx) error {
	viewer := viewerFromCtx(c)

	file, err := c.FormFile("banner")
	if err != nil {
		h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("missing banner file in updateBanner")
		return c.SendStatus(fiber.StatusBadRequest)
	}

	user, err := h.usecase.UpdateBanner(viewer, file)
	if err != nil {
		return h.handleImageErr(c, viewer.UserID, err, "banner")
	}
	return c.Status(fiber.StatusOK).JSON(user)
}

func (h *usersHandler) removeBanner(c fiber.Ctx) error {
	viewer := viewerFromCtx(c)
	user, err := h.usecase.RemoveBanner(viewer);
	if err != nil {
		return h.handleImageErr(c, viewer.UserID, err, "banner")
	}
	return c.Status(fiber.StatusOK).JSON(user)
}

func (h *usersHandler) handleImageErr(c fiber.Ctx, userID string, err error, kind string) error {
	if errors.Is(err, usecase.ErrFileTooLarge) || errors.Is(err, usecase.ErrInvalidFileType) {
		h.log.Error().Err(err).Str("userID", userID).Msgf("invalid %s upload", kind)
		return c.SendStatus(fiber.StatusBadRequest)
	}
	if errors.Is(err, usecase.ErrUserNotFound) {
		h.log.Error().Err(err).Str("userID", userID).Msgf("user not found in %s operation", kind)
		return c.SendStatus(fiber.StatusNotFound)
	}
	h.log.Error().Err(err).Str("userID", userID).Msgf("error in %s operation", kind)
	return c.SendStatus(fiber.StatusInternalServerError)
}

func (h *usersHandler) followUser(c fiber.Ctx) error {
	viewer := viewerFromCtx(c)
	targetUsername := c.Params("username")

	err := h.usecase.Follow(viewer, targetUsername)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			h.log.Error().Err(err).Str("followerID", viewer.UserID).Str("targetUsername", targetUsername).Msg("user not found in followUser")
			return c.SendStatus(fiber.StatusNotFound)
		}
		if errors.Is(err, usecase.ErrCannotFollowSelf) {
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
		if errors.Is(err, usecase.ErrUserNotFound) {
			h.log.Error().Err(err).Str("followerID", viewer.UserID).Str("targetUsername", targetUsername).Msg("user not found in unfollowUser")
			return c.SendStatus(fiber.StatusNotFound)
		}
		h.log.Error().Err(err).Str("followerID", viewer.UserID).Str("targetUsername", targetUsername).Msg("error unfollowing user")
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
