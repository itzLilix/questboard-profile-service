package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	_ "github.com/itzLilix/questboard-profile-service/docs"
	"github.com/itzLilix/questboard-profile-service/internal/middleware"
	"github.com/itzLilix/questboard-profile-service/internal/usecase"
	"github.com/itzLilix/questboard-shared/dtos"
	"github.com/rs/zerolog"
)

type UsersHandler interface {
	RegisterRoutes(router fiber.Router)
	RegisterInternalRoutes(router fiber.Router)
}

type usersHandler struct {
	rbac         middleware.RBACMiddleware
	internalOnly fiber.Handler
	log          zerolog.Logger
	usecase      usecase.UsersUsecase
}

// UpdateStatsRequest is the body for PATCH /internal/stats.
type UpdateStatsRequest struct {
	Stats    map[string]int `json:"stats"`
	StatName string         `json:"statName" enums:"played,hosted"`
}

func NewUsersHandler(usecase usecase.UsersUsecase, log zerolog.Logger, rbac middleware.RBACMiddleware, internalOnly fiber.Handler) UsersHandler {
	return &usersHandler{usecase: usecase, log: log, rbac: rbac, internalOnly: internalOnly}
}

func (h *usersHandler) RegisterRoutes(router fiber.Router) {
	users := router.Group("/users")

	users.Get("/:username", h.rbac.Optional(), h.getProfileByUsername)
	users.Patch("/me", h.rbac.Protected(), h.updateProfile)

	users.Put("/me/avatar", h.rbac.Protected(), h.updateAvatar)
	users.Delete("/me/avatar", h.rbac.Protected(), h.removeAvatar)
	users.Put("/me/banner", h.rbac.Protected(), h.updateBanner)
	users.Delete("/me/banner", h.rbac.Protected(), h.removeBanner)

	users.Post("/:username/follow", h.rbac.Protected(), h.followUser)
	users.Delete("/:username/follow", h.rbac.Protected(), h.unfollowUser)
}

func (h *usersHandler) RegisterInternalRoutes(router fiber.Router) {
	internal := router.Group("/internal")
	internal.Get("/briefs", h.getBriefs)
	internal.Patch("/stats", h.internalOnly, h.updateStats)
}

// @summary      Get public profile
// @tags         users
// @produce      json
// @param        username  path      string                  true  "Username"
// @success      200       {object}  dtos.PublicProfileData
// @failure      404       {object}  ErrorResponse
// @failure      500       {object}  ErrorResponse
// @router       /v1/users/{username} [get]
func (h *usersHandler) getProfileByUsername(c fiber.Ctx) error {
	username := c.Params("username")
	viewer := viewerFromCtx(c)
	profile, err := h.usecase.GetPublicProfile(c.Context(), viewer, username)
	if err != nil {
		h.log.Error().Err(err).Str("username", username).Msg("get public profile failed")
		return handleErr(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(profile)
}

// @summary      Update my profile
// @tags         users
// @accept       json
// @produce      json
// @security     CookieAuth
// @param        request  body      usecase.UpdateProfileInput  true  "Fields to update (all optional)"
// @success      200      {object}  dtos.PrivateProfileData
// @failure      400      {object}  ErrorResponse
// @failure      404      {object}  ErrorResponse
// @failure      409      {object}  ErrorResponse
// @failure      500      {object}  ErrorResponse
// @router       /v1/users/me [patch]
func (h *usersHandler) updateProfile(c fiber.Ctx) error {
	viewer := viewerFromCtx(c)
	req := &usecase.UpdateProfileInput{}
	if err := c.Bind().Body(req); err != nil {
		return handleErr(c, usecase.ErrInvalidData)
	}
	user, err := h.usecase.UpdateProfile(c.Context(), viewer, req)
	if err != nil {
		h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("update profile failed")
		return handleErr(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(user)
}

// @summary      Upload avatar
// @tags         users
// @accept       mpfd
// @produce      json
// @security     CookieAuth
// @param        avatar  formData  file  true  "Avatar image (JPEG/PNG/WebP)"
// @success      200     {object}  dtos.PrivateProfileData
// @failure      400     {object}  ErrorResponse
// @failure      404     {object}  ErrorResponse
// @failure      500     {object}  ErrorResponse
// @router       /v1/users/me/avatar [put]
func (h *usersHandler) updateAvatar(c fiber.Ctx) error {
	viewer := viewerFromCtx(c)
	file, err := c.FormFile("avatar")
	if err != nil {
		return handleErr(c, usecase.ErrInvalidData)
	}
	user, err := h.usecase.UpdateAvatar(c.Context(), viewer, file)
	if err != nil {
		h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("update avatar failed")
		return handleErr(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(user)
}

// @summary      Remove avatar
// @tags         users
// @produce      json
// @security     CookieAuth
// @success      200  {object}  dtos.PrivateProfileData
// @failure      404  {object}  ErrorResponse
// @failure      500  {object}  ErrorResponse
// @router       /v1/users/me/avatar [delete]
func (h *usersHandler) removeAvatar(c fiber.Ctx) error {
	viewer := viewerFromCtx(c)
	user, err := h.usecase.RemoveAvatar(c.Context(), viewer)
	if err != nil {
		h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("remove avatar failed")
		return handleErr(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(user)
}

// @summary      Upload banner
// @tags         users
// @accept       mpfd
// @produce      json
// @security     CookieAuth
// @param        banner  formData  file  true  "Banner image (JPEG/PNG/WebP)"
// @success      200     {object}  dtos.PrivateProfileData
// @failure      400     {object}  ErrorResponse
// @failure      404     {object}  ErrorResponse
// @failure      500     {object}  ErrorResponse
// @router       /v1/users/me/banner [put]
func (h *usersHandler) updateBanner(c fiber.Ctx) error {
	viewer := viewerFromCtx(c)
	file, err := c.FormFile("banner")
	if err != nil {
		return handleErr(c, usecase.ErrInvalidData)
	}
	user, err := h.usecase.UpdateBanner(c.Context(), viewer, file)
	if err != nil {
		h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("update banner failed")
		return handleErr(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(user)
}

// @summary      Remove banner
// @tags         users
// @produce      json
// @security     CookieAuth
// @success      200  {object}  dtos.PrivateProfileData
// @failure      404  {object}  ErrorResponse
// @failure      500  {object}  ErrorResponse
// @router       /v1/users/me/banner [delete]
func (h *usersHandler) removeBanner(c fiber.Ctx) error {
	viewer := viewerFromCtx(c)
	user, err := h.usecase.RemoveBanner(c.Context(), viewer)
	if err != nil {
		h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("remove banner failed")
		return handleErr(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(user)
}

// @summary      Follow user
// @tags         users
// @security     CookieAuth
// @param        username  path  string  true  "Target username"
// @success      204
// @failure      400  {object}  ErrorResponse
// @failure      404  {object}  ErrorResponse
// @failure      500  {object}  ErrorResponse
// @router       /v1/users/{username}/follow [post]
func (h *usersHandler) followUser(c fiber.Ctx) error {
	viewer := viewerFromCtx(c)
	targetUsername := c.Params("username")
	if err := h.usecase.Follow(c.Context(), viewer, targetUsername); err != nil {
		h.log.Error().Err(err).Str("followerID", viewer.UserID).Str("target", targetUsername).Msg("follow failed")
		return handleErr(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// @summary      Unfollow user
// @tags         users
// @security     CookieAuth
// @param        username  path  string  true  "Target username"
// @success      204
// @failure      404  {object}  ErrorResponse
// @failure      500  {object}  ErrorResponse
// @router       /v1/users/{username}/follow [delete]
func (h *usersHandler) unfollowUser(c fiber.Ctx) error {
	viewer := viewerFromCtx(c)
	targetUsername := c.Params("username")
	if err := h.usecase.Unfollow(c.Context(), viewer, targetUsername); err != nil {
		h.log.Error().Err(err).Str("followerID", viewer.UserID).Str("target", targetUsername).Msg("unfollow failed")
		return handleErr(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// @summary      Get user briefs by IDs
// @tags         internal
// @produce      json
// @param        ids  query     string          true  "Comma-separated user IDs"
// @success      200  {array}   dtos.UserBrief
// @failure      400  {object}  ErrorResponse
// @failure      500  {object}  ErrorResponse
// @router       /internal/briefs [get]
func (h *usersHandler) getBriefs(c fiber.Ctx) error {
	ids := strings.Split(c.Query("ids"), ",")
	if len(ids) == 0 {
		return handleErr(c, usecase.ErrInvalidData)
	}
	briefs, err := h.usecase.GetBriefs(c.Context(), ids)
	if err != nil {
		h.log.Error().Err(err).Strs("ids", ids).Msg("get briefs failed")
		return handleErr(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(briefs)
}

// @summary      Update user stats (internal)
// @tags         internal
// @accept       json
// @param        X-Internal-Token  header    string               true  "Internal service token"
// @param        request           body      UpdateStatsRequest   true  "Stats payload"
// @success      204
// @failure      400  {object}  ErrorResponse
// @failure      500  {object}  ErrorResponse
// @router       /internal/stats [patch]
func (h *usersHandler) updateStats(c fiber.Ctx) error {
	var req UpdateStatsRequest
	if err := c.Bind().Body(&req); err != nil {
		return handleErr(c, usecase.ErrInvalidData)
	}
	statName, err := parseStatName(req.StatName)
	if err != nil {
		return handleErr(c, usecase.ErrInvalidData)
	}
	if err := h.usecase.UpdateStats(c.Context(), statName, req.Stats); err != nil {
		h.log.Error().Err(err).Msg("update stats failed")
		return handleErr(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func parseStatName(s string) (dtos.UserStatName, error) {
	v := dtos.UserStatName(s)
	switch v {
	case dtos.PlayedStatName, dtos.HostedStatName:
		return v, nil
	default:
		return "", ErrBadReq
	}
}
